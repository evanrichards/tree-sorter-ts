package reconstruction

import (
	"github.com/evanrichards/tree-sorter-ts/internal/config"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"
)

// Reconstructor rebuilds AST content with sorted items
type Reconstructor interface {
	// Reconstruct generates new content with sorted items
	Reconstruct(sortable interfaces.Sortable, sortedItems []interfaces.SortableItem, config config.SortConfig, content []byte) ([]byte, error)
}