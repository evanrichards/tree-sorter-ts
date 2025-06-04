# tree-sorter-ts

A Go CLI tool that automatically sorts TypeScript object literals, arrays, and function parameters marked with special comments. It uses Tree-sitter for accurate AST parsing while preserving exact formatting, comments, and structure.

## Features

- 🔧 Sorts object properties alphabetically
- 📊 Sorts array elements with customizable sorting keys
- 🏗️ Sorts constructor/function parameters by name (ignoring modifiers)
- 🎯 Only touches objects/arrays/parameters marked with `/** tree-sorter-ts: keep-sorted **/`
- 💬 Preserves all comments (inline and block)
- 🔑 Handles computed property keys like `[EnumName.VALUE]`
- 📁 Processes files in parallel for performance
- ✨ Supports TypeScript and TSX files
- 🔍 Dry-run mode by default (see changes before applying)
- ✅ Check mode for CI/CD pipelines
- 📐 Optional `with-new-line` formatting for extra spacing
- 🚨 Optional `deprecated-at-end` to move `@deprecated` properties to the bottom

## Installation

### Using npm or yarn (Recommended)
```bash
# Using npm
npm install --save-dev tree-sorter-ts

# Using yarn
yarn add -D tree-sorter-ts

# Run the installed binary
npx tree-sorter-ts --help
# or
yarn tree-sorter-ts --help
```

### Using Go
```bash
# Using go install
go install github.com/evanrichards/tree-sorter-ts/cmd/tree-sorter-ts@latest

# Using go run
go run github.com/evanrichards/tree-sorter-ts@latest --help
```

### Building from source
```bash
git clone https://github.com/evanrichards/tree-sorter-ts.git
cd tree-sorter-ts
make build

# Binary will be in ./bin/tree-sorter-ts
./bin/tree-sorter-ts --help
```

## Usage

### Basic usage
```bash
# Dry-run mode (default) - shows what would change
tree-sorter-ts src/

# Write changes to files
tree-sorter-ts --write src/

# Check mode - exits with code 1 if files need sorting
tree-sorter-ts --check src/

# Check mode with detailed output
tree-sorter-ts --check --verbose src/

# Process a single file
tree-sorter-ts --write src/config.ts

# Process only .ts files (not .tsx)
tree-sorter-ts --extensions=".ts" src/
```

### Marking objects for sorting

Add the magic comment before any object literal you want to keep sorted:

```typescript
const config = {
  /** tree-sorter-ts: keep-sorted **/
  zebra: "last",
  alpha: "first", 
  beta: "second",
};
```

After running with `--write`, it becomes:

```typescript
const config = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: "first",
  beta: "second", 
  zebra: "last",
};
```

### Advanced: with-new-line option

For objects that need extra spacing between properties:

```typescript
const config = {
  /** tree-sorter-ts: keep-sorted with-new-line **/
  zebra: "last",
  alpha: "first",
  beta: "second",
};
```

After sorting:

```typescript
const config = {
  /** tree-sorter-ts: keep-sorted with-new-line **/
  alpha: "first",

  beta: "second",

  zebra: "last",
};
```

### Advanced: deprecated-at-end option

Move properties with `@deprecated` annotations to the bottom of the object:

```typescript
const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  activeFeature: true,
  /** @deprecated Use newApiUrl instead */
  oldApiUrl: "https://old.example.com",
  newApiUrl: "https://api.example.com",
  legacyMode: true, // @deprecated Will be removed in v2.0
};
```

After sorting:

```typescript
const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end **/
  activeFeature: true,
  newApiUrl: "https://api.example.com",
  legacyMode: true, // @deprecated Will be removed in v2.0
  /** @deprecated Use newApiUrl instead */
  oldApiUrl: "https://old.example.com",
};
```

You can also combine it with `with-new-line`:

```typescript
const config = {
  /** tree-sorter-ts: keep-sorted deprecated-at-end with-new-line **/
  alpha: "first",

  beta: "second",

  /** @deprecated */
  oldValue: "deprecated",
};
```

### Multiline magic comments

For better readability, you can split the magic comment across multiple lines:

```typescript
const config = {
  /**
   * tree-sorter-ts: keep-sorted
   *   with-new-line
   *   deprecated-at-end
   */
  activeFeature: true,
  beta: "second",
  /** @deprecated */
  oldFeature: false,
};

// Also works without asterisks on each line:
const config2 = {
  /** tree-sorter-ts: keep-sorted
      deprecated-at-end
      with-new-line **/
  gamma: true,
  alpha: "first",
};
```

