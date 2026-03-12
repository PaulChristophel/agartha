package jsonPathFilter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

/*
BuildJSONPathWhere constructs the JSONB containment filter for multiple JSON path filters.
It generates a SQL query string that applies the JSONB containment operator (@>) to filter
rows based on the specified JSON path filters.

Parameters:
  - jsonPathFilters: A slice of strings representing the JSON path filters. Each filter
    should be in the format "key:value::type" where:
  - "key" is the JSON path key (can contain periods and should be properly quoted).
  - "value" is the value to match.
  - "type" is the type of the value (int, float, bool, array, or default string).
  - column: The name of the column from which the JSON data is being filtered.

Returns:
  - A string representing the SQL query that applies the JSONB containment filter on the
    specified column, or an error if the input format is invalid or if a value cannot be parsed.

Functionality:
- Parses each filter and splits it into key, value, and type components.
- Handles keys with periods correctly by splitting and nesting the JSON structure appropriately.
- Parses the value according to its specified type (int, float, bool, array, or string).
- Constructs a combined JSON object from all filters and generates the corresponding SQL query.

Limitations:
  - The function assumes that keys in the JSON paths do not contain characters other than
    periods that require special handling (e.g., special characters in JSON keys).
  - The function does not handle cases where the format of the filter is incorrect (e.g., missing
    type or improperly quoted keys) and will return an error in such cases.
  - The function expects the value part of the filter to be properly quoted if it contains spaces
    or special characters.

Usage Example:

  - Single Filter:
    Input: jsonPathFilters = ["grains.id:test::string"], column = "data"
    Output: data @> '{"grains":{"id":"test"}}'::jsonb

  - Multiple Filters:
    Input: jsonPathFilters = ["grains.id:test::string", "grains.count:5::int"], column = "data"
    Output: data @> '{"grains":{"id":"test","count":5}}'::jsonb
*/
func BuildJSONPathWhere(jsonPathFilters []string, column string) (string, error) {
	if len(jsonPathFilters) == 0 {
		return "", fmt.Errorf("no filters provided")
	}

	var expressions []string

	for _, filter := range jsonPathFilters {
		keyPart, value, typ, operator, err := parseFilterParts(filter)
		if err != nil {
			return "", err
		}

		keys, err := splitKeys(keyPart)
		if err != nil {
			return "", err
		}

		parsedValue, err := parseTypedValue(value, typ)
		if err != nil {
			return "", err
		}

		expr, err := buildExpression(column, keys, value, parsedValue, operator)
		if err != nil {
			return "", err
		}

		expressions = append(expressions, expr)
	}

	return strings.Join(expressions, " AND "), nil
}

func parseFilterParts(filter string) (keyPart string, value string, typ string, operator string, err error) {
	parts := strings.Split(filter, "::")
	if len(parts) < 2 {
		return "", "", "", "", fmt.Errorf("invalid filter format")
	}

	pathAndValue := parts[0]
	lastColonIndex := strings.LastIndex(pathAndValue, ":")
	if lastColonIndex == -1 {
		return "", "", "", "", fmt.Errorf("invalid key/value format")
	}

	keyPart = pathAndValue[:lastColonIndex]
	value = strings.Trim(pathAndValue[lastColonIndex+1:], `"`)
	typ = parts[1]
	operator = "eq"
	if len(parts) > 2 && parts[2] != "" {
		operator = strings.ToLower(parts[2])
	}

	return keyPart, value, typ, operator, nil
}

func splitKeys(keyPart string) ([]string, error) {
	if keyPart == "" {
		return nil, fmt.Errorf("empty key part")
	}

	if strings.Contains(keyPart, `"`) {
		keyPart = strings.Trim(keyPart, `"`)
		keys := strings.Split(keyPart, `"."`)
		if len(keys) > 0 {
			keys[len(keys)-1] = strings.TrimSuffix(keys[len(keys)-1], `"`)
		}
		return keys, nil
	}

	return strings.Split(keyPart, "."), nil
}

func parseTypedValue(value string, typ string) (any, error) {
	var (
		parsedValue any
		parseErr    error
	)

	switch typ {
	case "int":
		parsedValue, parseErr = strconv.Atoi(value)
	case "float":
		parsedValue, parseErr = strconv.ParseFloat(value, 64)
	case "bool", "boolean":
		parsedValue, parseErr = strconv.ParseBool(value)
	case "array":
		parsedValue, parseErr = parseArray(value)
	case "null":
		parsedValue = nil
	default:
		parsedValue = value
	}

	if parseErr != nil {
		return nil, fmt.Errorf("error parsing value '%s' as %s: %v", value, typ, parseErr)
	}

	return parsedValue, nil
}

func buildExpression(column string, keys []string, rawValue string, parsedValue any, operator string) (string, error) {
	switch operator {
	case "eq":
		return buildContainmentExpr(column, keys, parsedValue)
	case "not", "neq":
		eqExpr, err := buildContainmentExpr(column, keys, parsedValue)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("NOT (%s)", eqExpr), nil
	case "like", "not_like":
		return buildLikeExpr(column, keys, rawValue, operator == "not_like"), nil
	default:
		return "", fmt.Errorf("unsupported operator: %s", operator)
	}
}

func buildContainmentExpr(column string, keys []string, value any) (string, error) {
	filterMap := make(map[string]any)
	currentMap := filterMap
	for i, key := range keys {
		if i == len(keys)-1 {
			currentMap[key] = value
			continue
		}

		if _, exists := currentMap[key]; !exists {
			currentMap[key] = make(map[string]any)
		}
		var ok bool
		currentMap, ok = currentMap[key].(map[string]any)
		if !ok {
			return "", fmt.Errorf("invalid nested key structure for %s", key)
		}
	}

	filterJSON, err := json.Marshal(filterMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s @> '%s'::jsonb", column, string(filterJSON)), nil
}

func buildLikeExpr(column string, keys []string, rawValue string, negate bool) string {
	pathExpression := buildJSONTextAccessor(column, keys)
	escapedValue := strings.ReplaceAll(rawValue, "'", "''")
	comparator := "LIKE"
	if negate {
		comparator = "NOT LIKE"
	}
	return fmt.Sprintf("(%s) %s '%s'", pathExpression, comparator, escapedValue)
}

func buildJSONTextAccessor(column string, keys []string) string {
	pathParts := make([]string, len(keys))
	for i, key := range keys {
		key = strings.ReplaceAll(key, `"`, `\"`)
		pathParts[i] = fmt.Sprintf(`"%s"`, key)
	}
	return fmt.Sprintf("%s #>> '{%s}'", column, strings.Join(pathParts, ","))
}

// parseArray parses a string representation of an array into an actual array.
func parseArray(value string) ([]any, error) {
	trimmedValue := strings.Trim(value, "[]")
	if trimmedValue == "" {
		return []any{}, nil
	}

	items := strings.Split(trimmedValue, ",")
	var array []any
	for _, item := range items {
		item = strings.TrimSpace(item)
		if intValue, err := strconv.Atoi(item); err == nil {
			array = append(array, intValue)
		} else if floatValue, err := strconv.ParseFloat(item, 64); err == nil {
			array = append(array, floatValue)
		} else if boolValue, err := strconv.ParseBool(item); err == nil {
			array = append(array, boolValue)
		} else {
			array = append(array, strings.Trim(item, `"`))
		}
	}
	return array, nil
}
