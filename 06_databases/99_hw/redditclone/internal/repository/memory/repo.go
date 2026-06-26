package memory

import (
	"sync"

	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/repository"
)

type Repository struct {
	UserRepo     *UserRepo
	PostsRepo    *PostsRepo
	CommentsRepo *CommentsRepository
}

type AnyRepo[T any] struct {
	data   T
	mu     *sync.RWMutex
	nextID uint64
}

func NewRepository() repository.Repositories {
	return repository.Repositories{
		User:  NewUserRepository(),
		Posts: NewPostsRepository(),
	}
}
