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
	// Initialize an empty map to build the combined JSON object
	filterMap := make(map[string]interface{})

	for _, filter := range jsonPathFilters {
		filters := strings.Split(filter, "::")
		if len(filters) != 2 {
			return "", fmt.Errorf("invalid filter format")
		}

		// Find the position of the last ':' and split the string accordingly
		lastColonIndex := strings.LastIndex(filters[0], ":")
		if lastColonIndex == -1 {
			return "", fmt.Errorf("invalid key/value format")
		}

		keyPart := filters[0][:lastColonIndex]
		value := filters[0][lastColonIndex+1:]

		// Handle keys with periods correctly
		var keys []string
		if strings.Contains(keyPart, `"`) {
			// Remove the leading and trailing quotes if present
			keyPart = strings.Trim(keyPart, `"`)
			// Split the keys by period
			keys = strings.Split(keyPart, `"."`)
			// Ensure the last key part is not mistakenly trimmed
			if len(keys) > 0 {
				keys[len(keys)-1] = strings.TrimSuffix(keys[len(keys)-1], `"`)
			}
		} else {
			keys = strings.Split(keyPart, ".")
		}

		// Strip quotes from value
		value = strings.Trim(value, `"`)

		typ := filters[1]

		// Parse the value according to its type
		var parsedValue interface{}
		var parseErr error
		switch typ {
		case "int":
			parsedValue, parseErr = strconv.Atoi(value)
		case "float":
			parsedValue, parseErr = strconv.ParseFloat(value, 64)
		case "bool", "boolean":
			parsedValue, parseErr = strconv.ParseBool(value)
		case "array":
			parsedValue, parseErr = parseArray(value)
		default:
			parsedValue = value
		}

		if parseErr != nil {
			return "", fmt.Errorf("error parsing value '%s' as %s: %v", value, typ, parseErr)
		}

		// Construct the JSON object for the filter
		currentMap := filterMap
		for i, key := range keys {
			if i == len(keys)-1 {
				currentMap[key] = parsedValue
			} else {
				if _, exists := currentMap[key]; !exists {
					currentMap[key] = make(map[string]interface{})
				}
				currentMap = currentMap[key].(map[string]interface{})
			}
		}
	}

	filterJSON, err := json.Marshal(filterMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s @> '%s'::jsonb", column, string(filterJSON)), nil
}

// parseArray parses a string representation of an array into an actual array.
func parseArray(value string) ([]interface{}, error) {
	trimmedValue := strings.Trim(value, "[]")
	if trimmedValue == "" {
		return []interface{}{}, nil
	}

	items := strings.Split(trimmedValue, ",")
	var array []interface{}
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
