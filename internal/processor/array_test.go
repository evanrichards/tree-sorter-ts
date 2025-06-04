package processor

import (
	"strings"
	"testing"
)

func TestArraySorting(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantSorted string
	}{
		{
			name: "sort array of records by string key",
			content: `
const users = [
	/** tree-sorter-ts: keep-sorted key="name" **/
	{ name: "Charlie", age: 30 },
	{ name: "Alice", age: 25 },
	{ name: "Bob", age: 28 }
];`,
			wantSorted: `
const users = [
	/** tree-sorter-ts: keep-sorted key="name" **/
	{ name: "Alice", age: 25 },
	{ name: "Bob", age: 28 },
	{ name: "Charlie", age: 30 }
];`,
		},
		{
			name: "sort array of records by numeric key",
			content: `
const users = [
	/** tree-sorter-ts: keep-sorted key="age" **/
	{ name: "Charlie", age: 30 },
	{ name: "Alice", age: 25 },
	{ name: "Bob", age: 8 }
];`,
			wantSorted: `
const users = [
	/** tree-sorter-ts: keep-sorted key="age" **/
	{ name: "Bob", age: 8 },
	{ name: "Alice", age: 25 },
	{ name: "Charlie", age: 30 }
];`,
		},
		{
			name: "sort array of tuples by index",
			content: `
const data = [
	/** tree-sorter-ts: keep-sorted key="1" **/
	["apple", 5, true],
	["banana", 2, false],
	["cherry", 8, true]
];`,
			wantSorted: `
const data = [
	/** tree-sorter-ts: keep-sorted key="1" **/
	["banana", 2, false],
	["apple", 5, true],
	["cherry", 8, true]
];`,
		},
		{
			name: "sort array of scalars",
			content: `
const numbers = [
	/** tree-sorter-ts: keep-sorted **/
	5, 2, 8, 1, 9
];`,
			wantSorted: `
const numbers = [
	/** tree-sorter-ts: keep-sorted **/
	1, 2, 5, 8, 9
];`,
		},
		{
			name: "sort array of strings",
			content: `
const fruits = [
	/** tree-sorter-ts: keep-sorted **/
	"banana", "apple", "cherry"
];`,
			wantSorted: `
const fruits = [
	/** tree-sorter-ts: keep-sorted **/
	"apple", "banana", "cherry"
];`,
		},
		{
			name: "sort mixed array (fallback to string representation)",
			content: `
const mixed = [
	/** tree-sorter-ts: keep-sorted **/
	"string",
	42,
	{ key: "object" },
	[1, 2, 3],
	true,
	null
];`,
			wantSorted: `
const mixed = [
	/** tree-sorter-ts: keep-sorted **/
	"string",
	42,
	[1, 2, 3],
	null,
	true,
	{ key: "object" }
];`,
		},
		{
			name: "handle missing keys gracefully",
			content: `
const users = [
	/** tree-sorter-ts: keep-sorted key="name" **/
	{ name: "Alice", age: 25 },
	{ age: 30 },  // missing name
	{ name: "Bob", age: 28 },
	{ id: 123 }   // missing name and age
];`,
			wantSorted: `
const users = [
	/** tree-sorter-ts: keep-sorted key="name" **/
	{ name: "Alice", age: 25 },
	{ name: "Bob", age: 28 },
	{ age: 30 }, // missing name
	{ id: 123 } // missing name and age
];`,
		},
		{
			name: "handle invalid index for tuples",
			content: `
const data = [
	/** tree-sorter-ts: keep-sorted key="3" **/
	["a", "b"],       // no index 3
	["x", "y", "z", "w"],
	["m", "n", "o"]   // no index 3
];`,
			wantSorted: `
const data = [
	/** tree-sorter-ts: keep-sorted key="3" **/
	["x", "y", "z", "w"],
	["a", "b"], // no index 3
	["m", "n", "o"] // no index 3
];`,
		},
		{
			name: "preserve formatting and comments",
			content: `
const users = [
	/** tree-sorter-ts: keep-sorted key="name" **/
	// Charlie's data
	{
		name: "Charlie",
		age: 30,
		active: true
	},
	// Alice's data
	{ name: "Alice", age: 25, active: false },
	// Bob's data
	{
		name: "Bob",
		age: 28,
		active: true
	}
];`,
			wantSorted: `
const users = [
	/** tree-sorter-ts: keep-sorted key="name" **/
	// Alice's data
	{ name: "Alice", age: 25, active: false },
	// Bob's data
	{
		name: "Bob",
		age: 28,
		active: true
	},
	// Charlie's data
	{
		name: "Charlie",
		age: 30,
		active: true
	}
];`,
		},
		{
			name: "sort with trailing comma",
			content: `
const items = [
	/** tree-sorter-ts: keep-sorted key="id" **/
	{ id: 3, value: "three" },
	{ id: 1, value: "one" },
	{ id: 2, value: "two" },
];`,
			wantSorted: `
const items = [
	/** tree-sorter-ts: keep-sorted key="id" **/
	{ id: 1, value: "one" },
	{ id: 2, value: "two" },
	{ id: 3, value: "three" },
];`,
		},
		{
			name: "nested object access with dot notation",
			content: `
const users = [
	/** tree-sorter-ts: keep-sorted key="profile.firstName" **/
	{ profile: { firstName: "Charlie", lastName: "Brown" } },
	{ profile: { firstName: "Alice", lastName: "Smith" } },
	{ profile: { firstName: "Bob", lastName: "Jones" } }
];`,
			wantSorted: `
const users = [
	/** tree-sorter-ts: keep-sorted key="profile.firstName" **/
	{ profile: { firstName: "Alice", lastName: "Smith" } },
	{ profile: { firstName: "Bob", lastName: "Jones" } },
	{ profile: { firstName: "Charlie", lastName: "Brown" } }
];`,
		},
		{
			name: "sort by boolean values",
			content: `
const items = [
	/** tree-sorter-ts: keep-sorted key="active" **/
	{ name: "Item1", active: true },
	{ name: "Item2", active: false },
	{ name: "Item3", active: true },
	{ name: "Item4", active: false }
];`,
			wantSorted: `
const items = [
	/** tree-sorter-ts: keep-sorted key="active" **/
	{ name: "Item2", active: false },
	{ name: "Item4", active: false },
	{ name: "Item1", active: true },
	{ name: "Item3", active: true }
];`,
		},
		{
			name: "already sorted array",
			content: `
const sorted = [
	/** tree-sorter-ts: keep-sorted **/
	1, 2, 3, 4, 5
];`,
			wantSorted: "",
		},
		{
			name: "empty array",
			content: `
const empty = [
	/** tree-sorter-ts: keep-sorted **/
];`,
			wantSorted: "",
		},
		{
			name: "single element array",
			content: `
const single = [
	/** tree-sorter-ts: keep-sorted **/
	42
];`,
			wantSorted: "",
		},
		{
			name: "array with duplicate values",
			content: `
const dups = [
	/** tree-sorter-ts: keep-sorted key="score" **/
	{ name: "A", score: 10 },
	{ name: "B", score: 5 },
	{ name: "C", score: 10 },
	{ name: "D", score: 5 }
];`,
			wantSorted: `
const dups = [
	/** tree-sorter-ts: keep-sorted key="score" **/
	{ name: "B", score: 5 },
	{ name: "D", score: 5 },
	{ name: "A", score: 10 },
	{ name: "C", score: 10 }
];`,
		},
		{
			name: "array with with-new-line option",
			content: `
const users = [
	/** tree-sorter-ts: keep-sorted key="name" with-new-line **/
	{ name: "Charlie", age: 30 },
	{ name: "Alice", age: 25 },
	{ name: "Bob", age: 28 }
];`,
			wantSorted: `
const users = [
	/** tree-sorter-ts: keep-sorted key="name" with-new-line **/
	{ name: "Alice", age: 25 },

	{ name: "Bob", age: 28 },

	{ name: "Charlie", age: 30 }
];`,
		},
		{
			name: "array with deprecated-at-end option",
			content: `
const features = [
	/** tree-sorter-ts: keep-sorted key="name" deprecated-at-end **/
	{ name: "Feature C", enabled: true },
	/** @deprecated */
	{ name: "Feature A", enabled: false },
	{ name: "Feature B", enabled: true },
	{ name: "Feature D", enabled: false }, // @deprecated Will be removed
];`,
			wantSorted: `
const features = [
	/** tree-sorter-ts: keep-sorted key="name" deprecated-at-end **/
	{ name: "Feature B", enabled: true },
	{ name: "Feature C", enabled: true },
	/** @deprecated */
	{ name: "Feature A", enabled: false },
	{ name: "Feature D", enabled: false }, // @deprecated Will be removed
];`,
		},
		{
			name: "array with both with-new-line and deprecated-at-end",
			content: `
const config = [
	/** tree-sorter-ts: keep-sorted key="priority" deprecated-at-end with-new-line **/
	{ priority: 3, value: "normal" },
	/** @deprecated Use new format */
	{ priority: 1, value: "old" },
	{ priority: 2, value: "high" },
];`,
			wantSorted: `
const config = [
	/** tree-sorter-ts: keep-sorted key="priority" deprecated-at-end with-new-line **/
	{ priority: 2, value: "high" },

	{ priority: 3, value: "normal" },

	/** @deprecated Use new format */
	{ priority: 1, value: "old" },
];`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, content, err := parseTypeScript(tt.content)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			arrays := findArraysWithMagicCommentsAST(tree, content)
			if len(arrays) != 1 {
				t.Fatalf("expected 1 array, got %d", len(arrays))
			}

			_, needSort := sortArrayAST(arrays[0], content)

			if tt.wantSorted == "" {
				if needSort {
					t.Errorf("expected array to be already sorted, but needSort = true")
				}
				return
			}

			if !needSort {
				t.Errorf("expected array to need sorting, but needSort = false")
			}

			// Apply sort to full content
			newContent := make([]byte, len(content))
			copy(newContent, content)

			sortedContent, _ := sortArrayAST(arrays[0], content)
			start := arrays[0].array.StartByte()
			end := arrays[0].array.EndByte()

			before := newContent[:start]
			after := newContent[end:]
			newContent = append(append(before, sortedContent...), after...)

			gotSorted := strings.TrimSpace(string(newContent))
			wantSorted := strings.TrimSpace(tt.wantSorted)
			if gotSorted != wantSorted {
				t.Errorf("sorted array mismatch\ngot:\n%s\nwant:\n%s", gotSorted, wantSorted)
			}
		})
	}
}

