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

## Flags

- `--check` - Check if files are sorted (exit 1 if not)
- `--write` - Write changes to files (default: dry-run)
- `--recursive` - Process directories recursively (default: true)
- `--extensions` - File extensions to process (default: ".ts,.tsx")
- `--workers` - Number of parallel workers (default: number of CPUs)
- `--verbose` - Show detailed output (default: false)

## Examples

### CI/CD Integration
```yaml
# GitHub Actions example
- name: Install tree-sorter-ts
  run: npm install --save-dev tree-sorter-ts

- name: Check TypeScript objects are sorted
  run: npx tree-sorter-ts --check src/
```

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
```
tree-sorter-ts/
├── cmd/tree-sorter-ts/     # CLI entry point
├── internal/
│   ├── app/                # Application logic
│   ├── processor/          # Core sorting logic
│   └── fileutil/           # File utilities
├── testdata/fixtures/      # Test files
└── main.go                 # Root entry (for backward compatibility)
```

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

- [ ] **Sort by comment contents** - Support sorting elements based on inline or preceding comment contents
  ```typescript
  const userIds = [
    /** tree-sorter-ts: keep-sorted sort-by-comment **/
    "u_8234", // Bob Smith
    "u_9823", // Alice Johnson
    "u_1234", // David Lee
    "u_4521", // Carol White
  ];
  
  // Would sort to:
  const userIds = [
    /** tree-sorter-ts: keep-sorted sort-by-comment **/
    "u_9823", // Alice Johnson
    "u_8234", // Bob Smith
    "u_4521", // Carol White
    "u_1234", // David Lee
  ];
  ```

## License

MIT