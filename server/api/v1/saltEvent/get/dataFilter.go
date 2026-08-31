package saltEvent

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
	dataMatchKeyExists      = "key_exists"
	dataMatchStringEquals   = "string_equals"
	dataMatchKeyFieldEquals = "key_field_equals"
	dataQueryMaxClauses     = 20
	dataQueryMaxPathDepth   = 16
)

type dataQuery struct {
	Logic   string            `json:"logic"`
	Clauses []dataQueryClause `json:"clauses"`
}

type dataQueryClause struct {
	Scope         string   `json:"scope"`
	Key           string   `json:"key,omitempty"`
	ContainerPath []string `json:"container_path,omitempty"`
	Path          []string `json:"path,omitempty"`
	Operator      string   `json:"operator"`
	Value         string   `json:"value,omitempty"`
	ValueType     string   `json:"value_type,omitempty"`
}

type dataFilterOptions struct {
	Match        string
	Filter       string
	Key          string
	Field        string
	Value        string
	ValueType    string
	HasMatch     bool
	HasFilter    bool
	HasKey       bool
	HasField     bool
	HasValue     bool
	HasValueType bool
}

func applyDataFilter(query *gorm.DB, options dataFilterOptions, jsonbEnabled bool) (*gorm.DB, error) {
	if !options.HasMatch && !options.hasFilterParameters() {
		return query, nil
	}
	if !options.HasMatch {
		return query, errors.New("'data_match' is required when data filter parameters are provided")
	}
	if !jsonbEnabled {
		return query, errors.New("data filtering requires JSONB database tables")
	}

	switch strings.ToLower(strings.TrimSpace(options.Match)) {
	case dataMatchKeyExists:
		if !options.HasFilter {
			return query, errors.New("'data_filter' is required for key_exists")
		}
		if options.Filter == "" {
			return query, errors.New("'data_filter' cannot be empty for key_exists")
		}
		return query.Where(
			`jsonb_typeof("data") = 'object' AND jsonb_exists("data", ?)`,
			options.Filter,
		), nil
	case dataMatchStringEquals:
		if !options.HasFilter {
			return query, errors.New("'data_filter' is required for string_equals")
		}
		return query.Where(
			`jsonb_typeof("data") = 'string' AND "data" = to_jsonb(CAST(? AS text))`,
			options.Filter,
		), nil
	case dataMatchKeyFieldEquals:
		return applyKeyFieldEqualsFilter(query, options)
	default:
		return query, errors.New("invalid 'data_match' value: must be key_exists, string_equals, or key_field_equals")
	}
}

func applyAdvancedDataFilter(query *gorm.DB, rawQuery string, hasQuery bool, jsonbEnabled bool) (*gorm.DB, error) {
	if !hasQuery {
		return query, nil
	}
	if !jsonbEnabled {
		return query, errors.New("data filtering requires JSONB database tables")
	}
	if strings.TrimSpace(rawQuery) == "" {
		return query, errors.New("'data_query' cannot be empty")
	}

	var filter dataQuery
	decoder := json.NewDecoder(strings.NewReader(rawQuery))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&filter); err != nil {
		return query, errors.New("invalid 'data_query' JSON")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return query, errors.New("invalid 'data_query' JSON")
	}
	if len(filter.Clauses) == 0 {
		return query, errors.New("'data_query' must contain at least one clause")
	}
	if len(filter.Clauses) > dataQueryMaxClauses {
		return query, fmt.Errorf("'data_query' supports at most %d clauses", dataQueryMaxClauses)
	}

	logic := strings.ToLower(strings.TrimSpace(filter.Logic))
	if logic == "" {
		logic = "and"
	}
	if logic != "and" && logic != "or" {
		return query, errors.New("invalid data query logic: must be 'and' or 'or'")
	}

	predicates := make([]string, 0, len(filter.Clauses))
	args := make([]any, 0, len(filter.Clauses)*3)
	for index, clause := range filter.Clauses {
		predicate, clauseArgs, err := buildDataQueryClause(clause)
		if err != nil {
			return query, fmt.Errorf("invalid data query clause %d: %w", index+1, err)
		}
		predicates = append(predicates, "("+predicate+")")
		args = append(args, clauseArgs...)
	}

	return query.Where("("+strings.Join(predicates, " "+strings.ToUpper(logic)+" ")+")", args...), nil
}

