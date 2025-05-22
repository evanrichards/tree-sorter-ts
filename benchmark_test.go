package main

import (
	"fmt"
	"os"
	"testing"
)

func BenchmarkSingleFile(b *testing.B) {
	config := Config{
		Check:      true,
		Write:      false,
		Recursive:  true,
		Extensions: []string{".ts", ".tsx"},
		Path:       "test/basic.ts",
		Workers:    1,
	}
	
	for i := 0; i < b.N; i++ {
		processFileSimple("test/basic.ts", config)
	}
}

func BenchmarkParallelProcessing(b *testing.B) {
	config := Config{
		Check:      true,
		Write:      false,
		Recursive:  true,
		Extensions: []string{".ts", ".tsx"},
		Path:       "test/",
		Workers:    4,
	}
	
	files, _ := findFiles("test/", config.Extensions, config.Recursive)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processFilesParallel(files, config)
	}
}

// Create many test files for better benchmarking
func createBenchmarkFiles(count int) error {
	os.MkdirAll("benchmark_test_files", 0755)
	
	template := `const config = {
  /** tree-sorter-ts: keep-sorted **/
  zebra: "value1",
  alpha: "value2", // critical setting
  beta: "value3",
};`
	
	for i := 0; i < count; i++ {
		filename := fmt.Sprintf("benchmark_test_files/test_%d.ts", i)
		err := os.WriteFile(filename, []byte(template), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanupBenchmarkFiles() {
	os.RemoveAll("benchmark_test_files")
}