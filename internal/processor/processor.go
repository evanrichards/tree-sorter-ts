package processor

import (
	"context"
	"fmt"

	"github.com/evanrichards/tree-sorter-ts/internal/config"
	"github.com/evanrichards/tree-sorter-ts/internal/parser"
	"github.com/evanrichards/tree-sorter-ts/internal/reconstruction"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/strategies"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// Processor handles the complete sorting workflow for TypeScript/TSX files
type Processor struct {
	astParser             *sitter.Parser
	strategyFactory       *strategies.Factory
	reconstructionFactory *reconstruction.Factory
}

// NewProcessor creates a new processor with all dependencies
func NewProcessor() *Processor {
	astParser := sitter.NewParser()
	astParser.SetLanguage(typescript.GetLanguage())

	return &Processor{
		astParser:             astParser,
		strategyFactory:       strategies.NewFactory(),
		reconstructionFactory: reconstruction.NewFactory(),
	}
}

// ProcessContent processes TypeScript/TSX content and returns sorted result
func (p *Processor) ProcessContent(content []byte) ([]byte, error) {
	// Parse AST
	tree, err := p.astParser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	// Start with original content
	result := make([]byte, len(content))
	copy(result, content)

	// Process all objects with magic comments
	objects, err := parser.FindObjectsWithMagicComments(tree.RootNode(), content)
	if err != nil {
		return nil, fmt.Errorf("failed to find objects: %w", err)
	}

	for _, objectSorter := range objects {
		if updated, err := p.processSortable(objectSorter, result); err != nil {
			return nil, fmt.Errorf("failed to process object: %w", err)
		} else {
			result = updated
		}
	}

	// Process all arrays with magic comments
	arrays, err := parser.FindArraysWithMagicComments(tree.RootNode(), content)
	if err != nil {
		return nil, fmt.Errorf("failed to find arrays: %w", err)
	}

	for _, arraySorter := range arrays {
		if updated, err := p.processSortable(arraySorter, result); err != nil {
			return nil, fmt.Errorf("failed to process array: %w", err)
		} else {
			result = updated
		}
	}

	return result, nil
}

// processSortable handles the complete sorting workflow for a single sortable structure
func (p *Processor) processSortable(sortable interfaces.Sortable, content []byte) ([]byte, error) {
	// Extract the magic comment to get configuration
	cfg, err := p.extractConfig(sortable, content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Extract sortable items
	items, err := sortable.Extract(sortable.GetNode(), content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract items: %w", err)
	}

	// Get appropriate sorting strategy
	strategy, err := p.strategyFactory.CreateStrategy(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create strategy: %w", err)
	}

	// Check if already sorted
	if sortable.CheckIfSorted(items, strategy, cfg.DeprecatedAtEnd, content) {
		// Already sorted, no changes needed
		return content, nil
	}

	// Sort the items
	sortedItems, err := sortable.Sort(items, strategy, cfg.DeprecatedAtEnd, content)
	if err != nil {
		return nil, fmt.Errorf("failed to sort items: %w", err)
	}

	// Get appropriate reconstructor
	reconstructor, err := p.reconstructionFactory.CreateReconstructor(sortable)
	if err != nil {
		return nil, fmt.Errorf("failed to create reconstructor: %w", err)
	}

	// Reconstruct the content
	reconstructed, err := reconstructor.Reconstruct(sortable, sortedItems, cfg, content)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct content: %w", err)
	}

	return reconstructed, nil
}

// extractConfig extracts configuration from the magic comment
func (p *Processor) extractConfig(sortable interfaces.Sortable, content []byte) (config.SortConfig, error) {
	// Find the magic comment node
	node := sortable.GetNode()
	magicIndex := sortable.GetMagicCommentIndex()

	if magicIndex < 0 || magicIndex >= int(node.ChildCount()) {
		return config.SortConfig{}, fmt.Errorf("invalid magic comment index: %d", magicIndex)
	}

	magicComment := node.Child(magicIndex)
	if magicComment.Type() != "comment" {
		return config.SortConfig{}, fmt.Errorf("expected comment node, got %s", magicComment.Type())
	}

	// Extract comment text
	commentText := content[magicComment.StartByte():magicComment.EndByte()]

	// Parse configuration
	cfg := config.ParseSortConfig(commentText)
	return cfg, nil
}

// ProcessFile is a convenience method that reads, processes, and could write back a file
func (p *Processor) ProcessFile(filename string, content []byte) ([]byte, error) {
	return p.ProcessContent(content)
}