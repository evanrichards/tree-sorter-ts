package processor

import (
	"testing"
)

func BenchmarkProcessFileAST(b *testing.B) {
	config := Config{
		Check:      true,
		Write:      false,
		Recursive:  true,
		Extensions: []string{".ts", ".tsx"},
		Path:       "../../testdata/fixtures/basic.ts",
		Workers:    1,
	}

	for i := 0; i < b.N; i++ {
		ProcessFileAST("../../testdata/fixtures/basic.ts", config)
	}
}

func BenchmarkProcessFileSimple(b *testing.B) {
	config := Config{
		Check:      true,
		Write:      false,
		Recursive:  true,
		Extensions: []string{".ts", ".tsx"},
		Path:       "../../testdata/fixtures/basic.ts",
		Workers:    1,
	}

	for i := 0; i < b.N; i++ {
		processFileSimple("../../testdata/fixtures/basic.ts", config)
	}
}

func BenchmarkSortObjectAST(b *testing.B) {
	const testContent = `const config = {
  /** tree-sorter-ts: keep-sorted **/
  zebra: "value1",
  alpha: "value2",
  beta: "value3",
};`

	rootNode, content, err := parseTypeScript(testContent)
	if err != nil {
		b.Fatal(err)
	}

	objects := findObjectsWithMagicCommentsAST(rootNode, content)

	if len(objects) == 0 {
		b.Fatal("no objects found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sortObjectAST(objects[0], content)
	}
}