package arrays

import (
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"

	sitter "github.com/smacker/go-tree-sitter"
)

// Element represents an array element that can be sorted
type Element struct {
	Node         *sitter.Node
	BeforeNodes  []*sitter.Node // Comments before this element
	AfterNode    *sitter.Node   // Inline comment after element
	HasComma     bool
	CommaNode    *sitter.Node
	SortKey      string // The extracted key for sorting
	isDeprecated bool
}

// GetSortKey returns the key for sorting based on the strategy
func (e *Element) GetSortKey(strategy interfaces.SortStrategy, content []byte) (string, error) {
	// For array elements, delegate to the strategy to extract the key
	return strategy.ExtractKey(e, content)
}

// IsDeprecated returns true if this element has @deprecated annotation
func (e *Element) IsDeprecated() bool {
	return e.isDeprecated
}

// GetNode returns the underlying AST node
func (e *Element) GetNode() *sitter.Node {
	return e.Node
}

// GetBeforeComments returns comments that appear before this element
func (e *Element) GetBeforeComments() []*sitter.Node {
	return e.BeforeNodes
}

// GetAfterComment returns inline comment that appears after this element
func (e *Element) GetAfterComment() *sitter.Node {
	return e.AfterNode
}

// NewElement creates a new Element from an AST node
func NewElement(node *sitter.Node) *Element {
	return &Element{
		Node: node,
	}
}