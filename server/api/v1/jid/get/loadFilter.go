package jid

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

const (
	loadQueryMaxClauses   = 20
	loadQueryMaxPathDepth = 16
)

type loadQuery struct {
	Logic   string            `json:"logic"`
	Clauses []loadQueryClause `json:"clauses"`
}

type loadQueryClause struct {
	Scope         string   `json:"scope"`
	Key           string   `json:"key,omitempty"`
	ContainerPath []string `json:"container_path,omitempty"`
	Path          []string `json:"path,omitempty"`
	Operator      string   `json:"operator"`
	Value         string   `json:"value,omitempty"`
	ValueType     string   `json:"value_type,omitempty"`
}

func applyLoadFilter(query *gorm.DB, rawQuery string, hasQuery bool, jsonbEnabled bool) (*gorm.DB, error) {
	if !hasQuery {
		return query, nil
	}
	if !jsonbEnabled {
		return query, errors.New("load filtering requires JSONB database tables")
	}
	if strings.TrimSpace(rawQuery) == "" {
		return query, errors.New("'load_query' cannot be empty")
	}

	var filter loadQuery
	decoder := json.NewDecoder(strings.NewReader(rawQuery))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&filter); err != nil {
		return query, errors.New("invalid 'load_query' JSON")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return query, errors.New("invalid 'load_query' JSON")
	}
	if len(filter.Clauses) == 0 {
		return query, errors.New("'load_query' must contain at least one clause")
	}
	if len(filter.Clauses) > loadQueryMaxClauses {
		return query, fmt.Errorf("'load_query' supports at most %d clauses", loadQueryMaxClauses)
	}

	logic := strings.ToLower(strings.TrimSpace(filter.Logic))
	if logic == "" {
		logic = "and"
	}
	if logic != "and" && logic != "or" {
		return query, errors.New("invalid load query logic: must be 'and' or 'or'")
	}

	predicates := make([]string, 0, len(filter.Clauses))
	args := make([]any, 0, len(filter.Clauses)*3)
	for index, clause := range filter.Clauses {
		predicate, clauseArgs, err := buildLoadQueryClause(clause)
		if err != nil {
			return query, fmt.Errorf("invalid load query clause %d: %w", index+1, err)
		}
		predicates = append(predicates, "("+predicate+")")
		args = append(args, clauseArgs...)
	}

	return query.Where("("+strings.Join(predicates, " "+strings.ToUpper(logic)+" ")+")", args...), nil
}

func buildLoadQueryClause(clause loadQueryClause) (string, []any, error) {
	scope := strings.ToLower(strings.TrimSpace(clause.Scope))
	if scope == "" {
		scope = "key"
	}
	if len(clause.Path) > loadQueryMaxPathDepth || len(clause.ContainerPath) > loadQueryMaxPathDepth {
		return "", nil, fmt.Errorf("paths support at most %d components", loadQueryMaxPathDepth)
	}
	if slices.Contains(clause.Path, "") || slices.Contains(clause.ContainerPath, "") {
		return "", nil, errors.New("path components cannot be empty")
	}
	if len(clause.ContainerPath) > 0 && scope != "any_key" && scope != "any_state" {
		return "", nil, errors.New("container path is only supported for any_key scope")
	}

	var target string
	var args []any
	switch scope {
	case "root":
		target, args = loadJSONBExtractExpression(`"load"`, clause.Path)
	case "key", "state":
		if clause.Key == "" {
			return "", nil, errors.New("key is required for key scope")
		}
		target, args = loadJSONBExtractExpression(`"load"`, append([]string{clause.Key}, clause.Path...))
	case "any_key", "any_state":
		containerTarget, containerArgs := loadJSONBExtractExpression(`"load"`, clause.ContainerPath)
		target, args = loadJSONBExtractExpression("load_entry.value", clause.Path)
		predicate, comparisonArgs, err := buildLoadComparison(target, clause)
		if err != nil {
			return "", nil, err
		}
		args = append(containerArgs, args...)
		args = append(args, comparisonArgs...)
		return `EXISTS (SELECT 1 FROM (SELECT ` + containerTarget + ` AS value) AS load_container ` +
			`CROSS JOIN LATERAL jsonb_each(CASE WHEN jsonb_typeof(load_container.value) = 'object' ` +
			`THEN load_container.value ELSE '{}'::jsonb END) AS load_entry(key, value) WHERE ` + predicate + `)`, args, nil
	default:
		return "", nil, errors.New("scope must be root, key, or any_key")
	}

	predicate, comparisonArgs, err := buildLoadComparison(target, clause)
	if err != nil {
		return "", nil, err
	}
	return predicate, append(args, comparisonArgs...), nil
}

