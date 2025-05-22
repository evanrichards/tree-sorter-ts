package main

import (
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