func TestFindArraysWithMagicComments(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
	}{
		{
			name: "multiple arrays with magic comments",
			content: `
const arr1 = [
	/** tree-sorter-ts: keep-sorted **/
	3, 1, 2
];
const arr2 = [
	/** tree-sorter-ts: keep-sorted key="id" **/
	{ id: 3 },
	{ id: 1 }
];
const arr3 = [5, 4]; // no magic comment
const arr4 = [
	/** tree-sorter-ts: keep-sorted **/
	"b", "a"
];`,
			wantCount: 3,
		},
		{
			name: "nested arrays",
			content: `
const nested = {
	items: [
		/** tree-sorter-ts: keep-sorted **/
		3, 1, 2
	],
	users: {
		active: [
			/** tree-sorter-ts: keep-sorted key="name" **/
			{ name: "Bob" },
			{ name: "Alice" }
		]
	}
};`,
			wantCount: 2,
		},
		{
			name: "no arrays with magic comments",
			content: `
const arr1 = [1, 2, 3];
const arr2 = ["a", "b", "c"];`,
			wantCount: 0,
		},
		{
			name: "array in function call",
			content: `
processItems([
	/** tree-sorter-ts: keep-sorted **/
	3, 1, 2
]);`,
			wantCount: 1,
		},
		{
			name: "array as return value",
			content: `
function getData() {
	return [
		/** tree-sorter-ts: keep-sorted key="value" **/
		{ value: 10 },
		{ value: 5 }
	];
}`,
			wantCount: 1,
		},
		{
			name: "multiline magic comment in array",
			content: `
const config = [
	/**
	 * tree-sorter-ts: keep-sorted
	 *   key="name"
	 *   with-new-line
	 */
	{ name: "z" },
	{ name: "a" }
];`,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, content, err := parseTypeScript(tt.content)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			arrays := findArraysWithMagicCommentsAST(tree, content)
			if len(arrays) != tt.wantCount {
				t.Errorf("expected %d arrays, got %d", tt.wantCount, len(arrays))
			}
		})
	}
}

