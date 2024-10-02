package jsonPathFilter

import (
	"testing"
)

func TestBuildJSONPathWhere(t *testing.T) {
	tests := []struct {
		name            string
		jsonPathFilters []string
		column          string
		want            string
		wantErr         bool
	}{
		{
			name:            "Valid Filters",
			jsonPathFilters: []string{"grains.id:pcmtest09.example.com::string", "grains.os:RedHat::string", "grains.gtad:true::bool"},
			column:          "data",
			want:            "data @> '{\"grains\":{\"gtad\":true,\"id\":\"pcmtest09.example.com\",\"os\":\"RedHat\"}}'::jsonb",
			wantErr:         false,
		},
		{
			name:            "Period in key",
			jsonPathFilters: []string{"\"grains\".\"id.test\":pcmtest09.example.com::string", "grains.os:RedHat::string", "grains.gtad:true::bool"},
			column:          "data",
			want:            "data @> '{\"grains\":{\"gtad\":true,\"id.test\":\"pcmtest09.example.com\",\"os\":\"RedHat\"}}'::jsonb",
			wantErr:         false,
		},
		{
			name:            "Short Filters",
			jsonPathFilters: []string{"id:pcmtest09.example.com::string", "os:RedHat::string", "gtad:true::bool"},
			column:          "data",
			want:            "data @> '{\"gtad\":true,\"id\":\"pcmtest09.example.com\",\"os\":\"RedHat\"}'::jsonb",
			wantErr:         false,
		},
		{
			name:            "Invalid Filter Format",
			jsonPathFilters: []string{"grains.id:pcmtest09.example.com:string"},
			column:          "data",
			want:            "",
			wantErr:         true,
		},
		{
			name:            "Invalid Key Length",
			jsonPathFilters: []string{"::string"},
			column:          "data",
			want:            "",
			wantErr:         true,
		},
		{
			name:            "Nested JSON Objects",
			jsonPathFilters: []string{"grains.dns.nameservers:143.215.77.4::string"},
			column:          "data",
			want:            "data @> '{\"grains\":{\"dns\":{\"nameservers\":\"143.215.77.4\"}}}'::jsonb",
			wantErr:         false,
		},
		{
			name:            "Multiple Nested JSON Objects",
			jsonPathFilters: []string{"grains.dns.search:gatech.edu::string", "grains.dns.sortlist:[]::array"},
			column:          "data",
			want:            "data @> '{\"grains\":{\"dns\":{\"search\":\"gatech.edu\",\"sortlist\":[]}}}'::jsonb",
			wantErr:         false,
		},
		{
			name:            "Boolean Values",
			jsonPathFilters: []string{"grains.efi:false::bool"},
			column:          "data",
			want:            "data @> '{\"grains\":{\"efi\":false}}'::jsonb",
			wantErr:         false,
		},
		{
			name:            "Integer Values",
			jsonPathFilters: []string{"grains.gid:0::int"},
			column:          "data",
			want:            "data @> '{\"grains\":{\"gid\":0}}'::jsonb",
			wantErr:         false,
		},
		{
			name:            "Float Values",
			jsonPathFilters: []string{"grains.memory.size:16.1::float"},
			column:          "data",
			want:            "data @> '{\"grains\":{\"memory\":{\"size\":16.1}}}'::jsonb",
			wantErr:         false,
		},
		{
			name:            "Array Values",
			jsonPathFilters: []string{"grains.dns.ip4_nameservers:[143.215.77.4,130.207.244.251,130.207.244.244]::array"},
			column:          "data",
			want:            "data @> '{\"grains\":{\"dns\":{\"ip4_nameservers\":[\"143.215.77.4\",\"130.207.244.251\",\"130.207.244.244\"]}}}'::jsonb",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildJSONPathWhere(tt.jsonPathFilters, tt.column)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildJSONPathWhere() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BuildJSONPathWhere() = %v, want %v", got, tt.want)
			}
		})
	}
}
