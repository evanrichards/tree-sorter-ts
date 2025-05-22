package processor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// SortConfig contains configuration options from the magic comment
type SortConfig struct {
	WithNewLine bool
}

// Config holds the configuration for processing files
type Config struct {
	Check      bool
	Write      bool
	Recursive  bool
	Extensions []string
	Path       string
	Workers    int
}

// ProcessResult contains the result of processing a file
type ProcessResult struct {
	Changed         bool
	ObjectsFound    int
	ObjectsNeedSort int
}

// ProcessFileAST processes a file using full AST analysis
func ProcessFileAST(filePath string, config Config) (ProcessResult, error) {
	result := ProcessResult{}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return result, fmt.Errorf("reading file: %w", err)
	}

	// Early exit if no magic comment found
	if !magicCommentRegex.Match(content) {
		return result, nil
	}

	// Get parser from pool
	parser := parserPool.Get().(*sitter.Parser)
	defer parserPool.Put(parser)

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return result, fmt.Errorf("parsing file: %w", err)
	}

	rootNode := tree.RootNode()

	// Find all objects containing magic comments
	objects := findObjectsWithMagicCommentsAST(rootNode, content)
	if len(objects) == 0 {
		return result, nil
	}

	result.ObjectsFound = len(objects)

	// Process objects from end to beginning
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].object.StartByte() > objects[j].object.StartByte()
	})

	newContent := make([]byte, len(content))
	copy(newContent, content)

	// First pass: count how many need sorting
	for _, obj := range objects {
		_, wasChanged := sortObjectAST(obj, content)
		if wasChanged {
			result.ObjectsNeedSort++
		}
	}

	// Second pass: actually apply changes if needed
	if result.ObjectsNeedSort > 0 {
		result.Changed = true
		for _, obj := range objects {
			sortedContent, wasChanged := sortObjectAST(obj, newContent)
			if wasChanged {
				start := obj.object.StartByte()
				end := obj.object.EndByte()

				before := newContent[:start]
				after := newContent[end:]
				newContent = append(append(before, sortedContent...), after...)
			}
		}
	}

	if result.Changed && config.Write {
		err = os.WriteFile(filePath, newContent, 0o600)
		if err != nil {
			return result, fmt.Errorf("writing file: %w", err)
		}
	}

	return result, nil
}

type objectWithMagicComment struct {
	object       *sitter.Node
	magicComment *sitter.Node
	magicIndex   int // Index of magic comment in children
	sortConfig   SortConfig
}

func parseSortConfig(commentText []byte) SortConfig {
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

			// Parse configuration options
			options := strings.Fields(configPart)
			for _, opt := range options {
				if opt == "with-new-line" {
					config.WithNewLine = true
				}
			}
		}
	}

	return config
}

func findObjectsWithMagicCommentsAST(node *sitter.Node, content []byte) []objectWithMagicComment {
	var results []objectWithMagicComment

	var traverse func(*sitter.Node)
	traverse = func(n *sitter.Node) {
		if n.Type() == "object" {
			// Check children for magic comment
			for i := 0; i < int(n.ChildCount()); i++ {
				child := n.Child(i)
				if child.Type() == "comment" {
					text := content[child.StartByte():child.EndByte()]
					if magicCommentRegex.Match(text) {
						results = append(results, objectWithMagicComment{
							object:       n,
							magicComment: child,
							magicIndex:   i,
							sortConfig:   parseSortConfig(text),
						})
						break
					}
				}
			}
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			traverse(n.Child(i))
		}
	}

	traverse(node)
	return results
}

type astProperty struct {
	keyNode     *sitter.Node
	valueNode   *sitter.Node
	pairNode    *sitter.Node
	key         string
	beforeNodes []*sitter.Node // Comments before this property
	afterNode   *sitter.Node   // Inline comment after property
	hasComma    bool
	commaNode   *sitter.Node
}

func checkFormattingNeeded(obj objectWithMagicComment, properties []*astProperty, content []byte) bool {
	// Check if there's an extra newline between properties
	for i := 0; i < len(properties)-1; i++ {
		prop := properties[i]
		nextProp := properties[i+1]
		
		// Find the end of current property (including comma and inline comment)
		endNode := prop.pairNode
		if prop.afterNode != nil {
			endNode = prop.afterNode
		} else if prop.commaNode != nil {
			endNode = prop.commaNode
		}
		
		// Count newlines between properties
		startByte := endNode.EndByte()
		endByte := nextProp.pairNode.StartByte()
		
		// Handle beforeNodes of next property
		if len(nextProp.beforeNodes) > 0 {
			endByte = nextProp.beforeNodes[0].StartByte()
		}
		
		between := content[startByte:endByte]
		newlineCount := 0
		for _, b := range between {
			if b == '\n' {
				newlineCount++
			}
		}
		
		// If with-new-line is set, we expect 2 newlines between properties (one for the line end, one for spacing)
		// Otherwise, we expect only 1 newline
		expectedNewlines := 1
		if obj.sortConfig.WithNewLine {
			expectedNewlines = 2
		}
		
		if newlineCount != expectedNewlines {
			return true
		}
	}
	
	return false
}