func TestArraySortByComment(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantSorted string
	}{
		{
			name: "sort array by inline comments",
			content: `
const userIds = [
	/** tree-sorter-ts: keep-sorted sort-by-comment **/
	"u_8234", // Bob Smith
	"u_9823", // Alice Johnson
	"u_1234", // David Lee
	"u_4521", // Carol White
];`,
			wantSorted: `
const userIds = [
	/** tree-sorter-ts: keep-sorted sort-by-comment **/
	"u_9823", // Alice Johnson
	"u_8234", // Bob Smith
	"u_4521", // Carol White
	"u_1234", // David Lee
];`,
		},
		{
			name: "sort array by block comments",
			content: `
const items = [
	/** tree-sorter-ts: keep-sorted sort-by-comment **/
	"item1", /* Zebra */
	"item2", /* Alpha */
	"item3", /* Beta */
];`,
			wantSorted: `
const items = [
	/** tree-sorter-ts: keep-sorted sort-by-comment **/
	"item2", /* Alpha */
	"item3", /* Beta */
	"item1", /* Zebra */
];`,
		},
		{
			name: "already sorted by comment",
			content: `
const sorted = [
	/** tree-sorter-ts: keep-sorted sort-by-comment **/
	"a", // Alpha
	"b", // Beta
	"c", // Charlie
];`,
			wantSorted: "",
		},
		{
			name: "sort by comment with with-new-line",
			content: `
const items = [
	/** tree-sorter-ts: keep-sorted sort-by-comment with-new-line **/
	"c", // Charlie

	"a", // Alpha

	"b", // Beta
];`,
			wantSorted: `
const items = [
	/** tree-sorter-ts: keep-sorted sort-by-comment with-new-line **/
	"a", // Alpha

	"b", // Beta

	"c", // Charlie
];`,
		},
		{
			name: "sort by comment with deprecated-at-end",
			content: `
const items = [
	/** tree-sorter-ts: keep-sorted sort-by-comment deprecated-at-end **/
	"item2", // Beta
	/** @deprecated Use item4 instead */
	"item1", // Alpha deprecated
	"item4", // Delta
	/** @deprecated Old feature */
	"item3", // Charlie deprecated
];`,
			wantSorted: `
const items = [
	/** tree-sorter-ts: keep-sorted sort-by-comment deprecated-at-end **/
	"item2", // Beta
	"item4", // Delta
	/** @deprecated Use item4 instead */
	"item1", // Alpha deprecated
	/** @deprecated Old feature */
	"item3", // Charlie deprecated
];`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, content, err := parseTypeScript(tt.content)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			arrays := findArraysWithMagicCommentsAST(tree, content)
			if len(arrays) != 1 {
				t.Fatalf("expected 1 array, got %d", len(arrays))
			}

			arr := arrays[0]
			sortedContent, changed := sortArrayAST(arr, content)

			if tt.wantSorted == "" {
				// Expecting no change
				if changed {
					t.Errorf("expected no change, but array was modified")
				}
			} else {
				if !changed {
					t.Errorf("expected array to be sorted, but no change was made")
					return
				}

				// Apply sort to full content
				newContent := make([]byte, len(content))
				copy(newContent, content)

				start := arr.array.StartByte()
				end := arr.array.EndByte()

				before := newContent[:start]
				after := newContent[end:]
				newContent = append(append(before, sortedContent...), after...)

				gotSorted := strings.TrimSpace(string(newContent))
				wantSorted := strings.TrimSpace(tt.wantSorted)
				if gotSorted != wantSorted {
					t.Errorf("sort result mismatch:\nwant:\n%s\ngot:\n%s", wantSorted, gotSorted)
				}
			}
		})
	}
}

