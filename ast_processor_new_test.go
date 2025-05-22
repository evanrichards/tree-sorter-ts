package main

import (
	"strings"
	"testing"
)

func TestParseSortConfig(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    SortConfig
	}{
		{
			name:    "default_no_options",
			comment: `/** tree-sorter-ts: keep-sorted */`,
			want:    SortConfig{WithNewLine: false},
		},
		{
			name:    "with_new_line",
			comment: `/** tree-sorter-ts: keep-sorted with-new-line */`,
			want:    SortConfig{WithNewLine: true},
		},
		{
			name:    "with_new_line_extra_spaces",
			comment: `/**  tree-sorter-ts:  keep-sorted   with-new-line  **/`,
			want:    SortConfig{WithNewLine: true},
		},
		{
			name: "with_new_line_multiline_comment",
			comment: `/**
			 * tree-sorter-ts: keep-sorted with-new-line
			 */`,
			want: SortConfig{WithNewLine: true},
		},
		{
			name:    "single_star_comment",
			comment: `/* tree-sorter-ts: keep-sorted with-new-line */`,
			want:    SortConfig{WithNewLine: true},
		},
		{
			name:    "multiple_asterisks",
			comment: `/** tree-sorter-ts: keep-sorted with-new-line ***/`,
			want:    SortConfig{WithNewLine: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSortConfig([]byte(tt.comment))
			if got.WithNewLine != tt.want.WithNewLine {
				t.Errorf("parseSortConfig() WithNewLine = %v, want %v", got.WithNewLine, tt.want.WithNewLine)
			}
		})
	}
}

func TestSortObjectASTWithNewLine(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantSorted   string
		wantNeedSort int
	}{
		{
			name: "with_new_line_basic",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted with-new-line **/
  zebra: "value1",
  alpha: "value2",
  beta: "value3",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted with-new-line **/
  alpha: "value2",

  beta: "value3",

  zebra: "value1",
};`,
			wantNeedSort: 1,
		},
		{
			name: "with_new_line_with_comments",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted with-new-line **/
  // Comment for zebra
  zebra: "value1",
  // Comment for alpha
  alpha: "value2", // inline comment
  beta: "value3",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted with-new-line **/
  // Comment for alpha
  alpha: "value2", // inline comment

  beta: "value3",

  // Comment for zebra
  zebra: "value1",
};`,
			wantNeedSort: 1,
		},
		{
			name: "with_new_line_already_sorted",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted with-new-line **/
  alpha: "value2",

  beta: "value3",

  zebra: "value1",
};`,
			wantSorted:   "", // No change
			wantNeedSort: 0,
		},
		{
			name: "without_new_line_option",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  zebra: "value1",
  alpha: "value2",
  beta: "value3",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: "value2",
  beta: "value3",
  zebra: "value1",
};`,
			wantNeedSort: 1,
		},
		{
			name: "with_new_line_trailing_comma",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted with-new-line **/
  zebra: "value1",
  alpha: "value2",
  beta: "value3",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted with-new-line **/
  alpha: "value2",

  beta: "value3",

  zebra: "value1",
};`,
			wantNeedSort: 1,
		},
		{
			name: "add_new_line_to_already_sorted",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted with-new-line **/
  alpha: "value2",
  beta: "value3",
  zebra: "value1",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted with-new-line **/
  alpha: "value2",

  beta: "value3",

  zebra: "value1",
};`,
			wantNeedSort: 1,
		},
		{
			name: "remove_new_line_from_already_sorted",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: "value2",

  beta: "value3",

  zebra: "value1",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: "value2",
  beta: "value3",
  zebra: "value1",
};`,
			wantNeedSort: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test detection
			result := ProcessResult{}
			root, contentBytes, err := parseTypeScript(tt.content)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			objects := findObjectsWithMagicCommentsAST(root, contentBytes)
			result.ObjectsFound = len(objects)

			// Count how many need sorting
			for _, obj := range objects {
				_, needsSort := sortObjectAST(obj, contentBytes)
				if needsSort {
					result.ObjectsNeedSort++
				}
			}

			if result.ObjectsNeedSort != tt.wantNeedSort {
				t.Errorf("ObjectsNeedSort = %d, want %d", result.ObjectsNeedSort, tt.wantNeedSort)
			}

			// Test mutation if sorting is needed
			if tt.wantSorted != "" {
				// Apply sorts
				newContent := make([]byte, len(contentBytes))
				copy(newContent, contentBytes)

				// Sort from end to beginning
				for i := len(objects) - 1; i >= 0; i-- {
					sortedContent, needsSort := sortObjectAST(objects[i], newContent)
					if needsSort {
						start := objects[i].object.StartByte()
						end := objects[i].object.EndByte()

						before := newContent[:start]
						after := newContent[end:]
						newContent = append(append(before, sortedContent...), after...)
					}
				}

				got := string(newContent)
				want := tt.wantSorted

				// Normalize whitespace for comparison
				got = strings.TrimSpace(got)
				want = strings.TrimSpace(want)

				if got != want {
					t.Errorf("Sorted output mismatch.\nGot:\n%s\n\nWant:\n%s", got, want)
				}
			}
		})
	}
}
