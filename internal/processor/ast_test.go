package processor

import (
	"context"
	"strings"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

func parseTypeScript(content string) (*sitter.Node, []byte, error) {
	parser := parserPool.Get().(*sitter.Parser)
	defer parserPool.Put(parser)

	contentBytes := []byte(content)
	tree, err := parser.ParseCtx(context.Background(), nil, contentBytes)
	if err != nil {
		return nil, nil, err
	}
	return tree.RootNode(), contentBytes, nil
}

func TestFindObjectsWithMagicComments(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantCount   int
		wantIndices []int // Expected magic comment indices
	}{
		{
			name: "single_object",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  z: 1,
  a: 2,
};`,
			wantCount:   1,
			wantIndices: []int{1}, // Magic comment is at index 1 in object children
		},
		{
			name: "multiple_objects",
			content: `const a = {
  /** tree-sorter-ts: keep-sorted **/
  z: 1,
};
const b = {
  /** tree-sorter-ts: keep-sorted */
  y: 2,
};`,
			wantCount:   2,
			wantIndices: []int{1, 1},
		},
		{
			name:        "no_magic_comment",
			content:     `const a = { x: 1, y: 2 };`,
			wantCount:   0,
			wantIndices: []int{},
		},
		{
			name: "nested_objects",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  nested: {
    /** tree-sorter-ts: keep-sorted **/
    z: 1,
    a: 2,
  },
  other: true,
};`,
			wantCount:   2,
			wantIndices: []int{1, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, content, err := parseTypeScript(tt.content)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			objects := findObjectsWithMagicCommentsAST(root, content)

			if len(objects) != tt.wantCount {
				t.Errorf("Found %d objects, want %d", len(objects), tt.wantCount)
			}

			for i, obj := range objects {
				if i < len(tt.wantIndices) && obj.magicIndex != tt.wantIndices[i] {
					t.Errorf("Object %d: magic comment at index %d, want %d",
						i, obj.magicIndex, tt.wantIndices[i])
				}
			}
		})
	}
}

func TestSortObjectAST(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantSorted   string
		wantNeedSort int
	}{
		{
			name: "basic_sort",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  zebra: "value1",
  alpha: "value2", // critical setting
  beta: "value3",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: "value2", // critical setting
  beta: "value3",
  zebra: "value1",
};`,
			wantNeedSort: 1,
		},
		{
			name: "already_sorted",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: "value2",
  beta: "value3",
  zebra: "value1",
};`,
			wantSorted:   "", // No change
			wantNeedSort: 0,
		},
		{
			name: "computed_keys",
			content: `const handlers = {
  /** tree-sorter-ts: keep-sorted **/
  [Status.PENDING]: handle1,
  [Status.ACTIVE]: handle2,
  [Status.COMPLETED]: handle3,
};`,
			wantSorted: `const handlers = {
  /** tree-sorter-ts: keep-sorted **/
  [Status.ACTIVE]: handle2,
  [Status.COMPLETED]: handle3,
  [Status.PENDING]: handle1,
};`,
			wantNeedSort: 1,
		},
		{
			name:         "multiline_values",
			content:      "const messages = {\n  /** tree-sorter-ts: keep-sorted **/\n  error: `This is\na multiline\nerror`,\n  alert: \"Alert!\",\n};",
			wantSorted:   "const messages = {\n  /** tree-sorter-ts: keep-sorted **/\n  alert: \"Alert!\",\n  error: `This is\na multiline\nerror`,\n};",
			wantNeedSort: 1,
		},
		{
			name: "block_comments",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  zebra: false,
  /**
   * Important setting
   */
  alpha: true,
  beta: "test",
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  /**
   * Important setting
   */
  alpha: true,
  beta: "test",
  zebra: false,
};`,
			wantNeedSort: 1,
		},
		{
			name: "trailing_comma",
			content: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  z: 1,
  a: 2,
  b: 3,
};`,
			wantSorted: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  a: 2,
  b: 3,
  z: 1,
};`,
			wantNeedSort: 1,
		},
		{
			name: "multiple_objects_in_file",
			content: `const a = {
  /** tree-sorter-ts: keep-sorted **/
  z: 1,
  a: 2,
};

const b = {
  /** tree-sorter-ts: keep-sorted **/
  beta: true,
  alpha: false,
};`,
			wantSorted: `const a = {
  /** tree-sorter-ts: keep-sorted **/
  a: 2,
  z: 1,
};

const b = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: false,
  beta: true,
};`,
			wantNeedSort: 2,
		},
		{
			name: "record_type_with_semicolon",
			content: `export const TEST: Record<string, string> = {
  /** tree-sorter-ts: keep-sorted */
  zebra: "z value",
  alpha: "a value",
  beta: "b value",
};`,
			wantSorted: `export const TEST: Record<string, string> = {
  /** tree-sorter-ts: keep-sorted */
  alpha: "a value",
  beta: "b value",
  zebra: "z value",
};`,
			wantNeedSort: 1,
		},
		{
			name: "complex_multiline_values",
			content: `export const DESCRIPTIONS: Record<string, string> = {
  /** tree-sorter-ts: keep-sorted */
  [Types.Z]:
    DESCRIPTIONS[
      Type.BASE
    ].replace(/text/g, 'replaced'),
  [Types.A]:
    'A long description that spans a really long line and contains various details about the type',
  [Types.B]: "B value",
};`,
			wantSorted: `export const DESCRIPTIONS: Record<string, string> = {
  /** tree-sorter-ts: keep-sorted */
  [Types.A]:
    'A long description that spans a really long line and contains various details about the type',
  [Types.B]: "B value",
  [Types.Z]:
    DESCRIPTIONS[
      Type.BASE
    ].replace(/text/g, 'replaced'),
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

func TestExtractKeyAST(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string // Expected keys in order
	}{
		{
			name: "simple_keys",
			content: `{
  /** tree-sorter-ts: keep-sorted **/
  alpha: 1,
  beta: 2,
  "gamma": 3,
  'delta': 4,
}`,
			want: []string{"alpha", "beta", "gamma", "delta"},
		},
		{
			name: "computed_keys",
			content: `{
  /** tree-sorter-ts: keep-sorted **/
  [Enum.A]: 1,
  [Enum.B]: 2,
  ["computed"]: 3,
}`,
			want: []string{"[Enum.A]", "[Enum.B]", "[\"computed\"]"},
		},
		{
			name: "mixed_keys",
			content: `{
  /** tree-sorter-ts: keep-sorted **/
  regular: 1,
  "quoted": 2,
  [computed]: 3,
  123: 4,
}`,
			want: []string{"regular", "quoted", "[computed]", "123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, contentBytes, err := parseTypeScript(tt.content)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			objects := findObjectsWithMagicCommentsAST(root, contentBytes)
			if len(objects) != 1 {
				t.Fatalf("Expected 1 object, got %d", len(objects))
			}

			properties := extractPropertiesAST(objects[0], contentBytes)

			if len(properties) != len(tt.want) {
				t.Errorf("Got %d properties, want %d", len(properties), len(tt.want))
			}

			for i, prop := range properties {
				if i < len(tt.want) && prop.key != tt.want[i] {
					t.Errorf("Property %d: key = %q, want %q", i, prop.key, tt.want[i])
				}
			}
		})
	}
}