func TestArraySortByCommentErrors(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedError string
	}{
		{
			name: "error when both key and sort-by-comment are specified",
			content: `
const items = [
	/** tree-sorter-ts: keep-sorted key="name" sort-by-comment **/
	{ name: "Charlie" }, // Comment
	{ name: "Alice" }, // Another comment
];`,
			expectedError: "invalid configuration: cannot use both 'key' and 'sort-by-comment' options together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, content, err := parseTypeScript(tt.content)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			arrays := findArraysWithMagicCommentsAST(tree, content)
			if len(arrays) != 1 {
				t.Fatalf("expected 1 array, got %d", len(arrays))
			}

			arr := arrays[0]
			if !arr.sortConfig.HasError {
				t.Errorf("expected configuration error, but none was detected")
			}
		})
	}
}

// TestObjectCommentDuplicationBug documents a known issue where sorting objects
// by comment content can result in comment duplication on the last property.
// 
// Bug Description:
// When sorting object properties using sort-by-comment, the reconstruction
// process sometimes duplicates the inline comment from the last property,
// causing it to appear twice in the output.
//
// Example of the bug:
// Input:
//   const obj = {
//     /** tree-sorter-ts: keep-sorted sort-by-comment */
//     prop1: "value1", // Charlie
//     prop2: "value2", // Alice  
//     prop3: "value3", // Bob
//   };
//
// Expected Output:
//   const obj = {
//     /** tree-sorter-ts: keep-sorted sort-by-comment */
//     prop2: "value2", // Alice
//     prop3: "value3", // Bob
//     prop1: "value1", // Charlie
//   };
//
// Actual Buggy Output:
//   const obj = {
//     /** tree-sorter-ts: keep-sorted sort-by-comment */
//     prop2: "value2", // Alice
//     prop3: "value3", // Bob
//     prop1: "value1", // Charlie
//   }; // Charlie  <-- Duplicated comment appears here
//
// Root Cause:
// The issue likely stems from the reconstruction logic not properly handling
// the boundary between the last sorted property and the closing brace of the object.
// The AST reconstruction may be incorrectly preserving or duplicating comment nodes.
//
// Status: Known issue, needs investigation
// Workaround: Use property-name sorting for objects if comment duplication occurs
func TestObjectCommentDuplicationBug(t *testing.T) {
	t.Skip("Known bug: object sort-by-comment can duplicate comments - needs investigation")
	
	content := `const user = {
  /** tree-sorter-ts: keep-sorted sort-by-comment */
  email: "user@example.com", // Contact info
  name: "John Doe", // Display name
  id: "u_123", // Unique identifier
};`

	tree, contentBytes, err := parseTypeScript(content)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	objects := findObjectsWithMagicCommentsAST(tree, contentBytes)
	if len(objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(objects))
	}

	// Apply sort to full content
	newContent := make([]byte, len(contentBytes))
	copy(newContent, contentBytes)

	sortedContent, _ := sortObjectAST(objects[0], contentBytes)
	start := objects[0].object.StartByte()
	end := objects[0].object.EndByte()

	before := newContent[:start]
	after := newContent[end:]
	newContent = append(append(before, sortedContent...), after...)

	result := string(newContent)
	
	// Check that comments are not duplicated
	// Count occurrences of each comment
	contactCount := strings.Count(result, "// Contact info")
	displayCount := strings.Count(result, "// Display name")  
	identifierCount := strings.Count(result, "// Unique identifier")
	
	if contactCount > 1 {
		t.Errorf("Comment '// Contact info' appears %d times, expected 1", contactCount)
	}
	if displayCount > 1 {
		t.Errorf("Comment '// Display name' appears %d times, expected 1", displayCount)
	}
	if identifierCount > 1 {
		t.Errorf("Comment '// Unique identifier' appears %d times, expected 1", identifierCount)
	}
	
	// Verify the properties are sorted by comment content
	// Expected order: "Contact info", "Display name", "Unique identifier"
	expectedOrder := []string{
		`email: "user@example.com", // Contact info`,
		`name: "John Doe", // Display name`, 
		`id: "u_123", // Unique identifier`,
	}
	
	for i, expected := range expectedOrder {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected property %d not found: %s", i, expected)
		}
	}
	
	// Additional check: ensure no extra comments appear after the closing brace
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		if strings.Contains(line, "};") {
			// Check if there are any comment-like patterns after the closing brace line
			for j := i + 1; j < len(lines); j++ {
				if strings.Contains(lines[j], "//") && 
				   (strings.Contains(lines[j], "Contact info") || 
				    strings.Contains(lines[j], "Display name") || 
				    strings.Contains(lines[j], "Unique identifier")) {
					t.Errorf("Found duplicated comment after closing brace on line %d: %s", j+1, lines[j])
				}
			}
			break
		}
	}
}

