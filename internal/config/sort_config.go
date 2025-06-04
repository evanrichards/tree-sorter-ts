package config

import (
	"fmt"
	"strings"
)

// SortConfig contains configuration options from the magic comment
type SortConfig struct {
	WithNewLine     bool
	DeprecatedAtEnd bool
	Key             string // For array sorting
	SortByComment   bool   // Sort by comment content
	HasError        bool   // Indicates a validation error
}

// ParseSortConfig extracts configuration from a magic comment
func ParseSortConfig(commentText []byte) SortConfig {
	config := SortConfig{}

	// Extract configuration from magic comment
	text := string(commentText)
	// Look for the pattern inside the comment
	if strings.Contains(text, "tree-sorter-ts:") && strings.Contains(text, "keep-sorted") {
		// Find the part after "keep-sorted"
		parts := strings.Split(text, "keep-sorted")
		if len(parts) > 1 {
			// Extract the configuration part before the closing */
			configPart := parts[1]
			endIdx := strings.Index(configPart, "*/")
			if endIdx > 0 {
				configPart = configPart[:endIdx]
			}

			// Remove leading asterisks from each line (for multiline comments)
			lines := strings.Split(configPart, "\n")
			cleanedLines := make([]string, 0, len(lines))
			for _, line := range lines {
				// Trim spaces and asterisks from the beginning of each line
				line = strings.TrimSpace(line)
				line = strings.TrimPrefix(line, "*")
				line = strings.TrimSpace(line)
				if line != "" {
					cleanedLines = append(cleanedLines, line)
				}
			}
			configPart = strings.Join(cleanedLines, " ")

			// Parse configuration options
			options := strings.Fields(configPart)
			for i, opt := range options {
				switch opt {
				case "with-new-line":
					config.WithNewLine = true
				case "deprecated-at-end":
					config.DeprecatedAtEnd = true
				case "sort-by-comment":
					config.SortByComment = true
				default:
					// Check for key="value" pattern
					if strings.HasPrefix(opt, "key=") {
						// Extract the quoted value
						keyPart := opt[4:]
						keyPart = strings.Trim(keyPart, "\"'")
						config.Key = keyPart
					} else if opt == "key=" && i+1 < len(options) {
						// Handle case where key= and value are separate
						config.Key = strings.Trim(options[i+1], "\"'")
					}
				}
			}
		}
	}

	return config
}

// Validate checks for configuration conflicts and returns an error if found
func (c *SortConfig) Validate() error {
	// Validation: cannot use both key and sort-by-comment
	if c.Key != "" && c.SortByComment {
		c.HasError = true
		return fmt.Errorf("invalid configuration: cannot use both 'key' and 'sort-by-comment' options together")
	}
	return nil
}

// GetSortingMode returns a string describing the sorting mode for debugging
func (c *SortConfig) GetSortingMode() string {
	if c.SortByComment {
		return "sort-by-comment"
	}
	if c.Key != "" {
		return fmt.Sprintf("key=%q", c.Key)
	}
	return "property-name"
}