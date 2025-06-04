package reconstruction

import (
	"bytes"
	"fmt"

	"github.com/evanrichards/tree-sorter-ts/internal/config"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/types/arrays"

	sitter "github.com/smacker/go-tree-sitter"
)

// ArrayReconstructor rebuilds array content with sorted elements
type ArrayReconstructor struct{}

// NewArrayReconstructor creates a new array reconstructor
func NewArrayReconstructor() *ArrayReconstructor {
	return &ArrayReconstructor{}
}

// Reconstruct generates new content with sorted array elements
func (r *ArrayReconstructor) Reconstruct(sortable interfaces.Sortable, sortedItems []interfaces.SortableItem, configInterface interface{}, content []byte) ([]byte, error) {
	arraySorter, ok := sortable.(*arrays.ArraySorter)
	if !ok {
		return nil, fmt.Errorf("expected ArraySorter, got %T", sortable)
	}
	
	cfg, ok := configInterface.(config.SortConfig)
	if !ok {
		return nil, fmt.Errorf("expected config.SortConfig, got %T", configInterface)
	}

	arrayNode := arraySorter.GetNode()
	magicIndex := arraySorter.GetMagicCommentIndex()

	var result bytes.Buffer

	// Write content before the array
	result.Write(content[:arrayNode.StartByte()])

	// Write opening bracket and up to magic comment
	result.WriteByte('[')

	// Find and write content up to magic comment
	for i := 0; i <= magicIndex; i++ {
		child := arrayNode.Child(i)
		if i == 0 && child.Type() == "[" {
			continue // Skip opening bracket, already written
		}
		
		// Add whitespace/newlines before child if needed
		if i > 0 {
			prevChild := arrayNode.Child(i - 1)
			r.writeWhitespaceBetween(prevChild, child, content, &result)
		} else {
			// First child after opening bracket
			r.writeWhitespaceAfter(arrayNode, child, content, &result)
		}
		
		// Write the child content
		result.Write(content[child.StartByte():child.EndByte()])
	}

	// Write sorted elements
	for i, item := range sortedItems {
		elem := item.(*arrays.Element)
		
		// Add appropriate spacing/newlines
		if i == 0 {
			// First element - check if we need newline after magic comment
			if cfg.WithNewLine {
				result.WriteByte('\n')
			} else {
				result.WriteByte(' ')
			}
		} else {
			// Subsequent elements - use original spacing pattern
			result.WriteByte('\n')
		}

		// Write before comments (if any)
		for _, comment := range elem.BeforeNodes {
			r.writeIndentation(&result)
			result.Write(content[comment.StartByte():comment.EndByte()])
			result.WriteByte('\n')
		}

		// Write the element itself with proper indentation
		r.writeIndentation(&result)
		result.Write(content[elem.Node.StartByte():elem.Node.EndByte()])

		// Write comma if present
		if elem.HasComma {
			result.WriteByte(',')
		}

		// Write after comment if present (inline comment)
		if elem.AfterNode != nil {
			result.WriteByte(' ')
			result.Write(content[elem.AfterNode.StartByte():elem.AfterNode.EndByte()])
		}
	}

	// Find closing bracket and write final content
	for i := int(arrayNode.ChildCount()) - 1; i >= 0; i-- {
		child := arrayNode.Child(i)
		if child.Type() == "]" {
			// Add newline before closing bracket if we have elements
			if len(sortedItems) > 0 {
				result.WriteByte('\n')
			}
			result.Write(content[child.StartByte():child.EndByte()])
			break
		}
	}

	// Write content after the array
	result.Write(content[arrayNode.EndByte():])

	return result.Bytes(), nil
}

// writeWhitespaceBetween writes whitespace/newlines between two nodes
func (r *ArrayReconstructor) writeWhitespaceBetween(prev, current *sitter.Node, content []byte, result *bytes.Buffer) {
	if prev.EndByte() < current.StartByte() {
		whitespace := content[prev.EndByte():current.StartByte()]
		result.Write(whitespace)
	}
}

// writeWhitespaceAfter writes whitespace after opening bracket to first child
func (r *ArrayReconstructor) writeWhitespaceAfter(parent, child *sitter.Node, content []byte, result *bytes.Buffer) {
	// Find opening bracket
	openBracketEnd := parent.StartByte() + 1 // Assume '[' is one byte
	if openBracketEnd < child.StartByte() {
		whitespace := content[openBracketEnd:child.StartByte()]
		result.Write(whitespace)
	}
}

// writeIndentation writes proper indentation (2 spaces)
func (r *ArrayReconstructor) writeIndentation(result *bytes.Buffer) {
	result.WriteString("  ")
}

// CanHandle returns true if this reconstructor can handle the given sortable
func (r *ArrayReconstructor) CanHandle(sortable interfaces.Sortable) bool {
	_, ok := sortable.(*arrays.ArraySorter)
	return ok
}