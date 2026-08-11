package validate

import (
	"testing"
)

func TestOrderBy(t *testing.T) {
	tests := []struct {
		name         string
		orderBy      string
		validColumns []string
		jsonCol      string
		jsonPath     []string
		want         string
		wantErr      bool
		errMessage   string
	}{
		{
			name:         "Valid single column ascending",
			orderBy:      "name",
			validColumns: []string{"name", "age"},
			jsonCol:      "data",
			jsonPath:     []string{},
			want:         "name asc",
			wantErr:      false,
		},
		{
			name:         "Valid single column descending",
			orderBy:      "age desc",
			validColumns: []string{"name", "age"},
			jsonCol:      "data",
			jsonPath:     []string{},
			want:         "age desc",
			wantErr:      false,
		},
		{
			name:         "Valid multiple columns",
			orderBy:      "name, age desc",
			validColumns: []string{"name", "age"},
			jsonCol:      "data",
			jsonPath:     []string{},
			want:         "name asc, age desc",
			wantErr:      false,
		},
		{
			name:         "Invalid column name",
			orderBy:      "height",
			validColumns: []string{"name", "age"},
			jsonCol:      "data",
			jsonPath:     []string{},
			wantErr:      true,
			errMessage:   "invalid column name 'height'. Valid columns: [name age]",
		},
		{
			name:         "Invalid sort direction",
			orderBy:      "name ascending",
			validColumns: []string{"name", "age"},
			jsonCol:      "data",
			jsonPath:     []string{},
			wantErr:      true,
			errMessage:   "invalid sort direction, must be 'asc' or 'desc'",
		},
		{
			name:         "Empty orderBy string",
			orderBy:      "",
			validColumns: []string{"name", "age"},
			jsonCol:      "data",
			jsonPath:     []string{},
			want:         "",
			wantErr:      false,
		},
		{
			name:         "Valid JSON path column",
			orderBy:      "os desc",
			validColumns: []string{"name", "age"},
			jsonCol:      "data",
			jsonPath:     []string{"os", "osdist"},
			want:         "jsonb_extract_path_text(data, 'os') desc",
			wantErr:      false,
		},
		{
			name:         "Valid multiple columns including JSON path",
			orderBy:      "name, osdist desc",
			validColumns: []string{"name", "age"},
			jsonCol:      "data",
			jsonPath:     []string{"os", "osdist"},
			want:         "name asc, jsonb_extract_path_text(data, 'osdist') desc",
			wantErr:      false,
		},
		{
			name:         "Invalid column name in JSON path",
			orderBy:      "invalidpath",
			validColumns: []string{"name", "age"},
			jsonCol:      "data",
			jsonPath:     []string{"os", "osdist"},
			wantErr:      true,
			errMessage:   "invalid column name 'invalidpath'. Valid columns: [name age]",
		},
		{
			name:         "Valid quoted JSON path",
			orderBy:      `domain desc`,
			validColumns: []string{"name", "age"},
			jsonCol:      "data",
			jsonPath:     []string{"dns.domain"},
			want:         "jsonb_extract_path_text(data, 'domain') desc",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OrderBy(tt.orderBy, tt.validColumns, tt.jsonCol, tt.jsonPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderBy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMessage {
				t.Errorf("OrderBy() error = %v, wantErrMessage %v", err.Error(), tt.errMessage)
			}
			if got != tt.want {
				t.Errorf("OrderBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrderByJSONColumns(t *testing.T) {
	tests := []struct {
		name    string
		orderBy string
		want    string
		wantErr bool
	}{
		{
			name:    "qualified grain field",
			orderBy: "grains.os desc",
			want:    "jsonb_extract_path_text(grains, 'os') desc",
		},
		{
			name:    "qualified nested pillar field",
			orderBy: "pillar.application asc",
			want:    "jsonb_extract_path_text(pillar, 'roles', 'application') asc",
		},
		{
			name:    "legacy unqualified grain field",
			orderBy: "os",
			want:    "jsonb_extract_path_text(grains, 'os') asc",
		},
		{
			name:    "same leaf is disambiguated by qualifier",
			orderBy: "grains.role, pillar.role desc",
			want:    "jsonb_extract_path_text(grains, 'role') asc, jsonb_extract_path_text(pillar, 'role') desc",
		},
		{
			name:    "unselected pillar field rejected",
			orderBy: "pillar.missing",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OrderByJSONColumns(
				tt.orderBy,
				[]string{"minion_id", "alter_time"},
				map[string][]string{
					"grains": {"os", "role"},
					"pillar": {"roles.application", "role"},
				},
				"grains",
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf("OrderByJSONColumns() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("OrderByJSONColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}
