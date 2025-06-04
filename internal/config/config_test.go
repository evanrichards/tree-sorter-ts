package config

import (
	"testing"
)

func TestParseSortConfig(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    SortConfig
	}{
		{
			name:    "basic keep-sorted",
			comment: "/** tree-sorter-ts: keep-sorted */",
			want:    SortConfig{},
		},
		{
			name:    "with-new-line option",
			comment: "/** tree-sorter-ts: keep-sorted with-new-line */",
			want:    SortConfig{WithNewLine: true},
		},
		{
			name:    "deprecated-at-end option",
			comment: "/** tree-sorter-ts: keep-sorted deprecated-at-end */",
			want:    SortConfig{DeprecatedAtEnd: true},
		},
		{
			name:    "sort-by-comment option",
			comment: "/** tree-sorter-ts: keep-sorted sort-by-comment */",
			want:    SortConfig{SortByComment: true},
		},
		{
			name:    "key option with quotes",
			comment: `/** tree-sorter-ts: keep-sorted key="name" */`,
			want:    SortConfig{Key: "name"},
		},
		{
			name:    "multiple options",
			comment: "/** tree-sorter-ts: keep-sorted deprecated-at-end with-new-line */",
			want:    SortConfig{DeprecatedAtEnd: true, WithNewLine: true},
		},
		{
			name:    "multiline comment",
			comment: `/**
			 * tree-sorter-ts: keep-sorted
			 *   with-new-line
			 *   deprecated-at-end
			 */`,
			want: SortConfig{WithNewLine: true, DeprecatedAtEnd: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSortConfig([]byte(tt.comment))
			if got.WithNewLine != tt.want.WithNewLine {
				t.Errorf("WithNewLine = %v, want %v", got.WithNewLine, tt.want.WithNewLine)
			}
			if got.DeprecatedAtEnd != tt.want.DeprecatedAtEnd {
				t.Errorf("DeprecatedAtEnd = %v, want %v", got.DeprecatedAtEnd, tt.want.DeprecatedAtEnd)
			}
			if got.SortByComment != tt.want.SortByComment {
				t.Errorf("SortByComment = %v, want %v", got.SortByComment, tt.want.SortByComment)
			}
			if got.Key != tt.want.Key {
				t.Errorf("Key = %q, want %q", got.Key, tt.want.Key)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    SortConfig
		wantError bool
	}{
		{
			name:      "valid: property name sorting",
			config:    SortConfig{},
			wantError: false,
		},
		{
			name:      "valid: sort by comment",
			config:    SortConfig{SortByComment: true},
			wantError: false,
		},
		{
			name:      "valid: sort by key",
			config:    SortConfig{Key: "name"},
			wantError: false,
		},
		{
			name:      "invalid: both key and sort-by-comment",
			config:    SortConfig{Key: "name", SortByComment: true},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && !tt.config.HasError {
				t.Errorf("Expected HasError to be set when validation fails")
			}
		})
	}
}