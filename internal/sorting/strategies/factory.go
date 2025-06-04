package strategies

import (
	"github.com/evanrichards/tree-sorter-ts/internal/config"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"
)

// Factory creates sorting strategies based on configuration
type Factory struct{}

// CreateStrategy creates the appropriate strategy based on config
func (f *Factory) CreateStrategy(cfg config.SortConfig) (interfaces.SortStrategy, error) {
	if cfg.SortByComment {
		return &CommentContentStrategy{}, nil
	}
	
	if cfg.Key != "" {
		return &ArrayKeyStrategy{KeyPath: cfg.Key}, nil
	}
	
	return &PropertyNameStrategy{}, nil
}

// NewFactory creates a new strategy factory
func NewFactory() *Factory {
	return &Factory{}
}