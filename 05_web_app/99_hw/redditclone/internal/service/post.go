package service

import (
	"context"
	"errors"
	"time"

	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/domain"
	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/lib"
	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/repository"
)

type CreatePostDTO struct {
	Title    string
	Text     string
	Type     domain.PostType
	Category string
	URL      string
}

type PostsService struct {
	postsRepo    repository.PostsRepository
	commentsRepo repository.CommentRepository
}

func NewPostsService(postsRepo repository.PostsRepository, commentsRepo repository.CommentRepository) PostsService {
	return PostsService{
		postsRepo,
		commentsRepo,
	}
}

func (ps PostsService) CreatePost(ctx context.Context, postParams CreatePostDTO) (domain.Post, error) {
	user, err := lib.GetUserFromContext(ctx)
	if err != nil {
		return domain.Post{}, err
	}

	post, err := ps.postsRepo.CreatePost(ctx, domain.Post{
		Title:     postParams.Title,
		Text:      postParams.Text,
		Category:  postParams.Category,
		Type:      postParams.Type,
		URL:       postParams.URL,
		Author:    user,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return domain.Post{}, errors.Join(errors.New("got error from repository"), err)
	}

	return post, nil
}

func (ps PostsService) GetAllPosts(ctx context.Context) ([]domain.Post, error) {
	return ps.postsRepo.GetAllPosts(ctx)
}

func (ps PostsService) GetPostByCategory(ctx context.Context, category string) ([]domain.Post, error) {
	return ps.postsRepo.GetPostByCategory(ctx, category)
}

func (ps PostsService) GetPostByID(ctx context.Context, postID uint64) (domain.Post, error) {
	return ps.postsRepo.GetPostByID(ctx, postID)
}
func (ps PostsService) GetPostsByUsername(ctx context.Context, username string) ([]domain.Post, error) {
	return ps.postsRepo.GetPostsByUsername(ctx, username)
}

func (ps PostsService) Upvote(ctx context.Context, postID uint64) (domain.Post, error) {
	user, err := lib.GetUserFromContext(ctx)
	if err != nil {
		return domain.Post{}, err
	}

	vote, has := ps.postsRepo.GetUserVote(ctx, postID, user.ID)
	if has && vote.Vote == 1 {
		return ps.postsRepo.GetPostByID(ctx, postID)
	}

	if has && vote.Vote == -1 {
		return ps.postsRepo.ChangeVoteValue(ctx, postID, vote.UserID, 2)
	}

	return ps.postsRepo.AppendVote(ctx, postID, domain.Vote{
		UserID: user.ID,
		Vote:   1,
	})
}

func (ps PostsService) Downvote(ctx context.Context, postID uint64) (domain.Post, error) {
	user, err := lib.GetUserFromContext(ctx)
	if err != nil {
		return domain.Post{}, err
	}

	vote, has := ps.postsRepo.GetUserVote(ctx, postID, user.ID)
	if has && vote.Vote == -1 {
		return ps.postsRepo.GetPostByID(ctx, postID)
	}

	if has && vote.Vote == 1 {
		return ps.postsRepo.ChangeVoteValue(ctx, postID, vote.UserID, -2)
	}

	return ps.postsRepo.AppendVote(ctx, postID, domain.Vote{
		UserID: user.ID,
		Vote:   -1,
	})
}

func (ps PostsService) DiscardVote(ctx context.Context, postID uint64) (domain.Post, error) {
	user, err := lib.GetUserFromContext(ctx)
	if err != nil {
		return domain.Post{}, err
	}

	vote, has := ps.postsRepo.GetUserVote(ctx, postID, user.ID)
	if !has {
		return domain.Post{}, errors.New("this user doesn't have any votes")
	}

	return ps.postsRepo.ChangeVoteValue(ctx, postID, vote.UserID, -vote.Vote)
}

func (ps PostsService) AddCommentToPost(ctx context.Context, postID uint64, body string) (domain.Post, error) {
	user, err := lib.GetUserFromContext(ctx)
	if err != nil {
		return domain.Post{}, err
	}

	comment, err := ps.commentsRepo.CreateComment(ctx, domain.Comment{
		Body:      body,
		CreatedAt: time.Now(),
		Author:    user,
	})
	if err != nil {
		return domain.Post{}, err
	}
	return ps.postsRepo.AddCommentToPost(ctx, postID, comment)

}

func (ps PostsService) DeleteComment(ctx context.Context, postID uint64, commentID uint64) (domain.Post, error) {
	user, err := lib.GetUserFromContext(ctx)
	if err != nil {
		return domain.Post{}, err
	}
	comment, ok := ps.commentsRepo.GetComment(commentID)
	if !ok {
		return domain.Post{}, errors.New("comment not found")
	}

	if comment.Author.ID != user.ID {
		return domain.Post{}, errors.New("current user doesn't have access to deleting comment")
	}

	ps.commentsRepo.DeleteComment(commentID)
	return ps.postsRepo.DeleteCommentFromPost(ctx, postID, commentID)
}
