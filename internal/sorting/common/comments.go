package common

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// ExtractCommentText extracts the text content from a comment node, removing comment markers
func ExtractCommentText(commentNode *sitter.Node, content []byte) string {
	if commentNode == nil {
		return ""
	}
	// Extract comment text (remove comment markers)
	text := string(content[commentNode.StartByte():commentNode.EndByte()])
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "//") {
		text = strings.TrimSpace(text[2:])
	} else if strings.HasPrefix(text, "/*") && strings.HasSuffix(text, "*/") {
		text = strings.TrimSpace(text[2 : len(text)-2])
	}
	return text
}

// HasDeprecatedAnnotation checks if any of the comment nodes contain @deprecated annotation
func HasDeprecatedAnnotation(nodes []*sitter.Node, content []byte) bool {
	for _, node := range nodes {
		text := string(content[node.StartByte():node.EndByte()])
		if strings.Contains(text, "@deprecated") {
			return true
		}
	}
	return false
}

// FindCommentTextForSorting finds the appropriate comment text for sorting
// It checks inline comments first, then preceding comments
func FindCommentTextForSorting(beforeNodes []*sitter.Node, afterNode *sitter.Node, content []byte) (string, error) {
	// Check for inline comment after element first
	if afterNode != nil {
		text := ExtractCommentText(afterNode, content)
		if text != "" {
			return text, nil
		}
	}
	// Check for comment before element
	if len(beforeNodes) > 0 {
		// Use the last comment before the element
		for i := len(beforeNodes) - 1; i >= 0; i-- {
			text := ExtractCommentText(beforeNodes[i], content)
			if text != "" {
				return text, nil
			}
		}
	}
	// If no comment found, return error
	return "", fmt.Errorf("no comment found for sorting")
}

