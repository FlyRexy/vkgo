package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/domain"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/repository/postgres"
	redisRepo "gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/repository/redis"
)

type PostsRepository interface {
	CreatePost(ctx context.Context, post domain.Post) (domain.Post, error)
	GetAllPosts(ctx context.Context) ([]domain.Post, error)
	GetPostByCategory(ctx context.Context, category string) ([]domain.Post, error)
	GetPostByID(ctx context.Context, postID uint64) (domain.Post, error)
	GetPostsByUsername(ctx context.Context, username string) ([]domain.Post, error)

	// votes
	Vote(ctx context.Context, postID uint64, userID int64, amount int8) (domain.Post, error)
	DiscardVote(ctx context.Context, postID uint64, userID int64) (domain.Post, error)

	// comments
	AddCommentToPost(ctx context.Context, id uint64, comment domain.Comment) error
	DeleteCommentFromPost(ctx context.Context, postID, commentID, userID uint64) error
}

type UserRepository interface {
	Login(ctx context.Context, user domain.User) (domain.User, error)
	Signup(ctx context.Context, user domain.User) (domain.User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, user domain.User) (string, error)
}

type Repositories struct {
	Posts   PostsRepository
	User    UserRepository
	Session SessionRepository
}

func NewRepositories(db *pgxpool.Pool, redis *redis.Client) *Repositories {
	return &Repositories{
		Posts:   postgres.NewPostsPGRepository(db),
		User:    postgres.NewUsersRepository(db),
		Session: redisRepo.NewSessionRepository(redis),
	}
}
