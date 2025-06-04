package interfaces

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// SortableItem represents an item that can be sorted (property, array element, parameter)
type SortableItem interface {
	// GetSortKey returns the key used for sorting based on the strategy
	GetSortKey(strategy SortStrategy, content []byte) (string, error)
	
	// IsDeprecated returns true if this item has @deprecated annotation
	IsDeprecated() bool
	
	// GetNode returns the underlying AST node
	GetNode() *sitter.Node
	
	// GetBeforeComments returns comments that appear before this item
	GetBeforeComments() []*sitter.Node
	
	// GetAfterComment returns inline comment that appears after this item
	GetAfterComment() *sitter.Node
}

// SortStrategy defines how to extract sort keys from items
type SortStrategy interface {
	// ExtractKey extracts the sorting key from a sortable item
	ExtractKey(item SortableItem, content []byte) (string, error)
	
	// GetName returns the strategy name for debugging
	GetName() string
}

// Sortable represents a structure that can be sorted (object, array, constructor)
type Sortable interface {
	// Extract finds and extracts sortable items from the AST node
	Extract(node *sitter.Node, content []byte) ([]SortableItem, error)
	
	// Sort applies the strategy to sort the items
	Sort(items []SortableItem, strategy SortStrategy, deprecatedAtEnd bool, content []byte) ([]SortableItem, error)
	
	// CheckIfSorted determines if items are already sorted according to strategy
	CheckIfSorted(items []SortableItem, strategy SortStrategy, deprecatedAtEnd bool, content []byte) bool
	
	// GetMagicCommentIndex returns the index of the magic comment
	GetMagicCommentIndex() int
	
	// GetNode returns the underlying AST node
	GetNode() *sitter.Node
}

// Reconstructor rebuilds AST content with sorted items
type Reconstructor interface {
	// Reconstruct generates new content with sorted items
	Reconstruct(sortable Sortable, sortedItems []SortableItem, config interface{}, content []byte) ([]byte, error)
}

// SortConfig contains configuration options from the magic comment
type SortConfig struct {
	WithNewLine     bool
	DeprecatedAtEnd bool
	Key             string // For array sorting
	SortByComment   bool   // Sort by comment content
	HasError        bool   // Indicates a validation error
}

// StrategyFactory creates sorting strategies
type StrategyFactory interface {
	// CreateStrategy creates a strategy based on config
	CreateStrategy(config SortConfig) (SortStrategy, error)
}