func buildDataQueryClause(clause dataQueryClause) (string, []any, error) {
	scope := strings.ToLower(strings.TrimSpace(clause.Scope))
	if scope == "" {
		scope = "key"
	}
	if len(clause.Path) > dataQueryMaxPathDepth {
		return "", nil, fmt.Errorf("path supports at most %d components", dataQueryMaxPathDepth)
	}
	if slices.Contains(clause.Path, "") {
		return "", nil, errors.New("path components cannot be empty")
	}
	if len(clause.ContainerPath) > dataQueryMaxPathDepth {
		return "", nil, fmt.Errorf("container path supports at most %d components", dataQueryMaxPathDepth)
	}
	if slices.Contains(clause.ContainerPath, "") {
		return "", nil, errors.New("container path components cannot be empty")
	}
	if len(clause.ContainerPath) > 0 && scope != "any_key" && scope != "any_state" {
		return "", nil, errors.New("container path is only supported for any_key scope")
	}

	var target string
	var args []any
	switch scope {
	case "root":
		target, args = jsonbExtractExpression(`"data"`, clause.Path)
	case "key", "state":
		if clause.Key == "" {
			return "", nil, errors.New("key is required for key scope")
		}
		path := append([]string{clause.Key}, clause.Path...)
		target, args = jsonbExtractExpression(`"data"`, path)
	case "any_key", "any_state":
		containerTarget, containerArgs := jsonbExtractExpression(`"data"`, clause.ContainerPath)
		target, args = jsonbExtractExpression("data_entry.value", clause.Path)
		args = append(containerArgs, args...)
		predicate, comparisonArgs, err := buildDataComparison(target, clause)
		if err != nil {
			return "", nil, err
		}
		args = append(args, comparisonArgs...)
		predicate = `EXISTS (SELECT 1 FROM (SELECT ` + containerTarget + ` AS value) AS data_container ` +
			`CROSS JOIN LATERAL jsonb_each(CASE WHEN jsonb_typeof(data_container.value) = 'object' ` +
			`THEN data_container.value ELSE '{}'::jsonb END) AS data_entry(key, value) WHERE ` + predicate + `)`
		return predicate, args, nil
	default:
		return "", nil, errors.New("scope must be root, key, or any_key")
	}

	predicate, comparisonArgs, err := buildDataComparison(target, clause)
	if err != nil {
		return "", nil, err
	}
	args = append(args, comparisonArgs...)
	return predicate, args, nil
}

