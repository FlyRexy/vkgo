package repository

import (
	"context"
	"sync"
	"time"

	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/domain"
)

type CommentEntity struct {
	Body      string
	ID        uint64
	CreatedAt string
	Author    UserEntity
}

type CommentsRepository AnyRepo[[]CommentEntity]

func NewCommentsRepository() *CommentsRepository {
	return &CommentsRepository{
		data:   make([]CommentEntity, 0),
		mu:     &sync.RWMutex{},
		nextID: 1,
	}
}

func (c *CommentsRepository) CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	c.mu.Lock()
	newComment := CommentEntity{
		ID:        c.nextID,
		CreatedAt: comment.CreatedAt.Format(time.DateTime),
		Body:      comment.Body,
		Author:    UserEntity(comment.Author),
	}
	c.nextID++
	c.data = append(c.data, newComment)
	c.mu.Unlock()

	createdAt, err := time.Parse(time.DateTime, newComment.CreatedAt)
	if err != nil {
		return domain.Comment{}, err
	}

	return domain.Comment{
		ID:        newComment.ID,
		CreatedAt: createdAt,
		Body:      newComment.Body,
		Author:    domain.User(newComment.Author),
	}, nil
}

func (c *CommentsRepository) DeleteComment(commentID uint64) {
	c.mu.Lock()
	for idx, comment := range c.data {
		if comment.ID == commentID {
			c.data = append(c.data[:idx], c.data[idx+1:]...)
			break
		}
	}
	c.mu.Unlock()
}

func (c *CommentsRepository) GetComment(commentID uint64) (domain.Comment, bool) {
	c.mu.RLock()
	for _, comment := range c.data {
		if comment.ID == commentID {
			createdAt, err := time.Parse(time.DateTime, comment.CreatedAt)
			if err != nil {
				return domain.Comment{}, false
			}
			return domain.Comment{
				ID:        comment.ID,
				CreatedAt: createdAt,
				Body:      comment.Body,
				Author:    domain.User(comment.Author),
			}, true
		}
	}
	c.mu.RUnlock()
	return domain.Comment{}, false
}
