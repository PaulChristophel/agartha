package validate

import (
	"fmt"
	"strings"
)

func OrderBy(orderBy string, validColumns []string, jsonCol string, jsonPath []string) (string, error) {
	columns := strings.Split(orderBy, ",")
	var validatedColumns []string
	validColumnsMap := make(map[string]bool)
	jsonPathMap := make(map[string]bool)

	// Populate the valid columns map
	for _, col := range validColumns {
		validColumnsMap[col] = true
	}

	// Populate the json path map
	for _, path := range jsonPath {
		last := strings.Split(path, ".")
		jsonPathMap[last[len(last)-1]] = true
	}

	// Validate and build the order by columns
	for _, column := range columns {
		colParts := strings.Fields(column)
		if len(colParts) == 0 {
			continue
		}

		colName := colParts[0]
		sortDirection := "asc"

		if len(colParts) > 1 {
			if colParts[1] == "desc" || colParts[1] == "asc" {
				sortDirection = colParts[1]
			} else {
				return "", fmt.Errorf("invalid sort direction, must be 'asc' or 'desc'")
			}
		}

		// Check if colName is a valid column
		if validColumnsMap[colName] {
			validatedColumns = append(validatedColumns, colName+" "+sortDirection)
		} else if jsonPathMap[colName] || jsonPathMap["\""+colName+"\""] {
			// Handle JSON path columns
			validatedColumns = append(validatedColumns, fmt.Sprintf("jsonb_extract_path_text(%s, '%s') %s", jsonCol, strings.ReplaceAll(colName, ".", "', '"), sortDirection))
		} else {
			return "", fmt.Errorf("invalid column name '%s'. Valid columns: %v", colName, validColumns)
		}
	}

	return strings.Join(validatedColumns, ", "), nil
}

// OrderByJSONColumns validates SQL columns and selected JSON paths across multiple JSONB columns.
// JSON fields use a qualified order_by name such as "grains.os" or "pillar.role". Paths from
// defaultJSONColumn may also use their legacy unqualified leaf name for backward compatibility.
func OrderByJSONColumns(
	orderBy string,
	validColumns []string,
	jsonColumns map[string][]string,
	defaultJSONColumn string,
) (string, error) {
	columns := strings.Split(orderBy, ",")
	validColumnsMap := make(map[string]bool)
	jsonOrderExpressions := make(map[string]string)
	var validatedColumns []string

	for _, col := range validColumns {
		validColumnsMap[col] = true
	}

	for jsonColumn, paths := range jsonColumns {
		for _, path := range paths {
			if path == "" {
				continue
			}
			segments := strings.Split(path, ".")
			leaf := strings.Trim(segments[len(segments)-1], `"`)
			expression := fmt.Sprintf(
				"jsonb_extract_path_text(%s, '%s')",
				jsonColumn,
				strings.Join(segments, "', '"),
			)
			jsonOrderExpressions[jsonColumn+"."+leaf] = expression
			if jsonColumn == defaultJSONColumn {
				jsonOrderExpressions[leaf] = expression
			}
		}
	}

	for _, column := range columns {
		colParts := strings.Fields(column)
		if len(colParts) == 0 {
			continue
		}

		colName := colParts[0]
		sortDirection := "asc"
		if len(colParts) > 1 {
			if colParts[1] != "asc" && colParts[1] != "desc" {
				return "", fmt.Errorf("invalid sort direction, must be 'asc' or 'desc'")
			}
			sortDirection = colParts[1]
		}

		if validColumnsMap[colName] {
			validatedColumns = append(validatedColumns, colName+" "+sortDirection)
		} else if expression, ok := jsonOrderExpressions[colName]; ok {
			validatedColumns = append(validatedColumns, expression+" "+sortDirection)
		} else {
			return "", fmt.Errorf("invalid column name '%s'. Valid columns: %v", colName, validColumns)
		}
	}

	return strings.Join(validatedColumns, ", "), nil
}
