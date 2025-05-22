# Tree-Sorter-TS: TypeScript Code Sorting Utility

## Overview

A Go CLI utility that automatically sorts TypeScript arrays and object literals marked with special comments, while preserving formatting, comments, and semantic structure using Tree-sitter parsing.

## Core Functionality

The tool scans TypeScript files for the magic comment `/** tree-sorter-ts: keep-sorted **/` and sorts the contents of the immediately following object literal by property keys.

### Example Usage

**Basic object sorting:**

```typescript
const config = {
  /** tree-sorter-ts: keep-sorted **/
  zebra: "value1",
  alpha: "value2", // critical setting
  beta: "value3",
};
```

**Output:**

```typescript
const config = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: "value2", // critical setting
  beta: "value3",
  zebra: "value1",
};
```

**Block comments above keys:**

```typescript
const settings = {
  /** tree-sorter-ts: keep-sorted **/
  zebra: false,
  /**
   * This is a critical configuration
   * that affects the entire system
   */
  alpha: true,
  beta: "test", // inline comment
};
```

**Output:**

```typescript
const settings = {
  /** tree-sorter-ts: keep-sorted **/
  /**
   * This is a critical configuration
   * that affects the entire system
   */
  alpha: true,
  beta: "test", // inline comment
  zebra: false,
};
```

**Computed property keys (enum references):**

```typescript
const handlers = {
  /** tree-sorter-ts: keep-sorted **/
  [StatusEnum.PENDING]: handlePending,
  [StatusEnum.ACTIVE]: handleActive, // primary handler
  [StatusEnum.COMPLETED]: handleCompleted,
};
```

**Output:**

```typescript
const handlers = {
  /** tree-sorter-ts: keep-sorted **/
  [StatusEnum.ACTIVE]: handleActive, // primary handler
  [StatusEnum.COMPLETED]: handleCompleted,
  [StatusEnum.PENDING]: handlePending,
};
```

**Unattached comments (moved to top):**

```typescript
const config = {
  /** tree-sorter-ts: keep-sorted **/
  zebra: "value1",
  alpha: "value2", // attached comment

  /**
   * This comment is not attached to any property
   */

  beta: "value3",
};
```

**Output:**

```typescript
const config = {
  /** tree-sorter-ts: keep-sorted **/
  /**
   * This comment is not attached to any property
   */

  alpha: "value2", // attached comment
  beta: "value3",
  zebra: "value1",
};
```

## Technical Requirements

### Tree-sitter Integration

- Use `github.com/smacker/go-tree-sitter` with TypeScript grammar
- Parse files into AST to preserve exact formatting and comments
- Query for comment nodes containing the magic string
- Identify the subsequent array/object literal to sort

### Sorting Logic

- **Objects**: Sort by property keys (alphabetical)
- **String Keys**: Standard alphabetical sort (`"alpha"`, `"beta"`, `"zebra"`)
- **Computed Keys**: Sort by the string representation of the computed expression (`[SomeEnum.ACTIVE]` → `"SomeEnum.ACTIVE"`)
- **Comment Preservation**: Maintain inline and block comments with their associated properties during sorting
- **Unattached Comments**: Comments not directly associated with a property are moved to the top of the object (after the magic comment)

### File Processing

- Walk directory trees recursively
- Process `.ts` and `.tsx` files
- Modify files in-place or output to specified directory
- Handle multiple magic comments per file

## Implementation Approach

Based on the provided Tree-sitter examples, the implementation should:

1. **File Discovery**: Recursively walk directories for TypeScript files
2. **AST Parsing**: Use Tree-sitter to parse each file into an AST
3. **Query Execution**: Write Tree-sitter queries to find magic comments and associated structures
4. **Node Manipulation**: Extract, sort, and reconstruct the relevant AST nodes
5. **Code Generation**: Convert the modified AST back to source code

### Key Tree-sitter Queries Needed

```scheme
;; Find magic comments
(comment) @magic-comment
(#match? @magic-comment "tree-sorter-ts: keep-sorted")

;; Find object literals following magic comments
(object
  (pair) @property)*

;; Handle computed property names
(object
  (pair
    key: (computed_property_name) @computed-key
    value: (_) @value
  )
) @property

;; Handle regular property names with comments
(object
  (pair
    key: (property_identifier) @key
    value: (_) @value
  )
) @property
```

## CLI Interface

```bash
tree-sorter-ts [flags] <path>

Flags:
  --check         Check if files are sorted (exit 1 if not)
  --write         Write changes to files (default: dry-run)
  --recursive     Process directories recursively (default: true)
  --extensions    File extensions to process (default: .ts,.tsx)
  --workers       Number of parallel workers (default: number of CPUs, max 8)
```

## Performance

The tool processes files in parallel for optimal performance on large codebases:

- Uses a worker pool pattern with goroutines
- Default worker count is the number of CPU cores (max 8)
- Parser instances are pooled and reused
- Files are pre-filtered using regex before parsing for efficiency

## Features

- ✅ Supports multi-line property values (template literals, multi-line strings)
- ✅ Handles computed property keys (e.g., `[EnumValue.KEY]`)
- ✅ Preserves all comments and formatting
- ✅ Works with both `/** tree-sorter-ts: keep-sorted **/` and `/** tree-sorter-ts: keep-sorted */` formats
- ✅ Processes files in parallel for performance

## Error Handling

- Graceful handling of parse errors (skip malformed files with warnings)
- Preserve original files on sorting failures
- Clear error messages for invalid magic comment placement
- Validation that magic comments precede sortable structures

## Success Criteria

1. Correctly identifies and sorts marked object literals by property keys
2. Preserves all comments and formatting outside sorted regions
3. Maintains comment associations with properties during sorting (both inline and block comments)
4. Handles computed property keys (enum references, expressions)
5. Handles edge cases (empty objects, single properties, nested objects)
6. Provides clear feedback on processed files and any issues

## Dependencies

- `github.com/smacker/go-tree-sitter` - Core Tree-sitter bindings
- `github.com/smacker/go-tree-sitter/typescript/typescript` - TypeScript grammar
- Standard library for file system operations and CLI

This utility will be particularly valuable for maintaining consistent ordering in configuration objects, import/export lists, and other structured data in TypeScript codebases.