### Sorting constructor parameters

Constructor parameters (and function parameters) can be sorted alphabetically by parameter name, ignoring access modifiers:

```typescript
class UserService {
  constructor(
    /** tree-sorter-ts: keep-sorted **/
    private readonly userRepository: UserRepository,
    private readonly logger: Logger,
    private readonly cache: CacheService,
    private readonly eventBus: EventBus,
  ) {}
}
```

After running with `--write`, it becomes:

```typescript
class UserService {
  constructor(
    /** tree-sorter-ts: keep-sorted **/
    private readonly cache: CacheService,
    private readonly eventBus: EventBus,
    private readonly logger: Logger,
    private readonly userRepository: UserRepository,
  ) {}
}
```

**Features:**
- Sorts by parameter name, ignoring modifiers like `private`, `readonly`, `public`, `protected`
- Works with regular functions, arrow functions, methods, and constructors
- Supports optional parameters (`param?: Type`)
- Handles destructured parameters (`{ name }: { name: string }`)
- Preserves parameter types and default values
- Supports all sorting options (`with-new-line`, `deprecated-at-end`)

**Examples:**

Mixed access modifiers:
```typescript
class Service {
  constructor(
    /** tree-sorter-ts: keep-sorted **/
    protected readonly zService: ZService,
    public aService: AService,
    private bService: BService,
  ) {}
}
// Sorts to: aService, bService, zService
```

With comments and deprecated parameters:
```typescript
class Service {
  constructor(
    /** tree-sorter-ts: keep-sorted deprecated-at-end **/
    private readonly newService: NewService,
    /** @deprecated Use newService instead */
    private readonly oldService: OldService,
    private readonly activeService: ActiveService,
  ) {}
}
// Sorts to: activeService, newService, then oldService (deprecated last)
```

Arrow functions and regular functions:
```typescript
const handler = (
  /** tree-sorter-ts: keep-sorted **/
  zParam: string,
  aParam: number,
  mParam: boolean,
) => {}
// Sorts to: aParam, mParam, zParam
```

### Sorting arrays

Arrays can also be sorted by placing the magic comment inside the array:

```typescript
const users = [
  /** tree-sorter-ts: keep-sorted key="name" **/
  { name: "Zoe", age: 30 },
  { name: "Alice", age: 25 },
  { name: "Bob", age: 28 },
];
```

After sorting:

```typescript
const users = [
  /** tree-sorter-ts: keep-sorted key="name" **/
  { name: "Alice", age: 25 },
  { name: "Bob", age: 28 },
  { name: "Zoe", age: 30 },
];
```

#### Array sorting options

**Sort by object property:**
```typescript
const items = [
  /** tree-sorter-ts: keep-sorted key="priority" **/
  { name: "Task C", priority: 3 },
  { name: "Task A", priority: 1 },
  { name: "Task B", priority: 2 },
];
```

**Sort by array index (for tuples):**
```typescript
const data = [
  /** tree-sorter-ts: keep-sorted key="1" **/
  ["apple", 5, true],
  ["banana", 2, false],
  ["cherry", 8, true]
];
// Sorts by the second element (index 1): 2, 5, 8
```

**Sort scalar arrays (no key needed):**
```typescript
const numbers = [
  /** tree-sorter-ts: keep-sorted **/
  5, 2, 8, 1, 9
];
// Result: [1, 2, 5, 8, 9]

const words = [
  /** tree-sorter-ts: keep-sorted **/
  "banana", "apple", "cherry"
];
// Result: ["apple", "banana", "cherry"]
```

**Nested property access:**
```typescript
const users = [
  /** tree-sorter-ts: keep-sorted key="profile.firstName" **/
  { profile: { firstName: "Charlie", lastName: "Brown" } },
  { profile: { firstName: "Alice", lastName: "Smith" } },
  { profile: { firstName: "Bob", lastName: "Jones" } }
];
```

**With options (with-new-line and deprecated-at-end):**
```typescript
const features = [
  /** tree-sorter-ts: keep-sorted key="name" deprecated-at-end with-new-line **/
  { name: "Feature C", enabled: true },
  /** @deprecated */
  { name: "Feature A", enabled: false },
  { name: "Feature B", enabled: true },
];
```

