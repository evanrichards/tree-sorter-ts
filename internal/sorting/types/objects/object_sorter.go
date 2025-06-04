package objects

import (
	"sort"
	"strings"

	"github.com/evanrichards/tree-sorter-ts/internal/sorting/common"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"

	sitter "github.com/smacker/go-tree-sitter"
)

// ObjectSorter handles sorting of object properties
type ObjectSorter struct {
	node         *sitter.Node
	magicComment *sitter.Node
	magicIndex   int
}

// NewObjectSorter creates a new object sorter
func NewObjectSorter(objectNode, magicComment *sitter.Node, magicIndex int) *ObjectSorter {
	return &ObjectSorter{
		node:         objectNode,
		magicComment: magicComment,
		magicIndex:   magicIndex,
	}
}

// Extract finds and extracts sortable properties from the object
func (o *ObjectSorter) Extract(node *sitter.Node, content []byte) ([]interfaces.SortableItem, error) {
	var properties []interfaces.SortableItem
	var pendingComments []*sitter.Node

	// Start after magic comment
	startIdx := o.magicIndex + 1

	for i := startIdx; i < int(o.node.ChildCount()); i++ {
		child := o.node.Child(i)

		switch child.Type() {
		case "comment":
			// Accumulate comments
			pendingComments = append(pendingComments, child)

		case "pair":
			prop := NewProperty(child, content)
			prop.BeforeNodes = pendingComments

			// Check if this property has @deprecated annotation
			prop.isDeprecated = common.HasDeprecatedAnnotation(pendingComments, content)

			// Check if followed by comma and/or inline comment
			j := i + 1
			continueLoop := true
			for j < int(o.node.ChildCount()) && continueLoop {
				next := o.node.Child(j)
				switch next.Type() {
				case ",":
					prop.HasComma = true
					prop.CommaNode = next
					j++
				case "comment":
					// Check if it's on the same line
					if next.StartPoint().Row == child.EndPoint().Row {
						prop.AfterNode = next
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
			if !prop.isDeprecated && prop.AfterNode != nil {
				text := string(content[prop.AfterNode.StartByte():prop.AfterNode.EndByte()])
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

	return properties, nil
}

// Sort applies the strategy to sort the properties
func (o *ObjectSorter) Sort(items []interfaces.SortableItem, strategy interfaces.SortStrategy, deprecatedAtEnd bool, content []byte) ([]interfaces.SortableItem, error) {
	if len(items) <= 1 {
		return items, nil
	}

	// Extract sort keys for each property
	for _, item := range items {
		prop := item.(*Property)
		sortKey, err := item.GetSortKey(strategy, content)
		if err != nil {
			// For missing/invalid keys, mark with special prefix to sort last
			prop.SortKey = "\uffff" + prop.Key
		} else {
			prop.SortKey = sortKey
		}
	}

	// Make a copy for sorting
	sorted := make([]interfaces.SortableItem, len(items))
	copy(sorted, items)

	// Sort properties, considering deprecated-at-end flag
	if deprecatedAtEnd {
		sort.Slice(sorted, func(i, j int) bool {
			propI := sorted[i].(*Property)
			propJ := sorted[j].(*Property)
			// If one is deprecated and the other isn't, put non-deprecated first
			if propI.isDeprecated != propJ.isDeprecated {
				return !propI.isDeprecated
			}
			// Otherwise sort alphabetically
			return propI.SortKey < propJ.SortKey
		})
	} else {
		sort.Slice(sorted, func(i, j int) bool {
			propI := sorted[i].(*Property)
			propJ := sorted[j].(*Property)
			return propI.SortKey < propJ.SortKey
		})
	}

	return sorted, nil
}

// CheckIfSorted determines if properties are already sorted according to strategy
func (o *ObjectSorter) CheckIfSorted(items []interfaces.SortableItem, strategy interfaces.SortStrategy, deprecatedAtEnd bool, content []byte) bool {
	if len(items) <= 1 {
		return true
	}

	sorted, err := o.Sort(items, strategy, deprecatedAtEnd, content)
	if err != nil {
		return false
	}

	// Compare original order with sorted order
	for i := range items {
		propOriginal := items[i].(*Property)
		propSorted := sorted[i].(*Property)
		if propOriginal.SortKey != propSorted.SortKey {
			return false
		}
		// For deprecated-at-end, also check if deprecated properties are in the right place
		if deprecatedAtEnd && propOriginal.isDeprecated != propSorted.isDeprecated {
			return false
		}
	}

	return true
}

// GetMagicCommentIndex returns the index of the magic comment
func (o *ObjectSorter) GetMagicCommentIndex() int {
	return o.magicIndex
}

// GetNode returns the underlying AST node
func (o *ObjectSorter) GetNode() *sitter.Node {
	return o.node
}

