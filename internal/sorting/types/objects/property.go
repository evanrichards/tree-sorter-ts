package objects

import (
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/common"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"

	sitter "github.com/smacker/go-tree-sitter"
)

// Property represents an object property that can be sorted
type Property struct {
	KeyNode      *sitter.Node
	ValueNode    *sitter.Node
	PairNode     *sitter.Node
	Key          string
	SortKey      string          // The key used for sorting (may be different from key when using sort-by-comment)
	BeforeNodes  []*sitter.Node // Comments before this property
	AfterNode    *sitter.Node   // Inline comment after property
	HasComma     bool
	CommaNode    *sitter.Node
	isDeprecated bool // Whether this property has @deprecated annotation
}

// GetSortKey returns the key for sorting based on the strategy
func (p *Property) GetSortKey(strategy interfaces.SortStrategy, content []byte) (string, error) {
	// For property name strategy, return the property key
	if strategy.GetName() == "property-name" {
		return p.Key, nil
	}
	
	// For other strategies, delegate to the strategy
	return strategy.ExtractKey(p, content)
}

// IsDeprecated returns true if this property has @deprecated annotation
func (p *Property) IsDeprecated() bool {
	return p.isDeprecated
}

// GetNode returns the underlying AST node
func (p *Property) GetNode() *sitter.Node {
	return p.PairNode
}

// GetBeforeComments returns comments that appear before this property
func (p *Property) GetBeforeComments() []*sitter.Node {
	return p.BeforeNodes
}

// GetAfterComment returns inline comment that appears after this property
func (p *Property) GetAfterComment() *sitter.Node {
	return p.AfterNode
}

// NewProperty creates a new Property from AST nodes
func NewProperty(pairNode *sitter.Node, content []byte) *Property {
	prop := &Property{
		PairNode: pairNode,
	}
	
	// Extract key and value
	keyNode := pairNode.ChildByFieldName("key")
	valueNode := pairNode.ChildByFieldName("value")
	
	if keyNode != nil {
		prop.KeyNode = keyNode
		prop.Key = common.ExtractKeyFromNode(keyNode, content)
	}
	
	if valueNode != nil {
		prop.ValueNode = valueNode
	}
	
	return prop
}