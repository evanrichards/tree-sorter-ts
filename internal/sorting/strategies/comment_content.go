package strategies

import (
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/common"
	"github.com/evanrichards/tree-sorter-ts/internal/sorting/interfaces"
)

// CommentContentStrategy sorts by comment content
type CommentContentStrategy struct{}

func (s *CommentContentStrategy) ExtractKey(item interfaces.SortableItem, content []byte) (string, error) {
	return common.FindCommentTextForSorting(
		item.GetBeforeComments(),
		item.GetAfterComment(),
		content,
	)
}

func (s *CommentContentStrategy) GetName() string {
	return "comment-content"
}