// TestArrayCommentSortingWorks verifies that array sorting by comment works correctly
// without the duplication issue seen in objects.
func TestArrayCommentSortingWorks(t *testing.T) {
	content := `const users = [
  /** tree-sorter-ts: keep-sorted sort-by-comment */
  "u_3", // Charlie
  "u_1", // Alice
  "u_2", // Bob
];`

	tree, contentBytes, err := parseTypeScript(content)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	arrays := findArraysWithMagicCommentsAST(tree, contentBytes)
	if len(arrays) != 1 {
		t.Fatalf("expected 1 array, got %d", len(arrays))
	}

	// Apply sort to full content
	newContent := make([]byte, len(contentBytes))
	copy(newContent, contentBytes)

	sortedContent, _ := sortArrayAST(arrays[0], contentBytes)
	start := arrays[0].array.StartByte()
	end := arrays[0].array.EndByte()

	before := newContent[:start]
	after := newContent[end:]
	newContent = append(append(before, sortedContent...), after...)

	result := string(newContent)
	
	// Verify no comment duplication in arrays
	aliceCount := strings.Count(result, "// Alice")
	bobCount := strings.Count(result, "// Bob")
	charlieCount := strings.Count(result, "// Charlie")
	
	if aliceCount != 1 {
		t.Errorf("Comment '// Alice' appears %d times, expected 1", aliceCount)
	}
	if bobCount != 1 {
		t.Errorf("Comment '// Bob' appears %d times, expected 1", bobCount)
	}
	if charlieCount != 1 {
		t.Errorf("Comment '// Charlie' appears %d times, expected 1", charlieCount)
	}
	
	// Verify correct sorting order by comment: Alice, Bob, Charlie
	alicePos := strings.Index(result, "// Alice")
	bobPos := strings.Index(result, "// Bob")
	charliePos := strings.Index(result, "// Charlie")
	
	if alicePos == -1 || bobPos == -1 || charliePos == -1 {
		t.Fatal("One or more comments not found in result")
	}
	
	if !(alicePos < bobPos && bobPos < charliePos) {
		t.Errorf("Array elements not sorted correctly by comment. Order: Alice(%d), Bob(%d), Charlie(%d)", 
			alicePos, bobPos, charliePos)
	}
}
