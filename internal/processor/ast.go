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
	WithNewLine     bool
	DeprecatedAtEnd bool
	Key             string // For array sorting
}

// Config holds the configuration for processing files
type Config struct {
	Check      bool
	Write      bool
	Recursive  bool
	Extensions []string
	Path       string
	Workers    int
	Verbose    bool
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

	// Find all objects, arrays, and constructors containing magic comments
	objects := findObjectsWithMagicCommentsAST(rootNode, content)
	arrays := findArraysWithMagicCommentsAST(rootNode, content)
	constructors := findConstructorsWithMagicCommentsAST(rootNode, content)

	if len(objects) == 0 && len(arrays) == 0 && len(constructors) == 0 {
		return result, nil
	}

	result.ObjectsFound = len(objects) + len(arrays) + len(constructors)

	// Create a combined list of sortable items
	type sortableItem struct {
		startByte     uint32
		endByte       uint32
		isArray       bool
		isConstructor bool
		objIndex      int
		arrIndex      int
		constrIndex   int
	}

	// Pre-allocate items slice
	items := make([]sortableItem, 0, len(objects)+len(arrays)+len(constructors))
	for i, obj := range objects {
		items = append(items, sortableItem{
			startByte: obj.object.StartByte(),
			endByte:   obj.object.EndByte(),
			isArray:   false,
			objIndex:  i,
		})
	}
	for i, arr := range arrays {
		items = append(items, sortableItem{
			startByte: arr.array.StartByte(),
			endByte:   arr.array.EndByte(),
			isArray:   true,
			arrIndex:  i,
		})
	}
	for i, constr := range constructors {
		items = append(items, sortableItem{
			startByte:     constr.formalParams.StartByte(),
			endByte:       constr.formalParams.EndByte(),
			isConstructor: true,
			constrIndex:   i,
		})
	}

	// Process items from end to beginning
	sort.Slice(items, func(i, j int) bool {
		return items[i].startByte > items[j].startByte
	})

	newContent := make([]byte, len(content))
	copy(newContent, content)

	// First pass: count how many need sorting
	for _, item := range items {
		if item.isArray {
			_, wasChanged := sortArrayAST(arrays[item.arrIndex], content)
			if wasChanged {
				result.ObjectsNeedSort++
			}
		} else if item.isConstructor {
			_, wasChanged := sortConstructorAST(constructors[item.constrIndex], content)
			if wasChanged {
				result.ObjectsNeedSort++
			}
		} else {
			_, wasChanged := sortObjectAST(objects[item.objIndex], content)
			if wasChanged {
				result.ObjectsNeedSort++
			}
		}
	}

	// Second pass: actually apply changes if needed
	if result.ObjectsNeedSort > 0 {
		result.Changed = true
		for _, item := range items {
			var sortedContent []byte
			var wasChanged bool

			if item.isArray {
				sortedContent, wasChanged = sortArrayAST(arrays[item.arrIndex], content)
			} else if item.isConstructor {
				sortedContent, wasChanged = sortConstructorAST(constructors[item.constrIndex], content)
			} else {
				sortedContent, wasChanged = sortObjectAST(objects[item.objIndex], content)
			}

			if wasChanged {
				start := item.startByte
				end := item.endByte

				// Create a new slice to avoid corruption when content size changes
				result := make([]byte, 0, len(newContent)-int(end-start)+len(sortedContent))
				result = append(result, newContent[:start]...)
				result = append(result, sortedContent...)
				result = append(result, newContent[end:]...)
				newContent = result
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
	keyNode      *sitter.Node
	valueNode    *sitter.Node
	pairNode     *sitter.Node
	key          string
	beforeNodes  []*sitter.Node // Comments before this property
	afterNode    *sitter.Node   // Inline comment after property
	hasComma     bool
	commaNode    *sitter.Node
	isDeprecated bool // Whether this property has @deprecated annotation
}

func hasDeprecatedAnnotation(nodes []*sitter.Node, content []byte) bool {
	for _, node := range nodes {
		text := string(content[node.StartByte():node.EndByte()])
		if strings.Contains(text, "@deprecated") {
			return true
		}
	}
	return false
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

	// Sort properties, considering deprecated-at-end flag
	if obj.sortConfig.DeprecatedAtEnd {
		sort.Slice(sorted, func(i, j int) bool {
			// If one is deprecated and the other isn't, put non-deprecated first
			if sorted[i].isDeprecated != sorted[j].isDeprecated {
				return !sorted[i].isDeprecated
			}
			// Otherwise sort alphabetically
			return sorted[i].key < sorted[j].key
		})
	} else {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].key < sorted[j].key
		})
	}

	alreadySorted := true
	for i := range properties {
		if properties[i].key != sorted[i].key {
			alreadySorted = false
			break
		}
		// For deprecated-at-end, also check if deprecated properties are in the right place
		if obj.sortConfig.DeprecatedAtEnd && properties[i].isDeprecated != sorted[i].isDeprecated {
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

			// Check if this property has @deprecated annotation
			prop.isDeprecated = hasDeprecatedAnnotation(pendingComments, content)

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

			// Also check inline comment for @deprecated
			if !prop.isDeprecated && prop.afterNode != nil {
				text := string(content[prop.afterNode.StartByte():prop.afterNode.EndByte()])
				if strings.Contains(text, "@deprecated") {
					prop.isDeprecated = true
				}
			}

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

// Array sorting functionality

type arrayWithMagicComment struct {
	array        *sitter.Node
	magicComment *sitter.Node
	magicIndex   int // Index of magic comment in children
	sortConfig   SortConfig
}

func findArraysWithMagicCommentsAST(node *sitter.Node, content []byte) []arrayWithMagicComment {
	var results []arrayWithMagicComment

	var traverse func(*sitter.Node)
	traverse = func(n *sitter.Node) {
		if n.Type() == "array" {
			// Check children for magic comment
			for i := 0; i < int(n.ChildCount()); i++ {
				child := n.Child(i)
				if child.Type() == "comment" {
					text := content[child.StartByte():child.EndByte()]
					if magicCommentRegex.Match(text) {
						results = append(results, arrayWithMagicComment{
							array:        n,
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

type arrayElement struct {
	node         *sitter.Node
	beforeNodes  []*sitter.Node // Comments before this element
	afterNode    *sitter.Node   // Inline comment after element
	hasComma     bool
	commaNode    *sitter.Node
	sortKey      string // The extracted key for sorting
	isDeprecated bool
}

func sortArrayAST(arr arrayWithMagicComment, content []byte) (result []byte, changed bool) {
	// Extract elements after magic comment
	elements := extractArrayElementsAST(arr, content)

	if len(elements) <= 1 {
		return nil, false
	}

	// Extract sort keys for each element
	for _, elem := range elements {
		key, err := extractElementKey(elem, arr.sortConfig.Key, content)
		if err != nil {
			// For missing/invalid keys, mark with special prefix to sort last
			elem.sortKey = "\uffff" + string(content[elem.node.StartByte():elem.node.EndByte()])
		} else {
			elem.sortKey = key
		}
	}

	// Check if already sorted
	sorted := make([]*arrayElement, len(elements))
	copy(sorted, elements)

	// Sort elements, considering deprecated-at-end flag
	if arr.sortConfig.DeprecatedAtEnd {
		sort.Slice(sorted, func(i, j int) bool {
			// If one is deprecated and the other isn't, put non-deprecated first
			if sorted[i].isDeprecated != sorted[j].isDeprecated {
				return !sorted[i].isDeprecated
			}
			// Use string comparison to ensure \uffff prefix works for missing keys
			return sorted[i].sortKey < sorted[j].sortKey
		})
	} else {
		sort.Slice(sorted, func(i, j int) bool {
			// Check if either has missing key prefix
			iHasMissingKey := strings.HasPrefix(sorted[i].sortKey, "\uffff")
			jHasMissingKey := strings.HasPrefix(sorted[j].sortKey, "\uffff")

			// If one has missing key and other doesn't, sort missing key last
			if iHasMissingKey != jHasMissingKey {
				return !iHasMissingKey
			}

			// If both have or don't have missing keys, compare normally
			if iHasMissingKey && jHasMissingKey {
				// Both have missing keys, compare without prefix
				return sorted[i].sortKey < sorted[j].sortKey
			}

			// Neither has missing key, use compareKeys for proper type handling
			return compareKeys(sorted[i].sortKey, sorted[j].sortKey)
		})
	}

	alreadySorted := true
	for i := range elements {
		// Compare by node pointer to check if order changed
		if elements[i].node != sorted[i].node {
			alreadySorted = false
			break
		}
	}

	// Even if already sorted, check if formatting needs to change
	needsFormatting := false
	if alreadySorted && len(elements) > 1 {
		needsFormatting = checkArrayFormattingNeeded(arr, elements, content)
	}

	if alreadySorted && !needsFormatting {
		return nil, false
	}

	// Reconstruct the array with sorted elements
	return reconstructArrayAST(arr, sorted, content), true
}

func extractArrayElementsAST(arr arrayWithMagicComment, content []byte) []*arrayElement {
	// Pre-allocate elements slice based on estimated count
	elements := make([]*arrayElement, 0, int(arr.array.ChildCount())/2)
	var pendingComments []*sitter.Node

	// Start after magic comment
	startIdx := arr.magicIndex + 1

	for i := startIdx; i < int(arr.array.ChildCount()); i++ {
		child := arr.array.Child(i)

		switch child.Type() {
		case "comment":
			// Accumulate comments
			pendingComments = append(pendingComments, child)

		case ",":
			// Skip standalone commas
			continue

		case "]":
			// End of array
			break

		default:
			// This is an array element
			elem := &arrayElement{
				node:        child,
				beforeNodes: pendingComments,
			}

			// Check if this element has @deprecated annotation
			elem.isDeprecated = hasDeprecatedAnnotation(pendingComments, content)

			// Check if followed by comma and/or inline comment
			j := i + 1
			continueLoop := true
			for j < int(arr.array.ChildCount()) && continueLoop {
				next := arr.array.Child(j)
				switch next.Type() {
				case ",":
					elem.hasComma = true
					elem.commaNode = next
					j++
				case "comment":
					// Check if it's on the same line
					if next.StartPoint().Row == child.EndPoint().Row {
						elem.afterNode = next
						j++
					} else {
						continueLoop = false
					}
				default:
					continueLoop = false
				}
			}
			i = j - 1 // Update loop counter to skip processed nodes

			// Also check inline comment for @deprecated
			if !elem.isDeprecated && elem.afterNode != nil {
				text := string(content[elem.afterNode.StartByte():elem.afterNode.EndByte()])
				if strings.Contains(text, "@deprecated") {
					elem.isDeprecated = true
				}
			}

			elements = append(elements, elem)
			pendingComments = nil // Reset comments
		}
	}

	return elements
}

func extractElementKey(elem *arrayElement, keyPath string, content []byte) (string, error) {
	// If no key specified, use the raw element text for sorting
	if keyPath == "" {
		// Use raw text to preserve quotes and maintain consistent sorting
		return string(content[elem.node.StartByte():elem.node.EndByte()]), nil
	}

	// Determine element type
	switch elem.node.Type() {
	case "object":
		// For objects, extract the specified property
		return extractObjectProperty(elem.node, keyPath, content)
	case "array":
		// For tuples, extract by index
		return extractArrayIndex(elem.node, keyPath, content)
	default:
		// For scalars, use the value itself
		return extractValueAsString(elem.node, content), nil
	}
}

func extractObjectProperty(objNode *sitter.Node, keyPath string, content []byte) (string, error) {
	// Split keyPath for nested access (e.g., "profile.firstName")
	keys := strings.Split(keyPath, ".")
	currentNode := objNode

	for _, key := range keys {
		found := false
		// Look for the property in the current object
		for i := 0; i < int(currentNode.ChildCount()); i++ {
			child := currentNode.Child(i)
			if child.Type() == "pair" {
				keyNode := child.ChildByFieldName("key")
				if keyNode != nil {
					propKey := extractKeyAST(keyNode, content)
					if propKey == key {
						valueNode := child.ChildByFieldName("value")
						if valueNode != nil {
							if len(keys) > 1 && valueNode.Type() == "object" {
								// Continue traversing for nested property
								currentNode = valueNode
								found = true
								break
							}
							// Found the final value
							return extractValueAsString(valueNode, content), nil
						}
					}
				}
			}
		}
		if !found {
			return "", fmt.Errorf("key not found: %s", key)
		}
	}

	return "", fmt.Errorf("key not found: %s", keyPath)
}

func extractArrayIndex(arrNode *sitter.Node, indexStr string, content []byte) (string, error) {
	index := 0
	_, err := fmt.Sscanf(indexStr, "%d", &index)
	if err != nil {
		return "", fmt.Errorf("invalid index: %s", indexStr)
	}

	// Count actual elements (skip commas and comments)
	elementCount := 0
	for i := 0; i < int(arrNode.ChildCount()); i++ {
		child := arrNode.Child(i)
		if child.Type() != "," && child.Type() != "comment" && child.Type() != "[" && child.Type() != "]" {
			if elementCount == index {
				return extractValueAsString(child, content), nil
			}
			elementCount++
		}
	}

	return "", fmt.Errorf("index out of bounds: %d", index)
}

func extractValueAsString(node *sitter.Node, content []byte) string {
	switch node.Type() {
	case "string":
		// Remove quotes from strings
		text := string(content[node.StartByte():node.EndByte()])
		return strings.Trim(text, "\"'`")
	case "number":
		return string(content[node.StartByte():node.EndByte()])
	case "true", "false":
		return string(content[node.StartByte():node.EndByte()])
	case "null":
		return "null"
	default:
		// For complex types, use the raw text
		return string(content[node.StartByte():node.EndByte()])
	}
}

func compareKeys(a, b string) bool {
	// Try to compare as numbers first
	var numA, numB float64
	_, errA := fmt.Sscanf(a, "%f", &numA)
	_, errB := fmt.Sscanf(b, "%f", &numB)

	if errA == nil && errB == nil {
		// Both are numbers
		return numA < numB
	}

	// Try to compare as booleans
	if (a == "true" || a == "false") && (b == "true" || b == "false") {
		// false < true
		return a == "false" && b == "true"
	}

	// Default to string comparison
	return a < b
}

func checkArrayFormattingNeeded(arr arrayWithMagicComment, elements []*arrayElement, content []byte) bool {
	// For single-line arrays, no formatting changes needed
	if len(elements) > 0 {
		firstElem := elements[0]
		lastElem := elements[len(elements)-1]
		if firstElem.node.StartPoint().Row == lastElem.node.EndPoint().Row {
			return false
		}
	}

	// Check if there's an extra newline between elements
	for i := 0; i < len(elements)-1; i++ {
		elem := elements[i]
		nextElem := elements[i+1]

		// Find the end of current element (including comma and inline comment)
		endNode := elem.node
		if elem.afterNode != nil {
			endNode = elem.afterNode
		} else if elem.commaNode != nil {
			endNode = elem.commaNode
		}

		// Count newlines between elements
		startByte := endNode.EndByte()
		endByte := nextElem.node.StartByte()

		// Handle beforeNodes of next element
		if len(nextElem.beforeNodes) > 0 {
			endByte = nextElem.beforeNodes[0].StartByte()
		}

		between := content[startByte:endByte]
		newlineCount := 0
		for _, b := range between {
			if b == '\n' {
				newlineCount++
			}
		}

		// If with-new-line is set, we expect 2 newlines between elements
		expectedNewlines := 1
		if arr.sortConfig.WithNewLine {
			expectedNewlines = 2
		}

		if newlineCount != expectedNewlines {
			return true
		}
	}

	return false
}

func reconstructArrayAST(arr arrayWithMagicComment, sortedElems []*arrayElement, content []byte) []byte {
	var result bytes.Buffer

	// Check if this is a single-line array (all elements on same line)
	isSingleLine := false
	if len(sortedElems) > 0 {
		firstElem := sortedElems[0]
		lastElem := sortedElems[len(sortedElems)-1]
		if firstElem.node.StartPoint().Row == lastElem.node.EndPoint().Row {
			isSingleLine = true
		}
	}

	// Extract common indentation
	commonIndent := ""
	for i := arr.magicIndex + 1; i < int(arr.array.ChildCount()); i++ {
		child := arr.array.Child(i)
		if child.Type() != "," && child.Type() != "comment" && child.Type() != "]" {
			// Found first element
			elemStart := child.StartByte()
			lineStart := elemStart
			for lineStart > 0 && content[lineStart-1] != '\n' {
				lineStart--
			}
			if lineStart < elemStart {
				commonIndent = string(content[lineStart:elemStart])
			}
			break
		}
	}

	// Write everything up to and including the magic comment
	for i := 0; i <= arr.magicIndex; i++ {
		child := arr.array.Child(i)

		// Write any whitespace before this child
		if i == 0 {
			result.Write(content[arr.array.StartByte():child.StartByte()])
		} else {
			prevChild := arr.array.Child(i - 1)
			result.Write(content[prevChild.EndByte():child.StartByte()])
		}

		// Write the child itself
		result.Write(content[child.StartByte():child.EndByte()])
	}

	// Write newline after magic comment (unless single line)
	if arr.magicIndex+1 < int(arr.array.ChildCount()) && !isSingleLine {
		result.WriteByte('\n')
	}

	// Write sorted elements
	for i, elem := range sortedElems {
		if isSingleLine {
			// For single-line arrays, write minimal spacing
			if i == 0 {
				result.WriteByte('\n')
				result.WriteString(commonIndent)
			}
		} else {
			// Write any comments before this element
			for _, commentNode := range elem.beforeNodes {
				// Find whitespace before comment
				if len(result.Bytes()) > 0 {
					commentStart := commentNode.StartByte()
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

			// Use common indentation for all elements
			result.WriteString(commonIndent)
		}

		// Write the element itself
		result.Write(content[elem.node.StartByte():elem.node.EndByte()])

		// Handle comma
		if i < len(sortedElems)-1 {
			if elem.hasComma && elem.commaNode != nil {
				// Use original comma
				result.Write(content[elem.commaNode.StartByte():elem.commaNode.EndByte()])
			} else {
				// Add comma if missing
				result.WriteByte(',')
			}

			// For single-line arrays, add space after comma
			if isSingleLine {
				result.WriteByte(' ')
			}
		} else {
			// Last element - check if original had trailing comma
			originalLastElem := findOriginalLastArrayElement(arr)
			if originalLastElem != nil && originalLastElem.hasComma {
				result.WriteByte(',')
			}
		}

		// Write inline comment if present
		if elem.afterNode != nil {
			// Since elements may have been reordered, we can't rely on original byte positions
			// Just use a reasonable default spacing
			result.WriteByte(' ')
			result.Write(content[elem.afterNode.StartByte():elem.afterNode.EndByte()])
		}

		// Add newline if not last or if there's more content (and not single line)
		if !isSingleLine && i < len(sortedElems)-1 {
			result.WriteByte('\n')
			// Add extra newline if with-new-line option is set
			if arr.sortConfig.WithNewLine {
				result.WriteByte('\n')
			}
		}
	}

	// Write closing bracket and any trailing content
	closingBracketIdx := -1
	for i := int(arr.array.ChildCount()) - 1; i >= 0; i-- {
		child := arr.array.Child(i)
		if child.Type() == "]" {
			closingBracketIdx = i
			break
		}
	}

	if closingBracketIdx > 0 {
		closingBracket := arr.array.Child(closingBracketIdx)

		// Write any whitespace/newlines before closing bracket
		originalSpacing := findOriginalArrayClosingSpacing(arr)
		if originalSpacing != "" {
			result.WriteString(originalSpacing)
		} else {
			result.WriteByte('\n')
		}

		result.Write(content[closingBracket.StartByte():closingBracket.EndByte()])
	}

	return result.Bytes()
}

func findOriginalLastArrayElement(arr arrayWithMagicComment) *arrayElement {
	var lastElem *arrayElement

	for i := arr.magicIndex + 1; i < int(arr.array.ChildCount()); i++ {
		child := arr.array.Child(i)
		if child.Type() != "," && child.Type() != "comment" && child.Type() != "]" {
			lastElem = &arrayElement{
				node: child,
			}
			// Check if followed by comma
			if i+1 < int(arr.array.ChildCount()) {
				next := arr.array.Child(i + 1)
				if next.Type() == "," {
					lastElem.hasComma = true
					lastElem.commaNode = next
				}
			}
		}
	}

	return lastElem
}

func findOriginalArrayClosingSpacing(arr arrayWithMagicComment) string {
	// Just return a newline - we handle inline comments separately
	// and don't want to duplicate them
	return "\n"
}

// Constructor sorting functionality

type constructorWithMagicComment struct {
	formalParams *sitter.Node
	magicComment *sitter.Node
	magicIndex   int // Index of magic comment in children
	sortConfig   SortConfig
}

func findConstructorsWithMagicCommentsAST(node *sitter.Node, content []byte) []constructorWithMagicComment {
	var results []constructorWithMagicComment

	var traverse func(*sitter.Node)
	traverse = func(n *sitter.Node) {
		if n.Type() == "formal_parameters" {
			// Check children for magic comment
			for i := 0; i < int(n.ChildCount()); i++ {
				child := n.Child(i)
				if child.Type() == "comment" {
					text := content[child.StartByte():child.EndByte()]
					if magicCommentRegex.Match(text) {
						results = append(results, constructorWithMagicComment{
							formalParams: n,
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

type constructorParam struct {
	node         *sitter.Node // The required_parameter node
	name         string       // Parameter name (from identifier)
	beforeNodes  []*sitter.Node // Comments before this parameter
	afterNode    *sitter.Node   // Inline comment after parameter
	hasComma     bool
	commaNode    *sitter.Node
	isDeprecated bool
}

func sortConstructorAST(constr constructorWithMagicComment, content []byte) ([]byte, bool) {
	// Extract parameters after magic comment
	params := extractConstructorParamsAST(constr, content)

	if len(params) <= 1 {
		return nil, false
	}

	// Check if already sorted
	sorted := make([]*constructorParam, len(params))
	copy(sorted, params)

	// Sort parameters, considering deprecated-at-end flag
	if constr.sortConfig.DeprecatedAtEnd {
		sort.Slice(sorted, func(i, j int) bool {
			// If one is deprecated and the other isn't, put non-deprecated first
			if sorted[i].isDeprecated != sorted[j].isDeprecated {
				return !sorted[i].isDeprecated
			}
			// Otherwise sort alphabetically by parameter name
			return sorted[i].name < sorted[j].name
		})
	} else {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].name < sorted[j].name
		})
	}

	alreadySorted := true
	for i := range params {
		if params[i].name != sorted[i].name {
			alreadySorted = false
			break
		}
		// For deprecated-at-end, also check if deprecated parameters are in the right place
		if constr.sortConfig.DeprecatedAtEnd && params[i].isDeprecated != sorted[i].isDeprecated {
			alreadySorted = false
			break
		}
	}

	// Even if already sorted, check if formatting needs to change
	needsFormatting := false
	if alreadySorted && len(params) > 1 {
		needsFormatting = checkConstructorFormattingNeeded(constr, params, content)
	}

	if alreadySorted && !needsFormatting {
		return nil, false
	}

	// Reconstruct the constructor with sorted parameters
	return reconstructConstructorAST(constr, sorted, content), true
}

func extractConstructorParamsAST(constr constructorWithMagicComment, content []byte) []*constructorParam {
	var params []*constructorParam
	var pendingComments []*sitter.Node

	// Start after magic comment
	startIdx := constr.magicIndex + 1

	for i := startIdx; i < int(constr.formalParams.ChildCount()); i++ {
		child := constr.formalParams.Child(i)

		switch child.Type() {
		case "comment":
			// Accumulate comments
			pendingComments = append(pendingComments, child)

		case "required_parameter", "optional_parameter":
			param := &constructorParam{
				node:        child,
				beforeNodes: pendingComments,
			}

			// Check if this parameter has @deprecated annotation
			param.isDeprecated = hasDeprecatedAnnotation(pendingComments, content)

			// Extract parameter name from pattern field
			patternNode := child.ChildByFieldName("pattern")
			if patternNode != nil {
				switch patternNode.Type() {
				case "identifier":
					// Simple parameter like: someParam: string
					param.name = string(content[patternNode.StartByte():patternNode.EndByte()])
				case "object_pattern":
					// Destructured parameter like: { someParam }: { someParam: string }
					// For sorting, use the first property name
					for i := 0; i < int(patternNode.ChildCount()); i++ {
						propChild := patternNode.Child(i)
						if propChild.Type() == "shorthand_property_identifier_pattern" {
							param.name = string(content[propChild.StartByte():propChild.EndByte()])
							break
						}
					}
				default:
					// For other patterns (array destructuring, etc.), use the full text
					param.name = string(content[patternNode.StartByte():patternNode.EndByte()])
				}
			}

			// Check if followed by comma and/or inline comment
			j := i + 1
			continueLoop := true
			lastNode := child // Track the last node for line comparison
			for j < int(constr.formalParams.ChildCount()) && continueLoop {
				next := constr.formalParams.Child(j)
				switch next.Type() {
				case ",":
					param.hasComma = true
					param.commaNode = next
					lastNode = next
					j++
				case "comment":
					// Check if it's on the same line as the parameter or comma
					if next.StartPoint().Row == lastNode.EndPoint().Row {
						param.afterNode = next
						j++
					} else {
						continueLoop = false
					}
				default:
					continueLoop = false
				}
			}
			i = j - 1 // Update loop counter to skip processed nodes

			// Also check inline comment for @deprecated
			if !param.isDeprecated && param.afterNode != nil {
				text := string(content[param.afterNode.StartByte():param.afterNode.EndByte()])
				if strings.Contains(text, "@deprecated") {
					param.isDeprecated = true
				}
			}

			params = append(params, param)
			pendingComments = nil // Reset comments

		case ",":
			// Standalone comma (shouldn't happen if we handle it above)
			continue

		case ")":
			// End of parameters
			break
		}
	}

	return params
}

func checkConstructorFormattingNeeded(constr constructorWithMagicComment, params []*constructorParam, content []byte) bool {
	// Check if there's an extra newline between parameters
	for i := 0; i < len(params)-1; i++ {
		param := params[i]
		nextParam := params[i+1]

		// Find the end of current parameter (including comma and inline comment)
		endNode := param.node
		if param.afterNode != nil {
			endNode = param.afterNode
		} else if param.commaNode != nil {
			endNode = param.commaNode
		}

		// Count newlines between parameters
		startByte := endNode.EndByte()
		endByte := nextParam.node.StartByte()

		// Handle beforeNodes of next parameter
		if len(nextParam.beforeNodes) > 0 {
			endByte = nextParam.beforeNodes[0].StartByte()
		}

		between := content[startByte:endByte]
		newlineCount := 0
		for _, b := range between {
			if b == '\n' {
				newlineCount++
			}
		}

		// If with-new-line is set, we expect 2 newlines between parameters
		expectedNewlines := 1
		if constr.sortConfig.WithNewLine {
			expectedNewlines = 2
		}

		if newlineCount != expectedNewlines {
			return true
		}
	}

	return false
}

func reconstructConstructorAST(constr constructorWithMagicComment, sortedParams []*constructorParam, content []byte) []byte {
	var result bytes.Buffer

	// Extract common indentation from first original parameter after magic comment
	commonIndent := ""
	for i := constr.magicIndex + 1; i < int(constr.formalParams.ChildCount()); i++ {
		child := constr.formalParams.Child(i)
		if child.Type() == "required_parameter" || child.Type() == "optional_parameter" {
			// Found first parameter
			paramStart := child.StartByte()
			lineStart := paramStart
			for lineStart > 0 && content[lineStart-1] != '\n' {
				lineStart--
			}
			if lineStart < paramStart {
				commonIndent = string(content[lineStart:paramStart])
			}
			break
		}
	}

	// Write everything up to and including the magic comment
	for i := 0; i <= constr.magicIndex; i++ {
		child := constr.formalParams.Child(i)

		// Write any whitespace before this child
		if i == 0 {
			result.Write(content[constr.formalParams.StartByte():child.StartByte()])
		} else {
			prevChild := constr.formalParams.Child(i - 1)
			result.Write(content[prevChild.EndByte():child.StartByte()])
		}

		// Write the child itself
		result.Write(content[child.StartByte():child.EndByte()])
	}

	// Write newline after magic comment
	if constr.magicIndex+1 < int(constr.formalParams.ChildCount()) {
		result.WriteByte('\n')
	}

	// Write sorted parameters
	for i, param := range sortedParams {
		// Write any comments before this parameter
		for _, commentNode := range param.beforeNodes {
			// Find whitespace before comment
			if len(result.Bytes()) > 0 {
				commentStart := commentNode.StartByte()
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

		// Use common indentation for all parameters
		result.WriteString(commonIndent)

		// Write the parameter itself (preserving all formatting)
		result.Write(content[param.node.StartByte():param.node.EndByte()])

		// Handle comma
		if i < len(sortedParams)-1 {
			if param.hasComma && param.commaNode != nil {
				// Use original comma
				result.Write(content[param.commaNode.StartByte():param.commaNode.EndByte()])
			} else {
				// Add comma if missing
				result.WriteByte(',')
			}
		} else {
			// Last parameter - check if original had trailing comma
			originalLastParam := findOriginalLastConstructorParam(constr, content)
			if originalLastParam != nil && originalLastParam.hasComma {
				result.WriteByte(',')
			}
		}

		// Write inline comment if present
		if param.afterNode != nil {
			result.WriteByte(' ')
			result.Write(content[param.afterNode.StartByte():param.afterNode.EndByte()])
		}

		// Add newline if not last or if there's more content
		if i < len(sortedParams)-1 {
			result.WriteByte('\n')
			// Add extra newline if with-new-line option is set
			if constr.sortConfig.WithNewLine {
				result.WriteByte('\n')
			}
		}
	}

	// Write closing parenthesis and any trailing content
	closingParenIdx := -1
	for i := int(constr.formalParams.ChildCount()) - 1; i >= 0; i-- {
		child := constr.formalParams.Child(i)
		if child.Type() == ")" {
			closingParenIdx = i
			break
		}
	}

	if closingParenIdx > 0 {
		closingParen := constr.formalParams.Child(closingParenIdx)

		// Write any whitespace/newlines before closing parenthesis
		originalSpacing := findOriginalConstructorClosingSpacing(constr, content)
		if originalSpacing != "" {
			result.WriteString(originalSpacing)
		} else {
			result.WriteByte('\n')
		}

		result.Write(content[closingParen.StartByte():closingParen.EndByte()])
	}

	return result.Bytes()
}

func findOriginalLastConstructorParam(constr constructorWithMagicComment, content []byte) *constructorParam {
	// This function is only used to check if the original last parameter had a trailing comma
	// We don't need to extract the full parameter info, just check for trailing comma
	
	// Find the last parameter node
	var lastParamNode *sitter.Node
	for i := constr.magicIndex + 1; i < int(constr.formalParams.ChildCount()); i++ {
		child := constr.formalParams.Child(i)
		if child.Type() == "required_parameter" || child.Type() == "optional_parameter" {
			lastParamNode = child
		}
	}
	
	if lastParamNode == nil {
		return nil
	}
	
	// Check if there's a comma after the last parameter
	foundLastParam := false
	for i := 0; i < int(constr.formalParams.ChildCount()); i++ {
		child := constr.formalParams.Child(i)
		if child == lastParamNode {
			foundLastParam = true
			continue
		}
		if foundLastParam && child.Type() == "," {
			return &constructorParam{hasComma: true}
		}
		if foundLastParam && (child.Type() == "required_parameter" || child.Type() == "optional_parameter") {
			// Found another parameter, so the last one didn't have a trailing comma
			break
		}
	}
	
	return &constructorParam{hasComma: false}
}

func findOriginalConstructorClosingSpacing(constr constructorWithMagicComment, content []byte) string {
	// Find the last meaningful content (parameter, comma, or comment)
	lastContentEnd := constr.magicComment.EndByte()

	for i := int(constr.formalParams.ChildCount()) - 1; i > constr.magicIndex; i-- {
		child := constr.formalParams.Child(i)
		if child.Type() == "required_parameter" || child.Type() == "optional_parameter" || child.Type() == "," || child.Type() == "comment" {
			lastContentEnd = child.EndByte()
			break
		}
	}

	// Find closing parenthesis
	for i := int(constr.formalParams.ChildCount()) - 1; i >= 0; i-- {
		child := constr.formalParams.Child(i)
		if child.Type() == ")" {
			return string(content[lastContentEnd:child.StartByte()])
		}
	}

	return "\n"
}