Results in:
```typescript
const features = [
  /** tree-sorter-ts: keep-sorted key="name" deprecated-at-end with-new-line **/
  { name: "Feature B", enabled: true },

  { name: "Feature C", enabled: true },

  /** @deprecated */
  { name: "Feature A", enabled: false },
];
```

**Graceful handling of missing keys:**
Elements without the specified key are sorted to the end:
```typescript
const mixed = [
  /** tree-sorter-ts: keep-sorted key="id" **/
  { id: 3, name: "Three" },
  { name: "No ID" },  // Missing 'id' - will sort to end
  { id: 1, name: "One" },
];
// Result: elements with 'id' first (sorted), then elements without 'id'
```

### Sorting by comment content

Both arrays and objects can be sorted by their associated comment content using the `sort-by-comment` option:

```typescript
const userIds = [
  /** tree-sorter-ts: keep-sorted sort-by-comment **/
  "u_8234", // Bob Smith
  "u_9823", // Alice Johnson
  "u_1234", // David Lee
  "u_4521", // Carol White
];
```

After sorting:
```typescript
const userIds = [
  /** tree-sorter-ts: keep-sorted sort-by-comment **/
  "u_9823", // Alice Johnson
  "u_8234", // Bob Smith
  "u_4521", // Carol White
  "u_1234", // David Lee
];
```

**Features:**
- Works with inline comments (`// comment` or `/* comment */`)
- Works with preceding comments (comments on lines before the element)
- Supports multiline comments
- Compatible with `deprecated-at-end` option
- Cannot be used together with `key` option (will show an error)

**Examples:**

Preceding comments:
```typescript
const items = [
  /** tree-sorter-ts: keep-sorted sort-by-comment **/
  // B
  "first",
  /**
   * A
   */
  "second",
  // C
  "third"
];
// Sorts to: second (A), first (B), third (C)
```

Mixed comment positions:
```typescript
const mixed = [
  /** tree-sorter-ts: keep-sorted sort-by-comment **/
  // Delta (before)
  "item1",
  "item2", // Beta (after)
  /**
   * Alpha (before multiline)
   */
  "item3",
  "item4" /* Charlie (after block) */
];
// Sorts by: Alpha, Beta, Charlie, Delta
```

Objects with comment sorting:
```typescript
const config = {
  /** tree-sorter-ts: keep-sorted sort-by-comment **/
  // Production settings
  prodUrl: "https://api.example.com",
  // Development settings  
  devUrl: "http://localhost:3000",
  // Staging settings
  stagingUrl: "https://staging.example.com",
};
// Sorts by: Development, Production, Staging
```

**Known limitation:** Object sorting with inline comments (after property values) currently has a bug where the last property may get a duplicated comment. As a workaround, use preceding comments for objects or use the default property-name sorting.

## Flags

- `--check` - Check if files are sorted (exit 1 if not)
- `--write` - Write changes to files (default: dry-run)
- `--recursive` - Process directories recursively (default: true)
- `--extensions` - File extensions to process (default: ".ts,.tsx")
- `--workers` - Number of parallel workers (default: number of CPUs)
- `--verbose` - Show detailed output (default: false)

## Examples

### CI/CD Integration

The `--check` mode provides clear, CI-friendly output that shows exactly which files need sorting:

```bash
# When files need sorting:
$ tree-sorter-ts --check src/
✗ src/config.ts needs sorting (3 items)
✗ src/services/user.ts needs sorting (1 items)

Processed 15 files
❌ 2 file(s) need sorting
   4 item(s) need to be sorted
Error: some files are not properly sorted
```

```yaml
# GitHub Actions example
- name: Install tree-sorter-ts
  run: npm install --save-dev tree-sorter-ts

- name: Check TypeScript objects are sorted
  run: npx tree-sorter-ts --check src/
```

The tool exits with code 1 and provides specific file paths when sorting is needed, making it easy to identify issues in CI logs.

### Pre-commit Hook
```bash
#!/bin/sh
# .git/hooks/pre-commit
tree-sorter-ts --write $(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(ts|tsx)$')
```

### With Make
```bash
# Run on all TypeScript files
make run

# Check mode
make check

# Run tests
make test
```

## Development

### Project Structure

The codebase follows a modular architecture with clear separation of concerns:

