package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFunctionality(t *testing.T) {
	tests := []struct {
		name            string
		initialContent  string
		expectedContent string
		shouldChange    bool
	}{
		{
			name: "basic_write",
			initialContent: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  zebra: "value1",
  alpha: "value2",
  beta: "value3",
};`,
			expectedContent: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: "value2",
  beta: "value3",
  zebra: "value1",
};`,
			shouldChange: true,
		},
		{
			name: "already_sorted_no_write",
			initialContent: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: "value2",
  beta: "value3",
  zebra: "value1",
};`,
			expectedContent: `const config = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: "value2",
  beta: "value3",
  zebra: "value1",
};`,
			shouldChange: false,
		},
		{
			name: "no_magic_comment_no_write",
			initialContent: `const config = {
  zebra: "value1",
  alpha: "value2",
  beta: "value3",
};`,
			expectedContent: `const config = {
  zebra: "value1",
  alpha: "value2",
  beta: "value3",
};`,
			shouldChange: false,
		},
		{
			name: "multiline_values_write",
			initialContent: `const messages = {
  /** tree-sorter-ts: keep-sorted **/
  error: \` + "`" + `This is
a multiline
error\` + "`" + `,
  alert: "Alert!",
};`,
			expectedContent: `const messages = {
  /** tree-sorter-ts: keep-sorted **/
  alert: "Alert!",
  error: \` + "`" + `This is
a multiline
error\` + "`" + `,
};`,
			shouldChange: true,
		},
		{
			name: "preserve_semicolon",
			initialContent: `export const TEST: Record<string, string> = {
  /** tree-sorter-ts: keep-sorted **/
  z: "value",
  a: "value",
};`,
			expectedContent: `export const TEST: Record<string, string> = {
  /** tree-sorter-ts: keep-sorted **/
  a: "value",
  z: "value",
};`,
			shouldChange: true,
		},
	}

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "write-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, tt.name+".ts")
			err := os.WriteFile(testFile, []byte(tt.initialContent), 0o644)
			if err != nil {
				t.Fatalf("Failed to write initial file: %v", err)
			}

			// Test with Write: false first
			config := Config{Write: false}
			result, err := ProcessFileAST(testFile, config)
			if err != nil {
				t.Fatalf("ProcessFileAST failed: %v", err)
			}

			if result.Changed != tt.shouldChange {
				t.Errorf("Expected Changed=%v, got %v", tt.shouldChange, result.Changed)
			}

			// Read file to ensure it wasn't modified
			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			if string(content) != tt.initialContent {
				t.Error("File was modified when Write=false")
			}

			// Now test with Write: true
			if tt.shouldChange {
				config.Write = true
				result2, err := ProcessFileAST(testFile, config)
				if err != nil {
					t.Fatalf("ProcessFileAST with write failed: %v", err)
				}

				if !result2.Changed {
					t.Error("Expected Changed=true when writing")
				}

				// Read the modified content
				modifiedContent, err := os.ReadFile(testFile)
				if err != nil {
					t.Fatalf("Failed to read modified file: %v", err)
				}

				// Normalize for comparison
				expected := strings.TrimSpace(tt.expectedContent)
				actual := strings.TrimSpace(string(modifiedContent))

				if actual != expected {
					t.Errorf("File content mismatch after write.\nExpected:\n%s\n\nGot:\n%s", expected, actual)
				}

				// Test that running again shows no changes needed
				result3, err := ProcessFileAST(testFile, config)
				if err != nil {
					t.Fatalf("ProcessFileAST on sorted file failed: %v", err)
				}

				if result3.Changed {
					t.Error("File still needs changes after sorting")
				}
			}
		})
	}
}

func TestWriteMultipleObjects(t *testing.T) {
	initialContent := `const config1 = {
  /** tree-sorter-ts: keep-sorted **/
  z: 1,
  a: 2,
};

const config2 = {
  x: 1,  // No magic comment
  y: 2,
};

const config3 = {
  /** tree-sorter-ts: keep-sorted **/
  beta: true,
  alpha: false,
};`

	expectedContent := `const config1 = {
  /** tree-sorter-ts: keep-sorted **/
  a: 2,
  z: 1,
};

const config2 = {
  x: 1,  // No magic comment
  y: 2,
};

const config3 = {
  /** tree-sorter-ts: keep-sorted **/
  alpha: false,
  beta: true,
};`

	// Create temp file
	tempDir, err := os.MkdirTemp("", "multi-write-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "multi.ts")
	err = os.WriteFile(testFile, []byte(initialContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Process with write
	config := Config{Write: true}
	result, err := ProcessFileAST(testFile, config)
	if err != nil {
		t.Fatalf("ProcessFileAST failed: %v", err)
	}

	// Check results
	if !result.Changed {
		t.Error("Expected file to be changed")
	}
	if result.ObjectsFound != 2 {
		t.Errorf("Expected 2 objects found, got %d", result.ObjectsFound)
	}
	if result.ObjectsNeedSort != 2 {
		t.Errorf("Expected 2 objects to need sorting, got %d", result.ObjectsNeedSort)
	}

	// Check file content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expected := strings.TrimSpace(expectedContent)
	actual := strings.TrimSpace(string(content))

	if actual != expected {
		t.Errorf("File content mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, actual)
	}
}

func TestWritePermissions(t *testing.T) {
	// Test that file permissions are preserved
	tempDir, err := os.MkdirTemp("", "perm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "perm.ts")
	content := `const x = {
  /** tree-sorter-ts: keep-sorted **/
  b: 2,
  a: 1,
};`

	// Create file with specific permissions
	err = os.WriteFile(testFile, []byte(content), 0o755)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Get original file info
	originalInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat original file: %v", err)
	}

	// Process with write
	config := Config{Write: true}
	_, err = ProcessFileAST(testFile, config)
	if err != nil {
		t.Fatalf("ProcessFileAST failed: %v", err)
	}

	// Check permissions are preserved
	newInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat modified file: %v", err)
	}

	if originalInfo.Mode() != newInfo.Mode() {
		t.Errorf("File permissions changed from %v to %v", originalInfo.Mode(), newInfo.Mode())
	}
}
