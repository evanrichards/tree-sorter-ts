package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

type Config struct {
	Check      bool
	Write      bool
	Recursive  bool
	Extensions []string
	Path       string
	Workers    int
}

func main() {
	config := parseFlags()

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() Config {
	var config Config
	var extensions string

	flag.BoolVar(&config.Check, "check", false, "Check if files are sorted (exit 1 if not)")
	flag.BoolVar(&config.Write, "write", false, "Write changes to files (default: dry-run)")
	flag.BoolVar(&config.Recursive, "recursive", true, "Process directories recursively")
	flag.StringVar(&extensions, "extensions", ".ts,.tsx", "File extensions to process")
	flag.IntVar(&config.Workers, "workers", 0, "Number of parallel workers (0 = number of CPUs)")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <path>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	config.Path = args[0]
	config.Extensions = strings.Split(extensions, ",")

	return config
}

func run(config Config) error {
	fileInfo, err := os.Stat(config.Path)
	if err != nil {
		return fmt.Errorf("cannot access path %s: %w", config.Path, err)
	}

	var files []string

	if fileInfo.IsDir() {
		files, err = findFiles(config.Path, config.Extensions, config.Recursive)
		if err != nil {
			return fmt.Errorf("error finding files: %w", err)
		}
	} else {
		if hasValidExtension(config.Path, config.Extensions) {
			files = []string{config.Path}
		} else {
			return fmt.Errorf("file %s does not have a valid extension", config.Path)
		}
	}

	if len(files) == 0 {
		fmt.Println("No TypeScript files found")
		return nil
	}

	fmt.Printf("Found %d TypeScript file(s)\n", len(files))

	// Process files in parallel
	needsSorting, err := processFilesParallel(files, config)
	if err != nil {
		return err
	}

	if config.Check && needsSorting {
		return fmt.Errorf("files need sorting")
	}

	return nil
}

type fileResult struct {
	file            string
	changed         bool
	err             error
	objectsFound    int
	objectsNeedSort int
}

type summary struct {
	totalFiles      int
	filesNoChanges  int
	filesNeedSort   int
	errorFiles      int
	totalObjects    int
	objectsNeedSort int
}

func processFilesParallel(files []string, config Config) (bool, error) {
	// Determine worker count
	workerCount := config.Workers
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
		if workerCount > 8 {
			workerCount = 8
		}
	}

	// Create channels for work distribution
	fileChan := make(chan string, len(files))
	resultChan := make(chan fileResult, len(files))

	// Create wait group for workers
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				processResult, err := ProcessFileAST(file, config)
				resultChan <- fileResult{
					file:            file,
					changed:         processResult.Changed,
					err:             err,
					objectsFound:    processResult.ObjectsFound,
					objectsNeedSort: processResult.ObjectsNeedSort,
				}
			}
		}()
	}

	// Send files to workers
	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var needsSorting atomic.Bool
	var hasError atomic.Bool
	var errorMutex sync.Mutex
	var errors []error

	stats := summary{
		totalFiles: len(files),
	}

	for result := range resultChan {
		if result.err != nil {
			hasError.Store(true)
			errorMutex.Lock()
			errors = append(errors, fmt.Errorf("%s: %w", result.file, result.err))
			errorMutex.Unlock()
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", result.file, result.err)
			stats.errorFiles++
			continue
		}

		// Update object counts
		stats.totalObjects += result.objectsFound
		stats.objectsNeedSort += result.objectsNeedSort

		if result.changed {
			needsSorting.Store(true)
			stats.filesNeedSort++
			if config.Write {
				fmt.Printf("✓ Sorted %s (%d objects)\n", result.file, result.objectsNeedSort)
			} else {
				fmt.Printf("Would sort %s (%d objects)\n", result.file, result.objectsNeedSort)
			}
		} else {
			stats.filesNoChanges++
			if config.Check && result.objectsFound > 0 {
				fmt.Printf("✓ No changes needed %s (%d objects already sorted)\n", result.file, result.objectsFound)
			} else if config.Check {
				fmt.Printf("✓ No changes needed %s\n", result.file)
			}
		}
	}

	// Print summary for all modes when processing multiple files
	if stats.totalFiles > 1 {
		fmt.Println("\n─────────────────────────────────────")
		fmt.Printf("Total files:    %d\n", stats.totalFiles)

		if config.Check {
			fmt.Printf("No changes:     %d\n", stats.filesNoChanges)
			if stats.filesNeedSort > 0 {
				fmt.Printf("Need sorting:   %d ❌\n", stats.filesNeedSort)
			}
		} else if config.Write {
			fmt.Printf("Sorted:         %d\n", stats.filesNeedSort)
			fmt.Printf("No changes:     %d\n", stats.filesNoChanges)
		} else {
			// Dry-run mode
			fmt.Printf("Would sort:     %d\n", stats.filesNeedSort)
			fmt.Printf("No changes:     %d\n", stats.filesNoChanges)
		}

		if stats.errorFiles > 0 {
			fmt.Printf("Errors:         %d\n", stats.errorFiles)
		}

		// Object-level summary
		if stats.totalObjects > 0 {
			fmt.Printf("\nkeep-sorted objects:\n")
			fmt.Printf("Total found:    %d\n", stats.totalObjects)

			objectsSorted := stats.totalObjects - stats.objectsNeedSort
			if config.Write {
				// After writing, all objects are sorted
				objectsSorted = stats.totalObjects
			}

			if config.Check {
				fmt.Printf("Sorted:         %d\n", objectsSorted)
				if stats.objectsNeedSort > 0 {
					fmt.Printf("Need sorting:   %d ❌\n", stats.objectsNeedSort)
				}
			} else if config.Write {
				fmt.Printf("Sorted:         %d (was %d)\n", objectsSorted, objectsSorted-stats.objectsNeedSort)
			} else {
				// Dry-run mode
				fmt.Printf("Sorted:         %d\n", objectsSorted)
				fmt.Printf("Would sort:     %d\n", stats.objectsNeedSort)
			}

			// Calculate object-level compliance
			percentage := float64(objectsSorted) / float64(stats.totalObjects) * 100
			beforePercentage := float64(stats.totalObjects-stats.objectsNeedSort) / float64(stats.totalObjects) * 100

			if config.Check {
				fmt.Printf("\nCompliance:     %.1f%%\n", percentage)
			} else if config.Write {
				fmt.Printf("\nCompliance:     %.1f%% → %.1f%%\n", beforePercentage, percentage)
			} else {
				fmt.Printf("\nCompliance:     %.1f%% (would be %.1f%%)\n", beforePercentage, 100.0)
			}
		}
	}

	if hasError.Load() && len(errors) > 0 {
		return needsSorting.Load(), errors[0]
	}

	return needsSorting.Load(), nil
}

func findFiles(root string, extensions []string, recursive bool) ([]string, error) {
	var files []string

	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if !recursive && path != root {
				return filepath.SkipDir
			}
			return nil
		}

		if hasValidExtension(path, extensions) {
			files = append(files, path)
		}

		return nil
	}

	err := filepath.WalkDir(root, walkFn)
	return files, err
}

func hasValidExtension(path string, extensions []string) bool {
	for _, ext := range extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