func loadJSONBExtractExpression(base string, path []string) (string, []any) {
	if len(path) == 0 {
		return base, nil
	}
	placeholders := make([]string, len(path))
	args := make([]any, len(path))
	for index, component := range path {
		placeholders[index] = "?"
		args[index] = component
	}
	return "jsonb_extract_path(" + base + ", " + strings.Join(placeholders, ", ") + ")", args
}

func buildLoadComparison(target string, clause loadQueryClause) (string, []any, error) {
	operator := strings.ToLower(strings.TrimSpace(clause.Operator))
	const valueExpression = "load_target.value"
	wrap := func(predicate string) string {
		return "EXISTS (SELECT 1 FROM (SELECT " + target + " AS value) AS load_target WHERE " + predicate + ")"
	}
	if operator == "exists" {
		return wrap(valueExpression + " IS NOT NULL"), nil, nil
	}
	if operator == "not_exists" {
		return wrap(valueExpression + " IS NULL"), nil, nil
	}

	value, err := parseLoadFilterValue(clause.Value, clause.ValueType)
	if err != nil {
		return "", nil, err
	}
	switch operator {
	case "eq", "ne":
		encoded, err := json.Marshal(value)
		if err != nil {
			return "", nil, fmt.Errorf("encode comparison value: %w", err)
		}
		comparison := "="
		if operator == "ne" {
			comparison = "<>"
		}
		return wrap(valueExpression + " " + comparison + " CAST(? AS jsonb)"), []any{string(encoded)}, nil
	case "gt", "gte", "lt", "lte":
		if !isNumericLoadType(clause.ValueType) {
			return "", nil, errors.New("numeric operators require an int or float value type")
		}
		comparison := map[string]string{"gt": ">", "gte": ">=", "lt": "<", "lte": "<="}[operator]
		return wrap(`jsonb_typeof(` + valueExpression + `) = 'number' AND (` + valueExpression + ` #>> '{}')::numeric ` + comparison + ` CAST(? AS numeric)`), []any{clause.Value}, nil
	case "contains", "icontains", "regex":
		if strings.ToLower(strings.TrimSpace(clause.ValueType)) != "string" {
			return "", nil, errors.New("string operators require a string value type")
		}
		textTarget := "(" + valueExpression + ` #>> '{}')`
		if operator == "contains" {
			return wrap(`jsonb_typeof(` + valueExpression + `) = 'string' AND strpos(` + textTarget + `, ?) > 0`), []any{clause.Value}, nil
		}
		if operator == "icontains" {
			return wrap(`jsonb_typeof(` + valueExpression + `) = 'string' AND strpos(lower(` + textTarget + `), lower(?)) > 0`), []any{clause.Value}, nil
		}
		if _, err := regexp.Compile(clause.Value); err != nil {
			return "", nil, errors.New("invalid regular expression value")
		}
		return wrap(`jsonb_typeof(` + valueExpression + `) = 'string' AND ` + textTarget + ` ~ ?`), []any{clause.Value}, nil
	default:
		return "", nil, errors.New("operator must be eq, ne, gt, gte, lt, lte, contains, icontains, regex, exists, or not_exists")
	}
}

func parseLoadFilterValue(value string, valueType string) (any, error) {
	switch strings.ToLower(strings.TrimSpace(valueType)) {
	case "string":
		return value, nil
	case "bool", "boolean":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, errors.New("invalid boolean 'load_value'")
		}
		return parsed, nil
	case "int", "integer":
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, errors.New("invalid integer 'load_value'")
		}
		return parsed, nil
	case "float", "number":
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, errors.New("invalid numeric 'load_value'")
		}
		return parsed, nil
	case "null":
		return nil, nil
	default:
		return nil, errors.New("invalid 'load_value_type': must be string, bool, int, float, or null")
	}
}

func isNumericLoadType(valueType string) bool {
	switch strings.ToLower(strings.TrimSpace(valueType)) {
	case "int", "integer", "float", "number":
		return true
	default:
		return false
	}
}
