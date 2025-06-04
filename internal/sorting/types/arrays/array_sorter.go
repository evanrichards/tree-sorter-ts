package arrays

import (
	"sort"
	"strings"

	"github.com/evanrichards/tree-sorter-ts/internal/sorting/common"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"

	sitter "github.com/smacker/go-tree-sitter"
)

// ArraySorter handles sorting of array elements
type ArraySorter struct {
	node         *sitter.Node
	magicComment *sitter.Node
	magicIndex   int
}

// NewArraySorter creates a new array sorter
func NewArraySorter(arrayNode, magicComment *sitter.Node, magicIndex int) *ArraySorter {
	return &ArraySorter{
		node:         arrayNode,
		magicComment: magicComment,
		magicIndex:   magicIndex,
	}
}

// Extract finds and extracts sortable elements from the array
func (a *ArraySorter) Extract(node *sitter.Node, content []byte) ([]interfaces.SortableItem, error) {
	var elements []interfaces.SortableItem
	var pendingComments []*sitter.Node

	// Start after magic comment
	startIdx := a.magicIndex + 1

	for i := startIdx; i < int(a.node.ChildCount()); i++ {
		child := a.node.Child(i)

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
			elem := NewElement(child)
			elem.BeforeNodes = pendingComments

			// Check if this element has @deprecated annotation
			elem.isDeprecated = common.HasDeprecatedAnnotation(pendingComments, content)

			// Check if followed by comma and/or inline comment
			j := i + 1
			continueLoop := true
			for j < int(a.node.ChildCount()) && continueLoop {
				next := a.node.Child(j)
				switch next.Type() {
				case ",":
					elem.HasComma = true
					elem.CommaNode = next
					j++
				case "comment":
					// Check if it's on the same line
					if next.StartPoint().Row == child.EndPoint().Row {
						elem.AfterNode = next
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
			if !elem.isDeprecated && elem.AfterNode != nil {
				text := string(content[elem.AfterNode.StartByte():elem.AfterNode.EndByte()])
				if strings.Contains(text, "@deprecated") {
					elem.isDeprecated = true
				}
			}

			elements = append(elements, elem)
			pendingComments = nil // Reset comments
		}
	}

	return elements, nil
}

// Sort applies the strategy to sort the elements
func (a *ArraySorter) Sort(items []interfaces.SortableItem, strategy interfaces.SortStrategy, deprecatedAtEnd bool, content []byte) ([]interfaces.SortableItem, error) {
	if len(items) <= 1 {
		return items, nil
	}

	// Extract sort keys for each element
	for _, item := range items {
		elem := item.(*Element)
		sortKey, err := item.GetSortKey(strategy, content)
		if err != nil {
			// For missing/invalid keys, mark with special prefix to sort last
			elem.SortKey = "\uffff" + "unknown"
		} else {
			elem.SortKey = sortKey
		}
	}

	// Make a copy for sorting
	sorted := make([]interfaces.SortableItem, len(items))
	copy(sorted, items)

	// Sort elements, considering deprecated-at-end flag
	if deprecatedAtEnd {
		sort.Slice(sorted, func(i, j int) bool {
			elemI := sorted[i].(*Element)
			elemJ := sorted[j].(*Element)
			// If one is deprecated and the other isn't, put non-deprecated first
			if elemI.isDeprecated != elemJ.isDeprecated {
				return !elemI.isDeprecated
			}
			// Use string comparison to ensure \uffff prefix works for missing keys
			return elemI.SortKey < elemJ.SortKey
		})
	} else {
		sort.Slice(sorted, func(i, j int) bool {
			elemI := sorted[i].(*Element)
			elemJ := sorted[j].(*Element)
			// Check if either has missing key prefix
			iHasMissingKey := strings.HasPrefix(elemI.SortKey, "\uffff")
			jHasMissingKey := strings.HasPrefix(elemJ.SortKey, "\uffff")

			// If one has missing key and other doesn't, sort missing key last
			if iHasMissingKey != jHasMissingKey {
				return !iHasMissingKey
			}

			// If both have or don't have missing keys, compare normally
			if iHasMissingKey && jHasMissingKey {
				// Both have missing keys, compare without prefix
				return elemI.SortKey < elemJ.SortKey
			}

			// Neither has missing key, use compareKeys for proper type handling
			return common.CompareKeys(elemI.SortKey, elemJ.SortKey)
		})
	}

	return sorted, nil
}

// CheckIfSorted determines if elements are already sorted according to strategy
func (a *ArraySorter) CheckIfSorted(items []interfaces.SortableItem, strategy interfaces.SortStrategy, deprecatedAtEnd bool, content []byte) bool {
	if len(items) <= 1 {
		return true
	}

	sorted, err := a.Sort(items, strategy, deprecatedAtEnd, content)
	if err != nil {
		return false
	}

	// Compare original order with sorted order by comparing nodes
	for i := range items {
		if items[i].GetNode() != sorted[i].GetNode() {
			return false
		}
	}

	return true
}

// GetMagicCommentIndex returns the index of the magic comment
func (a *ArraySorter) GetMagicCommentIndex() int {
	return a.magicIndex
}

// GetNode returns the underlying AST node
func (a *ArraySorter) GetNode() *sitter.Node {
	return a.node
}