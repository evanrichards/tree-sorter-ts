package reconstruction

import (
	"bytes"
	"fmt"

	"github.com/evanrichards/tree-sorter-ts/internal/config"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/types/objects"

	sitter "github.com/smacker/go-tree-sitter"
)

// ObjectReconstructor rebuilds object content with sorted properties
type ObjectReconstructor struct{}

// NewObjectReconstructor creates a new object reconstructor
func NewObjectReconstructor() *ObjectReconstructor {
	return &ObjectReconstructor{}
}

// Reconstruct generates new content with sorted object properties
func (r *ObjectReconstructor) Reconstruct(sortable interfaces.Sortable, sortedItems []interfaces.SortableItem, configInterface interface{}, content []byte) ([]byte, error) {
	objectSorter, ok := sortable.(*objects.ObjectSorter)
	if !ok {
		return nil, fmt.Errorf("expected ObjectSorter, got %T", sortable)
	}
	
	cfg, ok := configInterface.(config.SortConfig)
	if !ok {
		return nil, fmt.Errorf("expected config.SortConfig, got %T", configInterface)
	}

	objectNode := objectSorter.GetNode()
	magicIndex := objectSorter.GetMagicCommentIndex()

	var result bytes.Buffer

	// Write content before the object
	result.Write(content[:objectNode.StartByte()])

	// Write opening brace and up to magic comment
	result.WriteByte('{')

	// Find and write content up to magic comment
	for i := 0; i <= magicIndex; i++ {
		child := objectNode.Child(i)
		if i == 0 && child.Type() == "{" {
			continue // Skip opening brace, already written
		}
		
		// Add whitespace/newlines before child if needed
		if i > 0 {
			prevChild := objectNode.Child(i - 1)
			r.writeWhitespaceBetween(prevChild, child, content, &result)
		} else {
			// First child after opening brace
			r.writeWhitespaceAfter(objectNode, child, content, &result)
		}
		
		// Write the child content
		result.Write(content[child.StartByte():child.EndByte()])
	}

	// Write sorted properties
	for i, item := range sortedItems {
		prop := item.(*objects.Property)
		
		// Add appropriate spacing/newlines
		if i == 0 {
			// First property - check if we need newline after magic comment
			if cfg.WithNewLine {
				result.WriteByte('\n')
			} else {
				result.WriteByte(' ')
			}
		} else {
			// Subsequent properties - use original spacing pattern
			result.WriteByte('\n')
		}

		// Write before comments (if any)
		for _, comment := range prop.BeforeNodes {
			r.writeIndentation(&result)
			result.Write(content[comment.StartByte():comment.EndByte()])
			result.WriteByte('\n')
		}

		// Write the property itself with proper indentation
		r.writeIndentation(&result)
		result.Write(content[prop.PairNode.StartByte():prop.PairNode.EndByte()])

		// Write comma if present
		if prop.HasComma {
			result.WriteByte(',')
		}

		// Write after comment if present (inline comment)
		if prop.AfterNode != nil {
			result.WriteByte(' ')
			result.Write(content[prop.AfterNode.StartByte():prop.AfterNode.EndByte()])
		}
	}

	// Find closing brace and write final content
	for i := int(objectNode.ChildCount()) - 1; i >= 0; i-- {
		child := objectNode.Child(i)
		if child.Type() == "}" {
			// Add newline before closing brace if we have properties
			if len(sortedItems) > 0 {
				result.WriteByte('\n')
			}
			result.Write(content[child.StartByte():child.EndByte()])
			break
		}
	}

	// Write content after the object
	result.Write(content[objectNode.EndByte():])

	return result.Bytes(), nil
}

// writeWhitespaceBetween writes whitespace/newlines between two nodes
func (r *ObjectReconstructor) writeWhitespaceBetween(prev, current *sitter.Node, content []byte, result *bytes.Buffer) {
	if prev.EndByte() < current.StartByte() {
		whitespace := content[prev.EndByte():current.StartByte()]
		result.Write(whitespace)
	}
}

// writeWhitespaceAfter writes whitespace after opening brace to first child
func (r *ObjectReconstructor) writeWhitespaceAfter(parent, child *sitter.Node, content []byte, result *bytes.Buffer) {
	// Find opening brace
	openBraceEnd := parent.StartByte() + 1 // Assume '{' is one byte
	if openBraceEnd < child.StartByte() {
		whitespace := content[openBraceEnd:child.StartByte()]
		result.Write(whitespace)
	}
}

// writeIndentation writes proper indentation (2 spaces)
func (r *ObjectReconstructor) writeIndentation(result *bytes.Buffer) {
	result.WriteString("  ")
}

// CanHandle returns true if this reconstructor can handle the given sortable
func (r *ObjectReconstructor) CanHandle(sortable interfaces.Sortable) bool {
	_, ok := sortable.(*objects.ObjectSorter)
	return ok
}