func sortObjectAST(obj objectWithMagicComment, content []byte) ([]byte, bool) {
	// Extract properties after magic comment
	properties := extractPropertiesAST(obj, content)

	if len(properties) <= 1 {
		return nil, false
	}

	// Check if already sorted
	sorted := make([]*astProperty, len(properties))
	copy(sorted, properties)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].key < sorted[j].key
	})

	alreadySorted := true
	for i := range properties {
		if properties[i].key != sorted[i].key {
			alreadySorted = false
			break
		}
	}

	// Even if already sorted, check if formatting needs to change
	needsFormatting := false
	if alreadySorted && len(properties) > 1 {
		// Check if current formatting matches the configuration
		needsFormatting = checkFormattingNeeded(obj, properties, content)
	}

	if alreadySorted && !needsFormatting {
		return nil, false
	}

	// Reconstruct the object with sorted properties
	return reconstructObjectAST(obj, sorted, content), true
}

func extractPropertiesAST(obj objectWithMagicComment, content []byte) []*astProperty {
	var properties []*astProperty
	var pendingComments []*sitter.Node

	// Start after magic comment
	startIdx := obj.magicIndex + 1

	for i := startIdx; i < int(obj.object.ChildCount()); i++ {
		child := obj.object.Child(i)

		switch child.Type() {
		case "comment":
			// Accumulate comments
			pendingComments = append(pendingComments, child)

		case "pair":
			prop := &astProperty{
				pairNode:    child,
				beforeNodes: pendingComments,
			}

			// Extract key and value
			keyNode := child.ChildByFieldName("key")
			valueNode := child.ChildByFieldName("value")

			if keyNode != nil {
				prop.keyNode = keyNode
				prop.key = extractKeyAST(keyNode, content)
			}

			if valueNode != nil {
				prop.valueNode = valueNode
			}

			// Check if followed by comma and/or inline comment
			j := i + 1
			continueLoop := true
			for j < int(obj.object.ChildCount()) && continueLoop {
				next := obj.object.Child(j)
				switch next.Type() {
				case ",":
					prop.hasComma = true
					prop.commaNode = next
					j++
				case "comment":
					// Check if it's on the same line
					if next.StartPoint().Row == child.EndPoint().Row {
						prop.afterNode = next
						j++
					} else {
						continueLoop = false
					}
				default:
					continueLoop = false
				}
			}
			i = j - 1 // Update loop counter to skip processed nodes

			properties = append(properties, prop)
			pendingComments = nil // Reset comments

		case ",":
			// Standalone comma (shouldn't happen if we handle it above)
			continue

		case "}":
			// End of object
			break
		}
	}

	return properties
}

func extractKeyAST(keyNode *sitter.Node, content []byte) string {
	switch keyNode.Type() {
	case "property_identifier":
		return string(content[keyNode.StartByte():keyNode.EndByte()])
	case "string":
		text := string(content[keyNode.StartByte():keyNode.EndByte()])
		return strings.Trim(text, "\"'`")
	case "computed_property_name":
		// For computed properties like [SomeEnum.VALUE], use the full text
		return string(content[keyNode.StartByte():keyNode.EndByte()])
	default:
		return string(content[keyNode.StartByte():keyNode.EndByte()])
	}
}

