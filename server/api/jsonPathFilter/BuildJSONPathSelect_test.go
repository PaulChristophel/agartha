package jsonPathFilter

import (
	"testing"
)

func TestBuildJSONPathSelect(t *testing.T) {
	tests := []struct {
		name      string
		jsonPaths []string
		column    string
		want      string
	}{
		{
			name:      "Single Path",
			jsonPaths: []string{"grains.os"},
			column:    "data",
			want:      "jsonb_build_object('os', jsonb_path_query(data, '$.grains.os')) AS data",
		},
		{
			name:      "Multiple Paths",
			jsonPaths: []string{"grains.os", "grains.id"},
			column:    "data",
			want:      "jsonb_build_object('os', jsonb_path_query(data, '$.grains.os'), 'id', jsonb_path_query(data, '$.grains.id')) AS data",
		},
		{
			name:      "Path with Period in Key",
			jsonPaths: []string{`store."book.author"`, `store.book.title`},
			column:    "data",
			want:      "jsonb_build_object('book.author', jsonb_path_query(data, '$.store.book.author'), 'title', jsonb_path_query(data, '$.store.book.title')) AS data",
		},
		{
			name:      "Path with Special Characters",
			jsonPaths: []string{`config."app-name"`, `config."log/level"`, `config."time:zone"`, `config."braces{}"`, `config."parens()"`},
			column:    "data",
			want:      "jsonb_build_object('app-name', jsonb_path_query(data, '$.config.\"app-name\"'), 'log/level', jsonb_path_query(data, '$.config.\"log/level\"'), 'time:zone', jsonb_path_query(data, '$.config.\"time:zone\"'), 'braces{}', jsonb_path_query(data, '$.config.\"braces{}\"'), 'parens()', jsonb_path_query(data, '$.config.\"parens()\"')) AS data",
		},
		// {
		// 	name:      "Array Index in Path",
		// 	jsonPaths: []string{`store.books[0].title`, `store.books[1].author`},
		// 	column:    "data",
		// 	want:      "jsonb_build_object('title', jsonb_path_query(data, '$.store.books[0].title'), 'author', jsonb_path_query(data, '$.store.books[1].author')) AS data",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildJSONPathSelect(tt.jsonPaths, tt.column); got != tt.want {
				t.Errorf("BuildJSONPathSelect() = %v, want %v", got, tt.want)
			}
		})
	}
}
