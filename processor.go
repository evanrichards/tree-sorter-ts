package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

var (
	magicCommentRegex = regexp.MustCompile(`/\*\*?\s*tree-sorter-ts:\s*keep-sorted\s*\*+/`)
	magicComment      = "tree-sorter-ts: keep-sorted"
	
	// Parser pool to avoid recreating parsers
	parserPool = sync.Pool{
		New: func() interface{} {
			parser := sitter.NewParser()
			parser.SetLanguage(typescript.GetLanguage())
			return parser
		},
	}
)

func processFileSimple(filePath string, config Config) (bool, error) {
	// Quick check: if file doesn't contain magic comment, skip parsing
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("reading file: %w", err)
	}
	
	// Early exit if no magic comment found
	if !magicCommentRegex.Match(content) {
		return false, nil
	}

	// Get parser from pool
	parser := parserPool.Get().(*sitter.Parser)
	defer parserPool.Put(parser)

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return false, fmt.Errorf("parsing file: %w", err)
	}

	rootNode := tree.RootNode()

	// Find all objects that contain magic comments
	objectsToSort := findObjectsContainingMagicComment(rootNode, content)
	if len(objectsToSort) == 0 {
		return false, nil
	}

	// Sort from end to beginning
	sort.Slice(objectsToSort, func(i, j int) bool {
		return objectsToSort[i].StartByte() > objectsToSort[j].StartByte()
	})

	changed := false
	newContent := string(content)

	for _, obj := range objectsToSort {
		sortedObj, wasChanged := sortObjectSimple(obj, []byte(newContent))
		if wasChanged {
			changed = true
			// Replace the object in the content
			before := newContent[:obj.StartByte()]
			after := newContent[obj.EndByte():]
			newContent = before + sortedObj + after
		}
	}

	if changed && config.Write {
		err = os.WriteFile(filePath, []byte(newContent), 0644)
		if err != nil {
			return false, fmt.Errorf("writing file: %w", err)
		}
	}

	return changed, nil
}

func findObjectsContainingMagicComment(node *sitter.Node, content []byte) []*sitter.Node {
	var objects []*sitter.Node

	var traverse func(*sitter.Node)
	traverse = func(n *sitter.Node) {
		if n.Type() == "object" {
			objContent := content[n.StartByte():n.EndByte()]
			if magicCommentRegex.Match(objContent) {
				objects = append(objects, n)
			}
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			traverse(n.Child(i))
		}
	}

	traverse(node)
	return objects
}

func sortObjectSimple(objectNode *sitter.Node, content []byte) (string, bool) {
	objContent := string(content[objectNode.StartByte():objectNode.EndByte()])

	// Find the magic comment position
	magicMatch := magicCommentRegex.FindStringIndex(objContent)
	if magicMatch == nil {
		return "", false
	}

	// Split the object into parts
	beforeMagic := objContent[:magicMatch[0]]
	magicComment := objContent[magicMatch[0]:magicMatch[1]]
	afterMagic := objContent[magicMatch[1]:]

	// Extract properties after the magic comment
	lines := strings.Split(afterMagic, "\n")
	var properties []PropertyLine
	var afterProperties string

	foundClosingBrace := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Check for closing brace
		if strings.HasPrefix(trimmed, "}") {
			afterProperties = strings.Join(lines[i:], "\n")
			foundClosingBrace = true
			break
		}

		// This is a property line
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") {
			properties = append(properties, PropertyLine{
				Content: line,
				Key:     extractPropertyKey(line),
			})
		}
	}

	if len(properties) <= 1 || !foundClosingBrace {
		return "", false
	}

	// Check if already sorted
	sortedProps := make([]PropertyLine, len(properties))
	copy(sortedProps, properties)
	sort.Slice(sortedProps, func(i, j int) bool {
		return sortedProps[i].Key < sortedProps[j].Key
	})

	alreadySorted := true
	for i := range properties {
		if properties[i].Key != sortedProps[i].Key {
			alreadySorted = false
			break
		}
	}

	if alreadySorted {
		return "", false
	}

	// Build the sorted object
	var result strings.Builder
	result.WriteString(beforeMagic)
	result.WriteString(magicComment)
	result.WriteString("\n")

	for i, prop := range sortedProps {
		result.WriteString(prop.Content)
		if i < len(sortedProps)-1 {
			// Make sure there's a comma
			trimmed := strings.TrimRight(prop.Content, " \t")
			if !strings.HasSuffix(trimmed, ",") {
				result.WriteString(",")
			}
		}
		result.WriteString("\n")
	}

	result.WriteString(afterProperties)

	return result.String(), true
}

type PropertyLine struct {
	Content string
	Key     string
}

func extractPropertyKey(line string) string {
	trimmed := strings.TrimSpace(line)

	// Handle computed property names [something]: value
	if strings.HasPrefix(trimmed, "[") {
		endIdx := strings.Index(trimmed, "]")
		if endIdx > 0 {
			return trimmed[:endIdx+1]
		}
	}

	// Handle regular property names
	// Look for the colon
	colonIdx := strings.Index(trimmed, ":")
	if colonIdx > 0 {
		key := strings.TrimSpace(trimmed[:colonIdx])
		// Remove quotes if present
		key = strings.Trim(key, "\"'")
		return key
	}

	return trimmed
}
