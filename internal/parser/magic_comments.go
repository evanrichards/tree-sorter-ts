package parser

import (
	"regexp"

	"github.com/evanrichards/tree-sorter-ts/internal/config"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/types/arrays"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/types/objects"

	sitter "github.com/smacker/go-tree-sitter"
)

var (
	magicCommentRegex = regexp.MustCompile(`(?s)/\*\*?.*?tree-sorter-ts:\s*keep-sorted\b.*?\*/`)
)

// FindObjectsWithMagicComments finds all objects containing magic comments
func FindObjectsWithMagicComments(node *sitter.Node, content []byte) ([]*objects.ObjectSorter, error) {
	var results []*objects.ObjectSorter

	var traverse func(*sitter.Node)
	traverse = func(n *sitter.Node) {
		if n.Type() == "object" {
			// Check children for magic comment
			for i := 0; i < int(n.ChildCount()); i++ {
				child := n.Child(i)
				if child.Type() == "comment" {
					text := content[child.StartByte():child.EndByte()]
					if magicCommentRegex.Match(text) {
						cfg := config.ParseSortConfig(text)
						if err := cfg.Validate(); err != nil {
							// Skip objects with invalid config
							continue
						}
						results = append(results, objects.NewObjectSorter(n, child, i))
						break
					}
				}
			}
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			traverse(n.Child(i))
		}
	}

	traverse(node)
	return results, nil
}

// FindArraysWithMagicComments finds all arrays containing magic comments
func FindArraysWithMagicComments(node *sitter.Node, content []byte) ([]*arrays.ArraySorter, error) {
	var results []*arrays.ArraySorter

	var traverse func(*sitter.Node)
	traverse = func(n *sitter.Node) {
		if n.Type() == "array" {
			// Check children for magic comment
			for i := 0; i < int(n.ChildCount()); i++ {
				child := n.Child(i)
				if child.Type() == "comment" {
					text := content[child.StartByte():child.EndByte()]
					if magicCommentRegex.Match(text) {
						cfg := config.ParseSortConfig(text)
						if err := cfg.Validate(); err != nil {
							// Skip arrays with invalid config
							continue
						}
						results = append(results, arrays.NewArraySorter(n, child, i))
						break
					}
				}
			}
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			traverse(n.Child(i))
		}
	}

	traverse(node)
	return results, nil
}