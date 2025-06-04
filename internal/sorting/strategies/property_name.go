package strategies

import (
	"strings"

	"github.com/evanrichards/tree-sorter-ts/internal/sorting/common"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/types/arrays"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/types/objects"
)

// PropertyNameStrategy sorts by property/element name
type PropertyNameStrategy struct{}

func (s *PropertyNameStrategy) ExtractKey(item interfaces.SortableItem, content []byte) (string, error) {
	switch typedItem := item.(type) {
	case *objects.Property:
		// For object properties, return the property key
		return typedItem.Key, nil
	case *arrays.Element:
		// For array elements, extract value as string
		nodeText := strings.TrimSpace(string(content[typedItem.GetNode().StartByte():typedItem.GetNode().EndByte()]))
		return common.TrimQuotes(nodeText), nil
	default:
		// Fallback for other types
		nodeText := strings.TrimSpace(string(content[item.GetNode().StartByte():item.GetNode().EndByte()]))
		return common.TrimQuotes(nodeText), nil
	}
}

func (s *PropertyNameStrategy) GetName() string {
	return "property-name"
}