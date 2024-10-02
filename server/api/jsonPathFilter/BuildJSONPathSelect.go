package jsonPathFilter

import (
	"fmt"
	"regexp"
	"strings"
)

/*
BuildJSONPathSelect constructs the JSON path query based on the provided paths
and a specified column. It generates a SQL query string that selects specific
fields from a JSONB column in a PostgreSQL database.

Parameters:
- jsonPaths: A slice of strings representing the JSON paths to be queried.
- column: The name of the column from which the JSON data is being queried.

Returns:
  - A string representing the SQL query that builds a JSON object from the
    specified JSON paths in the given column.

Functionality:
  - If only one JSON path is provided, it constructs a JSON path query for that
    single path and creates a JSON object with one key-value pair.
  - If multiple JSON paths are provided, it constructs JSON path queries for each
    path and creates a JSON object with multiple key-value pairs.

Limitations:
  - The function assumes that keys in the JSON paths do not contain periods. If
    keys contain periods, the generated JSON path query might not work correctly,
    because the function does not escape or handle periods within keys.
  - The function does not handle cases where keys contain special characters
    other than periods.
  - The function does not handle array indices in JSON paths correctly.

Usage Example:

  - Single Path:
    Input: jsonPaths = ["grains.os"], column = "data"
    Output: jsonb_build_object('os', jsonb_path_query(data, '$.grains.os')) AS data

  - Multiple Paths:
    Input: jsonPaths = ["grains.os", "grains.id"], column = "data"
    Output: jsonb_build_object('os', jsonb_path_query(data, '$.grains.os'), 'id', jsonb_path_query(data, '$.grains.id')) AS data
*/

func BuildJSONPathSelect(jsonPaths []string, column string) string {
	if len(jsonPaths) == 1 {
		key, jsonPath := extractJSONPathDetails(jsonPaths[0])
		return fmt.Sprintf("jsonb_build_object('%s', jsonb_path_query(%s, '%s')) AS %s", key, column, jsonPath, column)
	}

	queryParts := make([]string, len(jsonPaths)*2)
	for i, path := range jsonPaths {
		key, jsonPath := extractJSONPathDetails(path)
		queryParts[i*2] = fmt.Sprintf("'%s'", key)
		queryParts[i*2+1] = fmt.Sprintf("jsonb_path_query(%s, '%s')", column, jsonPath)
	}

	return fmt.Sprintf("jsonb_build_object(%s) AS %s", strings.Join(queryParts, ", "), column)
}

func extractJSONPathDetails(path string) (string, string) {
	// Regular expression to match keys, including those with special characters and array indices
	re := regexp.MustCompile(`(?:^|\.)("([^"]*)"|([^.\[\]]+)|(\[\d+\]))`)
	matches := re.FindAllStringSubmatch(path, -1)
	var keys []string

	for _, match := range matches {
		if match[2] != "" {
			keys = append(keys, match[2])
		} else if match[3] != "" {
			keys = append(keys, match[3])
		} else if match[4] != "" {
			keys = append(keys, match[4])
		}
	}

	key := keys[len(keys)-1]
	// Quote keys that need to be quoted in JSON path
	for i, k := range keys {
		if strings.ContainsAny(k, `-/:{}()`) || strings.HasPrefix(k, "[") {
			keys[i] = fmt.Sprintf(`"%s"`, k)
		}
	}

	jsonPath := fmt.Sprintf("$.%s", strings.Join(keys, "."))
	// Remove quotes from array indices
	jsonPath = strings.ReplaceAll(jsonPath, `"[`, `[`)
	jsonPath = strings.ReplaceAll(jsonPath, `]"`, `]`)
	return key, jsonPath
}
