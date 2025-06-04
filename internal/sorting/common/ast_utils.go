package common

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// ExtractKeyFromNode extracts the key text from an AST key node
func ExtractKeyFromNode(keyNode *sitter.Node, content []byte) string {
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

// ExtractValueAsString extracts a value node as a string for comparison
func ExtractValueAsString(node *sitter.Node, content []byte) string {
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

// TrimQuotes removes surrounding quotes from a string
func TrimQuotes(text string) string {
	return strings.Trim(text, "\"'`")
}