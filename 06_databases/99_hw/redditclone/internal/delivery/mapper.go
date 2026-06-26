package delivery

import (
	"time"

	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/domain"
)

type UserResponse struct {
	Username string `json:"username"`
	ID       int64  `json:"id"`
}

type CommentResponse struct {
	Body      string       `json:"body"`
	ID        uint64       `json:"id"`
	CreatedAt string       `json:"created_at"`
	Author    UserResponse `json:"author"`
}

type VoteResponse struct {
	UserID int64 `json:"user_id"`
	Vote   int8  `json:"vote"`
}

type PostDTO struct {
	Title    string `json:"title" validate:"required"`
	Text     string `json:"text" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=text link"`
	Category string `json:"category" validate:"required,oneof=music funny videos programming news"`
	URL      string `json:"url"`
}

type PostResponse struct {
	ID               uint64            `json:"id"`
	Author           UserResponse      `json:"author"`
	CreatedAt        time.Time         `json:"created_at"`
	Score            int64             `json:"score"`
	UpvotePercentage int8              `json:"upvotePercentage"`
	Views            uint64            `json:"views"`
	Comments         []CommentResponse `json:"comments"`
	Votes            []VoteResponse    `json:"votes"`
	Title            string            `json:"title"`
	Text             string            `json:"text"`
	Type             string            `json:"type"`
	Category         string            `json:"category"`
	URL              string            `json:"url"`
}

func toPostResponse(inp domain.Post) PostResponse {
	var comments []CommentResponse
	for _, comment := range inp.Comments {
		comments = append(comments, CommentResponse{
			Body:      comment.Body,
			CreatedAt: comment.CreatedAt.Format(time.DateTime),
			ID:        comment.ID,
			Author: UserResponse{
				ID:       comment.Author.ID,
				Username: comment.Author.Username,
			},
		})
	}
	var votes []VoteResponse
	for _, vote := range inp.Votes {
		votes = append(votes, VoteResponse{
			UserID: vote.UserID,
			Vote:   vote.Vote,
		})
	}
	return PostResponse{
		ID: inp.ID,
		Author: UserResponse{
			ID:       inp.Author.ID,
			Username: inp.Author.Username,
		},
		CreatedAt:        inp.CreatedAt,
		Score:            inp.Score,
		UpvotePercentage: inp.UpvotePercentage,
		Views:            inp.Views,
		Comments:         comments,
		Votes:            votes,
		Title:            inp.Title,
		Text:             inp.Text,
		Type:             string(inp.Type),
		Category:         inp.Category,
		URL:              inp.URL,
	}
}

type UserDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func toUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:       user.ID,
		Username: user.Username,
	}
}
