package app

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/evanrichards/tree-sorter-ts/internal/fileutil"
	"github.com/evanrichards/tree-sorter-ts/internal/processor"
)

func Run() {
	config := parseFlags()

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() processor.Config {
	var config processor.Config
	var extensions string

	flag.BoolVar(&config.Check, "check", false, "Check if files are sorted (exit 1 if not)")
	flag.BoolVar(&config.Write, "write", false, "Write changes to files (default: dry-run)")
	flag.BoolVar(&config.Recursive, "recursive", true, "Process directories recursively")
	flag.StringVar(&extensions, "extensions", ".ts,.tsx", "File extensions to process")
	flag.IntVar(&config.Workers, "workers", 0, "Number of parallel workers (0 = number of CPUs)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Show detailed output")

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

func run(config processor.Config) error {
	fileInfo, err := os.Stat(config.Path)
	if err != nil {
		return fmt.Errorf("cannot access path %s: %w", config.Path, err)
	}

	var files []string

	if fileInfo.IsDir() {
		files, err = fileutil.FindFiles(config.Path, config.Extensions, config.Recursive)
		if err != nil {
			return fmt.Errorf("error finding files: %w", err)
		}
	} else {
		if fileutil.HasValidExtension(config.Path, config.Extensions) {
			files = []string{config.Path}
		} else {
			return fmt.Errorf("file %s does not have a valid extension", config.Path)
		}
	}

	if len(files) == 0 {
		if config.Verbose {
			fmt.Println("No TypeScript files found")
		}
		return nil
	}

	if config.Verbose {
		fmt.Printf("Found %d TypeScript file(s)\n", len(files))
	}

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

type stats struct {
	totalFiles      int
	filesNeedSort   int
	filesNoChanges  int
	errorFiles      int
	totalObjects    int
	objectsNeedSort int
}

func processFilesParallel(files []string, config processor.Config) (bool, error) {
	// Set up worker pool
	workerCount := config.Workers
	if workerCount == 0 {
		workerCount = runtime.NumCPU()
	}

	// Channels for work distribution
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
				processResult, err := processor.ProcessFileAST(file, config)
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
	wg.Wait()
	close(resultChan)

	// Collect results
	fileStats := stats{
		totalFiles: len(files),
	}
	var needsSorting atomic.Bool
	var errors []error

	for result := range resultChan {
		if result.err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", result.file, result.err))
			fileStats.errorFiles++
			continue
		}

		fileStats.totalObjects += result.objectsFound
		fileStats.objectsNeedSort += result.objectsNeedSort

		if result.changed {
			needsSorting.Store(true)
			fileStats.filesNeedSort++

			if config.Verbose {
				switch {
				case config.Write:
					fmt.Printf("✓ Sorted %s (%d objects)\n", result.file, result.objectsNeedSort)
				case config.Check:
					fmt.Printf("✗ Needs sorting: %s (%d objects need sorting)\n", result.file, result.objectsNeedSort)
				default:
					// Dry-run mode
					fmt.Printf("Would sort %s (%d objects need sorting)\n", result.file, result.objectsNeedSort)
				}
			}
		} else {
			fileStats.filesNoChanges++
			if config.Verbose {
				// Only print in check mode if objects were found
				if config.Check && result.objectsFound > 0 {
					fmt.Printf("✓ No changes needed %s (%d objects already sorted)\n", result.file, result.objectsFound)
				} else if config.Check {
					fmt.Printf("✓ No changes needed %s\n", result.file)
				}
			}
		}
	}

	// Print summary for all modes when processing multiple files
	if config.Verbose && fileStats.totalFiles > 1 {
		fmt.Println("\n─────────────────────────────────────")
		fmt.Printf("Total files:    %d\n", fileStats.totalFiles)

		switch {
		case config.Check:
			fmt.Printf("No changes:     %d\n", fileStats.filesNoChanges)
			if fileStats.filesNeedSort > 0 {
				fmt.Printf("Need sorting:   %d ❌\n", fileStats.filesNeedSort)
			}
		case config.Write:
			fmt.Printf("Sorted:         %d\n", fileStats.filesNeedSort)
			fmt.Printf("No changes:     %d\n", fileStats.filesNoChanges)
		default:
			// Dry-run mode
			fmt.Printf("Would sort:     %d\n", fileStats.filesNeedSort)
			fmt.Printf("No changes:     %d\n", fileStats.filesNoChanges)
		}

		if fileStats.errorFiles > 0 {
			fmt.Printf("Errors:         %d\n", fileStats.errorFiles)
		}

		// Object-level summary
		if fileStats.totalObjects > 0 {
			fmt.Printf("\nkeep-sorted objects:\n")
			fmt.Printf("Total found:    %d\n", fileStats.totalObjects)

			objectsSorted := fileStats.totalObjects - fileStats.objectsNeedSort
			if config.Write {
				// After writing, all objects are sorted
				objectsSorted = fileStats.totalObjects
			}

			switch {
			case config.Check:
				fmt.Printf("Sorted:         %d\n", objectsSorted)
				if fileStats.objectsNeedSort > 0 {
					fmt.Printf("Need sorting:   %d ❌\n", fileStats.objectsNeedSort)
				}
			case config.Write:
				fmt.Printf("Sorted:         %d (was %d)\n", objectsSorted, objectsSorted-fileStats.objectsNeedSort)
			default:
				// Dry-run mode
				fmt.Printf("Would sort:     %d\n", fileStats.objectsNeedSort)
				fmt.Printf("Already sorted: %d\n", objectsSorted)
			}
		}
	}

	// Handle errors
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		// Only return first error to maintain backwards compatibility
		return needsSorting.Load(), errors[0]
	}

	return needsSorting.Load(), nil
}
