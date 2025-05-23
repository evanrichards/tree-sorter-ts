# tree-sorter-ts

A Go CLI tool that automatically sorts TypeScript object literals marked with special comments. It uses Tree-sitter for accurate AST parsing while preserving exact formatting, comments, and structure.

## Features

- 🔧 Sorts object properties alphabetically
- 🎯 Only touches objects marked with `/** tree-sorter-ts: keep-sorted **/`
- 💬 Preserves all comments (inline and block)
- 🔑 Handles computed property keys like `[EnumName.VALUE]`
- 📁 Processes files in parallel for performance
- ✨ Supports TypeScript and TSX files
- 🔍 Dry-run mode by default (see changes before applying)
- ✅ Check mode for CI/CD pipelines
- 📐 Optional `with-new-line` formatting for extra spacing
- 🚨 Optional `deprecated-at-end` to move `@deprecated` properties to the bottom

## Installation

### Using `go install`
```bash
# Install the latest version
go install github.com/evanrichards/tree-sorter-ts/cmd/tree-sorter-ts@latest

# Run the installed binary
tree-sorter-ts --help
```

### Using `go run`
```bash
# Run directly without installation
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

## Flags

- `--check` - Check if files are sorted (exit 1 if not)
- `--write` - Write changes to files (default: dry-run)
- `--recursive` - Process directories recursively (default: true)
- `--extensions` - File extensions to process (default: ".ts,.tsx")
- `--workers` - Number of parallel workers (default: number of CPUs)

## Examples

### CI/CD Integration
```yaml
# GitHub Actions example
- name: Check TypeScript objects are sorted
  run: |
    go run github.com/evanrichards/tree-sorter-ts@latest --check src/
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

- [ ] **Array sorting support** - Sort arrays with a `key` flag to sort objects by a specific path
  ```typescript
  const users = [
    /** tree-sorter-ts: keep-sorted key="name" **/
    { name: "Zoe", age: 30 },
    { name: "Alice", age: 25 },
    { name: "Bob", age: 28 },
  ];
  ```

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

- [ ] **Constructor argument sorting** - Support sorting constructor arguments in class definitions
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

## License

MIT