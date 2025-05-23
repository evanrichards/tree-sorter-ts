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
