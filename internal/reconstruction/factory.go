package reconstruction

import (
	"fmt"

	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"
)

// Factory creates appropriate reconstructors for different sortable types
type Factory struct {
	reconstructors []interfaces.Reconstructor
}

// NewFactory creates a new reconstruction factory
func NewFactory() *Factory {
	return &Factory{
		reconstructors: []interfaces.Reconstructor{
			NewObjectReconstructor(),
			NewArrayReconstructor(),
		},
	}
}

// CreateReconstructor returns the appropriate reconstructor for the given sortable
func (f *Factory) CreateReconstructor(sortable interfaces.Sortable) (interfaces.Reconstructor, error) {
	for _, reconstructor := range f.reconstructors {
		// Check if reconstructor can handle this type
		switch r := reconstructor.(type) {
		case *ObjectReconstructor:
			if r.CanHandle(sortable) {
				return r, nil
			}
		case *ArrayReconstructor:
			if r.CanHandle(sortable) {
				return r, nil
			}
		}
	}
	
	return nil, fmt.Errorf("no reconstructor found for sortable type %T", sortable)
}

// GetSupportedTypes returns the types of sortables this factory supports
func (f *Factory) GetSupportedTypes() []string {
	return []string{"objects.ObjectSorter", "arrays.ArraySorter"}
}