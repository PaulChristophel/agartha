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
			want:            "data @> '{\"grains\":{\"id\":\"pcmtest09.example.com\"}}'::jsonb AND data @> '{\"grains\":{\"os\":\"RedHat\"}}'::jsonb AND data @> '{\"grains\":{\"gtad\":true}}'::jsonb",
			wantErr:         false,
		},
		{
			name:            "Period in key",
			jsonPathFilters: []string{"\"grains\".\"id.test\":pcmtest09.example.com::string", "grains.os:RedHat::string", "grains.gtad:true::bool"},
			column:          "data",
			want:            "data @> '{\"grains\":{\"id.test\":\"pcmtest09.example.com\"}}'::jsonb AND data @> '{\"grains\":{\"os\":\"RedHat\"}}'::jsonb AND data @> '{\"grains\":{\"gtad\":true}}'::jsonb",
			wantErr:         false,
		},
		{
			name:            "Short Filters",
			jsonPathFilters: []string{"id:pcmtest09.example.com::string", "os:RedHat::string", "gtad:true::bool"},
			column:          "data",
			want:            "data @> '{\"id\":\"pcmtest09.example.com\"}'::jsonb AND data @> '{\"os\":\"RedHat\"}'::jsonb AND data @> '{\"gtad\":true}'::jsonb",
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
			want:            "data @> '{\"grains\":{\"dns\":{\"search\":\"gatech.edu\"}}}'::jsonb AND data @> '{\"grains\":{\"dns\":{\"sortlist\":[]}}}'::jsonb",
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
			name:            "Null Values",
			jsonPathFilters: []string{"grains.apparmor.profiles.1password:null::null"},
			column:          "data",
			want:            "data @> '{\"grains\":{\"apparmor\":{\"profiles\":{\"1password\":null}}}}'::jsonb",
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
		{
			name:            "Not Equals Filter",
			jsonPathFilters: []string{"grains.acc.installed:true::bool::not"},
			column:          "data",
			want:            "NOT (data @> '{\"grains\":{\"acc\":{\"installed\":true}}}'::jsonb)",
			wantErr:         false,
		},
		{
			name:            "LIKE Filter",
			jsonPathFilters: []string{"grains.os:RedHat%::string::like"},
			column:          "data",
			want:            "(data #>> '{\"grains\",\"os\"}') LIKE 'RedHat%'",
			wantErr:         false,
		},
		{
			name:            "NOT LIKE Filter",
			jsonPathFilters: []string{"grains.kernel:%windows%::string::not_like"},
			column:          "data",
			want:            "(data #>> '{\"grains\",\"kernel\"}') NOT LIKE '%windows%'",
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
