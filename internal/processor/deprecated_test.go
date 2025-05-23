package processor

import (
	"strings"
	"testing"
)

func TestParseSortConfigDeprecated(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    SortConfig
	}{
		{
			name:    "default_no_deprecated",
			comment: `/** tree-sorter-ts: keep-sorted */`,
			want:    SortConfig{WithNewLine: false, DeprecatedAtEnd: false},
		},
		{
			name:    "with_deprecated_at_end",
			comment: `/** tree-sorter-ts: keep-sorted deprecated-at-end */`,
			want:    SortConfig{WithNewLine: false, DeprecatedAtEnd: true},
		},
		{
			name:    "with_deprecated_and_newline",
			comment: `/** tree-sorter-ts: keep-sorted deprecated-at-end with-new-line */`,
			want:    SortConfig{WithNewLine: true, DeprecatedAtEnd: true},
		},
		{
			name:    "with_newline_and_deprecated",
			comment: `/** tree-sorter-ts: keep-sorted with-new-line deprecated-at-end */`,
			want:    SortConfig{WithNewLine: true, DeprecatedAtEnd: true},
		},
		{
			name:    "deprecated_extra_spaces",
			comment: `/**  tree-sorter-ts:  keep-sorted   deprecated-at-end  **/`,
			want:    SortConfig{WithNewLine: false, DeprecatedAtEnd: true},
		},
		{
			name: "deprecated_multiline_comment",
			comment: `/**
			 * tree-sorter-ts: keep-sorted deprecated-at-end
			 */`,
			want: SortConfig{WithNewLine: false, DeprecatedAtEnd: true},
		},
		{
			name:    "single_star_comment_deprecated",
			comment: `/* tree-sorter-ts: keep-sorted deprecated-at-end */`,
			want:    SortConfig{WithNewLine: false, DeprecatedAtEnd: true},
		},
		{
			name: "multiline_flags_on_separate_lines",
			comment: `/**
			 * tree-sorter-ts: keep-sorted
			 * deprecated-at-end
			 * with-new-line
			 */`,
			want: SortConfig{WithNewLine: true, DeprecatedAtEnd: true},
		},
		{
			name: "multiline_flags_with_extra_asterisks",
			comment: `/**
			 * tree-sorter-ts: keep-sorted
			 *   with-new-line
			 *   deprecated-at-end
			 */`,
			want: SortConfig{WithNewLine: true, DeprecatedAtEnd: true},
		},
		{
			name: "multiline_mixed_same_line",
			comment: `/**
			 * tree-sorter-ts: keep-sorted
			 * deprecated-at-end with-new-line
			 */`,
			want: SortConfig{WithNewLine: true, DeprecatedAtEnd: true},
		},
		{
			name: "multiline_no_asterisks",
			comment: `/** tree-sorter-ts: keep-sorted
			    deprecated-at-end
			    with-new-line **/`,
			want: SortConfig{WithNewLine: true, DeprecatedAtEnd: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSortConfig([]byte(tt.comment))
			if got.WithNewLine != tt.want.WithNewLine {
				t.Errorf("parseSortConfig() WithNewLine = %v, want %v", got.WithNewLine, tt.want.WithNewLine)
			}
			if got.DeprecatedAtEnd != tt.want.DeprecatedAtEnd {
				t.Errorf("parseSortConfig() DeprecatedAtEnd = %v, want %v", got.DeprecatedAtEnd, tt.want.DeprecatedAtEnd)
			}
		})
	}
}

func TestDeprecatedAtEnd(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantSorted   string
		wantNeedSort int
	}{
		{
			name: "deprecated_properties_at_end",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  gamma: true,
  /** @deprecated Use newApi instead */
  oldApi: "old",
  alpha: "first",
  newApi: "new",
  /** @deprecated */
  beta: false,
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  alpha: "first",
  gamma: true,
  newApi: "new",
  /** @deprecated */
  beta: false,
  /** @deprecated Use newApi instead */
  oldApi: "old",
};`,
			wantNeedSort: 1,
		},
		{
			name: "deprecated_in_inline_comment",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  gamma: 123,
  oldSetting: "old", // @deprecated use newSetting
  delta: true,
  epsilon: "test",
  newSetting: "new",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  delta: true,
  epsilon: "test",
  gamma: 123,
  newSetting: "new",
  oldSetting: "old", // @deprecated use newSetting
};`,
			wantNeedSort: 1,
		},
		{
			name: "mixed_deprecated_annotations",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  active: true,
  /** @deprecated Will be removed in v2.0 */
  oldFeature: false,
  beta: "test",
  /** This is not deprecated */
  alpha: "first",
  legacyMode: true, // @deprecated
  newFeature: true,
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  active: true,
  /** This is not deprecated */
  alpha: "first",
  beta: "test",
  newFeature: true,
  legacyMode: true, // @deprecated
  /** @deprecated Will be removed in v2.0 */
  oldFeature: false,
};`,
			wantNeedSort: 1,
		},
		{
			name: "no_deprecated_properties",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  zebra: "last",
  alpha: "first",
  beta: "second",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  alpha: "first",
  beta: "second",
  zebra: "last",
};`,
			wantNeedSort: 1,
		},
		{
			name: "deprecated_already_sorted",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  active: true,
  beta: "test",
  /** @deprecated */
  oldApi: "old",
  /** @deprecated */
  oldFeature: false,
};`,
			wantSorted:   "", // No change
			wantNeedSort: 0,
		},
		{
			name: "deprecated_with_newline",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end with-new-line **/
  gamma: true,
  /** @deprecated */
  beta: false,
  alpha: "first",
  newApi: "new",
  /** @deprecated Use newApi instead */
  oldApi: "old",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end with-new-line **/
  alpha: "first",

  gamma: true,

  newApi: "new",

  /** @deprecated */
  beta: false,

  /** @deprecated Use newApi instead */
  oldApi: "old",
};`,
			wantNeedSort: 1,
		},
		{
			name: "deprecated_with_trailing_comma",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  gamma: true,
  /** @deprecated */
  oldApi: "old",
  alpha: "first",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  alpha: "first",
  gamma: true,
  /** @deprecated */
  oldApi: "old",
};`,
			wantNeedSort: 1,
		},
		{
			name: "all_properties_deprecated",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  /** @deprecated */
  gamma: true,
  /** @deprecated Use newApi instead */
  beta: "old",
  /** @deprecated */
  alpha: "first",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  /** @deprecated */
  alpha: "first",
  /** @deprecated Use newApi instead */
  beta: "old",
  /** @deprecated */
  gamma: true,
};`,
			wantNeedSort: 1,
		},
		{
			name: "deprecated_without_flag",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  gamma: true,
  /** @deprecated Use newApi instead */
  oldApi: "old",
  alpha: "first",
  newApi: "new",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: "first",
  gamma: true,
  newApi: "new",
  /** @deprecated Use newApi instead */
  oldApi: "old",
};`,
			wantNeedSort: 1,
		},
		{
			name: "deprecated_multiline_comment",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  gamma: true,
  /** 
   * @deprecated This is a multiline
   * deprecated comment that should
   * still be detected
   */
  oldApi: "old",
  alpha: "first",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  alpha: "first",
  gamma: true,
  /** 
   * @deprecated This is a multiline
   * deprecated comment that should
   * still be detected
   */
  oldApi: "old",
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
