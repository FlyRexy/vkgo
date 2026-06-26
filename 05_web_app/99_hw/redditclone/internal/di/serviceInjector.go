package di

import (
	"context"

	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/domain"
	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/lib"
	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/repository"
	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/service"
)

type PostsService interface {
	CreatePost(ctx context.Context, postDTO service.CreatePostDTO) (domain.Post, error)
	GetAllPosts(ctx context.Context) ([]domain.Post, error)
	GetPostByCategory(ctx context.Context, category string) ([]domain.Post, error)
	GetPostByID(ctx context.Context, postID uint64) (domain.Post, error)
	GetPostsByUsername(ctx context.Context, username string) ([]domain.Post, error)
	Upvote(ctx context.Context, postID uint64) (domain.Post, error)
	Downvote(ctx context.Context, postID uint64) (domain.Post, error)
	DiscardVote(ctx context.Context, postID uint64) (domain.Post, error)
	AddCommentToPost(ctx context.Context, postID uint64, body string) (domain.Post, error)
	DeleteComment(ctx context.Context, postID uint64, commentID uint64) (domain.Post, error)
}

type AuthService interface {
	Login(ctx context.Context, userDTO service.UserDTO) (domain.User, lib.Token, error)
	Signup(ctx context.Context, userDTO service.UserDTO) (domain.User, lib.Token, error)
}

type ServiceInjector struct {
	PostsService PostsService
	AuthService  AuthService
}

func NewServiceInjector(repo repository.Repositories) ServiceInjector {
	return ServiceInjector{
		PostsService: service.NewPostsService(repo.Posts, repo.Comment),
		AuthService:  service.NewAuthService(repo.User),
	}
}