func reconstructObjectAST(obj objectWithMagicComment, sortedProps []*astProperty, content []byte) []byte {
	var result bytes.Buffer

	// Extract common indentation from first original property after magic comment
	commonIndent := ""
	for i := obj.magicIndex + 1; i < int(obj.object.ChildCount()); i++ {
		child := obj.object.Child(i)
		if child.Type() == "pair" {
			// Found first property
			propStart := child.StartByte()
			lineStart := propStart
			for lineStart > 0 && content[lineStart-1] != '\n' {
				lineStart--
			}
			if lineStart < propStart {
				commonIndent = string(content[lineStart:propStart])
			}
			break
		}
	}

	// Write everything up to and including the magic comment
	// Include the opening brace and everything up to the magic comment
	for i := 0; i <= obj.magicIndex; i++ {
		child := obj.object.Child(i)

		// Write any whitespace before this child
		if i == 0 {
			result.Write(content[obj.object.StartByte():child.StartByte()])
		} else {
			prevChild := obj.object.Child(i - 1)
			result.Write(content[prevChild.EndByte():child.StartByte()])
		}

		// Write the child itself
		result.Write(content[child.StartByte():child.EndByte()])
	}

	// Write newline after magic comment, but not the indentation
	// (indentation will be added when writing properties)
	if obj.magicIndex+1 < int(obj.object.ChildCount()) {
		result.WriteByte('\n')
	}

	// Write sorted properties
	for i, prop := range sortedProps {
		// Write any comments before this property
		for _, commentNode := range prop.beforeNodes {
			// Find whitespace before comment
			if len(result.Bytes()) > 0 {
				commentStart := commentNode.StartByte()
				// Try to preserve indentation
				lineStart := commentStart
				for lineStart > 0 && content[lineStart-1] != '\n' {
					lineStart--
				}
				if lineStart < commentStart {
					result.Write(content[lineStart:commentStart])
				}
			}
			result.Write(content[commentNode.StartByte():commentNode.EndByte()])
			result.WriteByte('\n')
		}

		// Use common indentation for all properties
		result.WriteString(commonIndent)

		// Write the property itself (preserving all formatting)
		result.Write(content[prop.pairNode.StartByte():prop.pairNode.EndByte()])

		// Handle comma
		if i < len(sortedProps)-1 {
			if prop.hasComma && prop.commaNode != nil {
				// Use original comma
				result.Write(content[prop.commaNode.StartByte():prop.commaNode.EndByte()])
			} else {
				// Add comma if missing
				result.WriteByte(',')
			}
		} else {
			// Last property - check if original had trailing comma
			originalLastProp := findOriginalLastProperty(obj, content)
			if originalLastProp != nil && originalLastProp.hasComma {
				result.WriteByte(',')
			}
		}

		// Write inline comment if present
		if prop.afterNode != nil {
			result.WriteByte(' ')
			result.Write(content[prop.afterNode.StartByte():prop.afterNode.EndByte()])
		}

		// Add newline if not last or if there's more content
		if i < len(sortedProps)-1 {
			result.WriteByte('\n')
			// Add extra newline if with-new-line option is set
			if obj.sortConfig.WithNewLine {
				result.WriteByte('\n')
			}
		}
	}

	// Write closing brace and any trailing content
	closingBraceIdx := -1
	for i := int(obj.object.ChildCount()) - 1; i >= 0; i-- {
		child := obj.object.Child(i)
		if child.Type() == "}" {
			closingBraceIdx = i
			break
		}
	}

	if closingBraceIdx > 0 {
		closingBrace := obj.object.Child(closingBraceIdx)

		// Write any whitespace/newlines before closing brace
		// Try to preserve the original spacing
		originalSpacing := findOriginalClosingSpacing(obj, content)
		if originalSpacing != "" {
			result.WriteString(originalSpacing)
		} else {
			result.WriteByte('\n')
		}

		result.Write(content[closingBrace.StartByte():closingBrace.EndByte()])
	}

	return result.Bytes()
}

func findOriginalLastProperty(obj objectWithMagicComment, _ []byte) *astProperty {
	var lastProp *astProperty

	for i := obj.magicIndex + 1; i < int(obj.object.ChildCount()); i++ {
		child := obj.object.Child(i)
		if child.Type() == "pair" {
			lastProp = &astProperty{
				pairNode: child,
			}
			// Check if followed by comma
			if i+1 < int(obj.object.ChildCount()) {
				next := obj.object.Child(i + 1)
				if next.Type() == "," {
					lastProp.hasComma = true
					lastProp.commaNode = next
				}
			}
		}
	}

	return lastProp
}

func findOriginalClosingSpacing(obj objectWithMagicComment, content []byte) string {
	// Find the last property or comma
	lastContentEnd := obj.magicComment.EndByte()

	for i := int(obj.object.ChildCount()) - 1; i > obj.magicIndex; i-- {
		child := obj.object.Child(i)
		if child.Type() == "pair" || child.Type() == "," {
			lastContentEnd = child.EndByte()
			break
		}
	}

	// Find closing brace
	for i := int(obj.object.ChildCount()) - 1; i >= 0; i-- {
		child := obj.object.Child(i)
		if child.Type() == "}" {
			return string(content[lastContentEnd:child.StartByte()])
		}
	}

	return "\n"
}
