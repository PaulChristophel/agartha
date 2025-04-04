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
