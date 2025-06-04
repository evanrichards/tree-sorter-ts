package strategies

import (
	"fmt"
	"strings"

	"github.com/evanrichards/tree-sorter-ts/internal/sorting/common"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"

	sitter "github.com/smacker/go-tree-sitter"
)

// ArrayKeyStrategy sorts array elements by a specified key path
type ArrayKeyStrategy struct {
	KeyPath string
}

func (s *ArrayKeyStrategy) ExtractKey(item interfaces.SortableItem, content []byte) (string, error) {
	node := item.GetNode()
	
	// If no key specified, use the raw element text for sorting
	if s.KeyPath == "" {
		return string(content[node.StartByte():node.EndByte()]), nil
	}

	// Determine element type
	switch node.Type() {
	case "object":
		// For objects, extract the specified property
		return s.extractObjectProperty(node, s.KeyPath, content)
	case "array":
		// For tuples, extract by index
		return s.extractArrayIndex(node, s.KeyPath, content)
	default:
		// For scalars, use the value itself
		return common.ExtractValueAsString(node, content), nil
	}
}

func (s *ArrayKeyStrategy) GetName() string {
	if s.KeyPath == "" {
		return "array-element-value"
	}
	return fmt.Sprintf("array-key[%s]", s.KeyPath)
}

func (s *ArrayKeyStrategy) extractObjectProperty(objNode *sitter.Node, keyPath string, content []byte) (string, error) {
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
					propKey := common.ExtractKeyFromNode(keyNode, content)
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
							return common.ExtractValueAsString(valueNode, content), nil
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

func (s *ArrayKeyStrategy) extractArrayIndex(arrNode *sitter.Node, indexStr string, content []byte) (string, error) {
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
				return common.ExtractValueAsString(child, content), nil
			}
			elementCount++
		}
	}

	return "", fmt.Errorf("index out of bounds: %d", index)
}