```
tree-sorter-ts/
├── cmd/tree-sorter-ts/         # CLI entry point
├── internal/
│   ├── app/                    # Application coordination
│   ├── fileutil/               # File system utilities
│   ├── processor/              # Main processing logic
│   │   ├── ast.go             # Legacy monolithic processor
│   │   ├── processor.go       # New modular processor
│   │   └── *_test.go          # Comprehensive test suite
│   ├── config/                 # Configuration parsing
│   │   └── sort_config.go     # Magic comment configuration
│   ├── parser/                 # AST parsing utilities
│   │   └── magic_comments.go  # Find sortable structures
│   ├── sorting/                # Core sorting abstractions
│   │   ├── interfaces/        # Core interfaces
│   │   ├── strategies/        # Sorting strategies (plugin-based)
│   │   │   ├── property_name.go   # Default alphabetical sorting
│   │   │   ├── comment_content.go # Sort by comment content
│   │   │   └── array_key.go       # Array key-based sorting
│   │   ├── types/             # Type-specific implementations
│   │   │   ├── arrays/        # Array sorting logic
│   │   │   └── objects/       # Object sorting logic
│   │   └── common/            # Shared utilities
│   └── reconstruction/         # AST reconstruction
│       ├── array_reconstructor.go  # Rebuild sorted arrays
│       └── object_reconstructor.go # Rebuild sorted objects
├── testdata/fixtures/          # Test files
└── main.go                     # Root entry (for backward compatibility)
```

### Architecture Overview

The new modular architecture separates concerns into distinct packages:

1. **Configuration (`config/`)** - Parses and validates magic comment options
2. **Parser (`parser/`)** - Finds structures marked for sorting in the AST
3. **Sorting (`sorting/`)** - Core sorting logic with plugin-based strategies
4. **Reconstruction (`reconstruction/`)** - Rebuilds the AST with sorted content

#### Key Design Patterns

**Strategy Pattern**: Different sorting strategies (property-name, comment-content, array-key) implement a common interface, allowing easy extension:

```go
type SortStrategy interface {
    ExtractKey(item SortableItem, content []byte) (string, error)
    GetName() string
}
```

**Factory Pattern**: Factories create appropriate strategies and reconstructors based on configuration:

```go
strategyFactory := strategies.NewFactory()
strategy, err := strategyFactory.CreateStrategy(config)
```

**Interface-Driven Design**: Core interfaces allow different types (arrays, objects, constructors) to be handled uniformly:

```go
type Sortable interface {
    Extract(node *sitter.Node, content []byte) ([]SortableItem, error)
    Sort(items []SortableItem, strategy SortStrategy, deprecatedAtEnd bool, content []byte) ([]SortableItem, error)
    CheckIfSorted(items []SortableItem, strategy SortStrategy, deprecatedAtEnd bool, content []byte) bool
}
```

#### Processing Flow

1. **Parse** - Tree-sitter parses TypeScript/TSX into an AST
2. **Find** - Locate structures with magic comments
3. **Extract** - Extract sortable items (properties, elements, parameters)
4. **Sort** - Apply the appropriate sorting strategy
5. **Reconstruct** - Rebuild the AST content with sorted items
6. **Write** - Update the file with sorted content

This architecture makes it easy to:
- Add new sorting strategies
- Support new structure types
- Test individual components
- Maintain and debug the codebase

### Building and Testing
```bash
# Build the binary
make build

# Run tests
make test

# Run benchmarks
make bench

# Install locally
make install
```

## TODO

- [ ] **Section sorting** - Support `start-sort` and `end-sort` comments for sorting subsections
  ```typescript
  const config = {
    // Critical settings - do not sort
    apiUrl: "https://api.example.com",
    timeout: 5000,
    
    /** tree-sorter-ts: start-sort **/
    featureFlags: {
      enableAnalytics: true,
      enableChat: false,
      enableNotifications: true,
    },
    permissions: {
      canEdit: true,
      canDelete: false,
      canView: true,
    },
    /** tree-sorter-ts: end-sort **/
    
    // Debug settings - must remain last
    debug: true,
  };
  ```


- [ ] **Class member sorting** - Sort class members including decorators and annotations
  ```typescript
  class APIClient {
    /** tree-sorter-ts: keep-sorted **/
    @deprecated()
    legacyEndpoint: string;
    
    @inject()
    httpClient: HttpClient;
    
    apiKey: string;
    
    baseUrl: string;
    
    @observable()
    isLoading: boolean;
  }
  ```


## License

MIT