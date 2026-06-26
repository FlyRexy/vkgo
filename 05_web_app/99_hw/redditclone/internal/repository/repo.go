package repository

import (
	"context"
	"sync"

	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/domain"
)

type Repository struct {
	UserRepo     *UserRepo
	PostsRepo    *PostsRepo
	CommentsRepo *CommentsRepository
}

type PostsRepository interface {
	CreatePost(ctx context.Context, post domain.Post) (domain.Post, error)
	GetAllPosts(ctx context.Context) ([]domain.Post, error)
	GetPostByCategory(ctx context.Context, category string) ([]domain.Post, error)
	GetPostByID(ctx context.Context, postID uint64) (domain.Post, error)
	GetUserVote(ctx context.Context, postID uint64, userID int64) (domain.Vote, bool)
	ChangeVoteValue(ctx context.Context, postID uint64, userID int64, amount int8) (domain.Post, error)
	AppendVote(ctx context.Context, postID uint64, vote domain.Vote) (domain.Post, error)
	GetPostsByUsername(ctx context.Context, username string) ([]domain.Post, error)
	AddCommentToPost(ctx context.Context, id uint64, comment domain.Comment) (domain.Post, error)
	DeleteCommentFromPost(ctx context.Context, postID, commentID uint64) (domain.Post, error)
}

type UserRepository interface {
	Login(ctx context.Context, user domain.User) (domain.User, error)
	Signup(ctx context.Context, user domain.User) (domain.User, error)
}

type CommentRepository interface {
	CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	DeleteComment(commentID uint64)
	GetComment(commentID uint64) (domain.Comment, bool)
}

type Repositories struct {
	Posts   PostsRepository
	User    UserRepository
	Comment CommentRepository
}

type AnyRepo[T any] struct {
	data   T
	mu     *sync.RWMutex
	nextID uint64
}

func NewRepository() Repositories {
	return Repositories{
		User:    NewUserRepository(),
		Posts:   NewPostsRepository(),
		Comment: NewCommentsRepository(),
	}
}