func jsonbExtractExpression(base string, path []string) (string, []any) {
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

func buildDataComparison(target string, clause dataQueryClause) (string, []any, error) {
	operator := strings.ToLower(strings.TrimSpace(clause.Operator))
	const valueExpression = "data_target.value"
	wrap := func(predicate string) string {
		return "EXISTS (SELECT 1 FROM (SELECT " + target + " AS value) AS data_target WHERE " + predicate + ")"
	}
	switch operator {
	case "exists":
		return wrap(valueExpression + " IS NOT NULL"), nil, nil
	case "not_exists":
		return wrap(valueExpression + " IS NULL"), nil, nil
	}

	value, err := parseDataFilterValue(clause.Value, clause.ValueType)
	if err != nil {
		return "", nil, err
	}
	switch operator {
	case "eq", "ne":
		encoded, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			return "", nil, fmt.Errorf("encode comparison value: %w", marshalErr)
		}
		comparison := "="
		if operator == "ne" {
			comparison = "<>"
		}
		return wrap(valueExpression + " " + comparison + " CAST(? AS jsonb)"), []any{string(encoded)}, nil
	case "gt", "gte", "lt", "lte":
		if !isNumericDataType(clause.ValueType) {
			return "", nil, errors.New("numeric operators require an int or float value type")
		}
		comparison := map[string]string{"gt": ">", "gte": ">=", "lt": "<", "lte": "<="}[operator]
		return wrap(`jsonb_typeof(` + valueExpression + `) = 'number' AND (` + valueExpression + ` #>> '{}')::numeric ` + comparison + ` CAST(? AS numeric)`), []any{clause.Value}, nil
	case "contains", "icontains", "regex":
		if strings.ToLower(strings.TrimSpace(clause.ValueType)) != "string" {
			return "", nil, errors.New("string operators require a string value type")
		}
		textTarget := "(" + valueExpression + ` #>> '{}')`
		switch operator {
		case "contains":
			return wrap(`jsonb_typeof(` + valueExpression + `) = 'string' AND strpos(` + textTarget + `, ?) > 0`), []any{clause.Value}, nil
		case "icontains":
			return wrap(`jsonb_typeof(` + valueExpression + `) = 'string' AND strpos(lower(` + textTarget + `), lower(?)) > 0`), []any{clause.Value}, nil
		default:
			if _, compileErr := regexp.Compile(clause.Value); compileErr != nil {
				return "", nil, errors.New("invalid regular expression value")
			}
			return wrap(`jsonb_typeof(` + valueExpression + `) = 'string' AND ` + textTarget + ` ~ ?`), []any{clause.Value}, nil
		}
	default:
		return "", nil, errors.New("operator must be eq, ne, gt, gte, lt, lte, contains, icontains, regex, exists, or not_exists")
	}
}

func isNumericDataType(valueType string) bool {
	switch strings.ToLower(strings.TrimSpace(valueType)) {
	case "int", "integer", "float", "number":
		return true
	default:
		return false
	}
}

func (options dataFilterOptions) hasFilterParameters() bool {
	return options.HasFilter || options.HasKey || options.HasField || options.HasValue || options.HasValueType
}

func applyKeyFieldEqualsFilter(query *gorm.DB, options dataFilterOptions) (*gorm.DB, error) {
	if !options.HasKey || options.Key == "" {
		return query, errors.New("'data_key' is required for key_field_equals")
	}
	if !options.HasField || options.Field == "" {
		return query, errors.New("'data_field' is required for key_field_equals")
	}
	if !options.HasValue {
		return query, errors.New("'data_value' is required for key_field_equals")
	}
	if !options.HasValueType {
		return query, errors.New("'data_value_type' is required for key_field_equals")
	}

	value, err := parseDataFilterValue(options.Value, options.ValueType)
	if err != nil {
		return query, err
	}
	filterJSON, err := json.Marshal(map[string]any{
		options.Key: map[string]any{options.Field: value},
	})
	if err != nil {
		return query, fmt.Errorf("build key field filter: %w", err)
	}

	return query.Where(
		`jsonb_typeof("data") = 'object' AND "data" @> CAST(? AS jsonb)`,
		string(filterJSON),
	), nil
}

func parseDataFilterValue(value string, valueType string) (any, error) {
	switch strings.ToLower(strings.TrimSpace(valueType)) {
	case "string":
		return value, nil
	case "bool", "boolean":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, errors.New("invalid boolean 'data_value'")
		}
		return parsed, nil
	case "int", "integer":
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, errors.New("invalid integer 'data_value'")
		}
		return parsed, nil
	case "float", "number":
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, errors.New("invalid numeric 'data_value'")
		}
		return parsed, nil
	case "null":
		return nil, nil
	default:
		return nil, errors.New("invalid 'data_value_type': must be string, bool, int, float, or null")
	}
}
