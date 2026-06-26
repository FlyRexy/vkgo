package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/domain"
)

type PostsRepo struct {
	AnyRepo[[]domain.Post]
	nextID uint64
}

func NewPostsRepository() *PostsRepo {
	return &PostsRepo{
		nextID: 1,
		AnyRepo: AnyRepo[[]domain.Post]{
			mu:   &sync.RWMutex{},
			data: make([]domain.Post, 0),
		},
	}
}

func (p *PostsRepo) CreatePost(ctx context.Context, post domain.Post) (domain.Post, error) {
	p.mu.Lock()
	newPost := domain.Post{
		ID:        p.nextID,
		CreatedAt: post.CreatedAt,
		Title:     post.Title,
		Text:      post.Text,
		Category:  post.Category,
		Type:      post.Type,
		URL:       post.URL,
		Author:    post.Author,
	}
	p.data = append(p.data, newPost)
	p.nextID++
	p.mu.Unlock()

	return newPost, nil
}

func (p *PostsRepo) GetPostByID(ctx context.Context, postID uint64) (domain.Post, error) {
	var post domain.Post

	p.mu.RLock()
	for _, curPost := range p.data {
		if curPost.ID == postID {
			post = curPost
		}
	}
	p.mu.RUnlock()
	if post.ID == 0 {
		return domain.Post{}, fmt.Errorf("post with id %d doesn't exist", postID)
	}

	return post, nil
}

func (p *PostsRepo) GetAllPosts(ctx context.Context) ([]domain.Post, error) {
	p.mu.RLock()
	posts := p.data
	p.mu.RUnlock()
	return posts, nil
}

func (p *PostsRepo) GetPostByCategory(ctx context.Context, cat string) ([]domain.Post, error) {
	var suitablePosts []domain.Post
	p.mu.RLock()
	for _, post := range p.data {
		if post.Category == cat {
			suitablePosts = append(suitablePosts, post)
		}
	}
	p.mu.RUnlock()

	return suitablePosts, nil
}

func (p *PostsRepo) GetUserVote(ctx context.Context, postID uint64, userID int64) (domain.Vote, bool) {
	post, err := p.GetPostByID(ctx, postID)
	if err != nil {
		return domain.Vote{}, false
	}

	for _, vote := range post.Votes {
		if vote.UserID == userID {
			return vote, true
		}
	}

	return domain.Vote{}, false
}

func (p *PostsRepo) ChangeVoteValue(ctx context.Context, postID uint64, userID int64, amount int8) (domain.Post, error) {
	p.mu.Lock()
	for idx, curPost := range p.data {
		if curPost.ID == postID {
			for voteIdx := range p.data[idx].Votes {
				if p.data[idx].Votes[voteIdx].UserID == userID {
					p.data[idx].Votes[voteIdx].Vote += amount
					p.data[idx].Score += int64(amount)
					p.data[idx].UpvotePercentage = int8(p.data[idx].Score / int64(len(p.data[idx].Votes)) * 100)
					p.mu.Unlock()

					return p.data[idx], nil
				}
			}
		}
	}
	p.mu.Unlock()

	return domain.Post{}, errors.New("no post for the vote")
}

func (p *PostsRepo) AppendVote(ctx context.Context, postID uint64, vote domain.Vote) (domain.Post, error) {
	p.mu.Lock()
	for idx, curPost := range p.data {
		if curPost.ID == postID {
			p.data[idx].Votes = append(p.data[idx].Votes, vote)
			p.data[idx].Score += int64(vote.Vote)
			p.data[idx].UpvotePercentage = int8(p.data[idx].Score / int64(len(p.data[idx].Votes)) * 100)
			p.mu.Unlock()
			return p.data[idx], nil
		}
	}
	p.mu.Unlock()

	return domain.Post{}, errors.New("no post for the vote")
}

func (p *PostsRepo) AddCommentToPost(ctx context.Context, id uint64, comment domain.Comment) (domain.Post, error) {
	var post domain.Post

	p.mu.Lock()
	for idx, curPost := range p.data {
		if curPost.ID == id {
			p.data[idx].Comments = append(p.data[idx].Comments, comment)
			post = p.data[idx]
		}
	}
	p.mu.Unlock()

	return post, nil
}

func (p *PostsRepo) DeleteCommentFromPost(ctx context.Context, postID, commentID, userID uint64) (domain.Post, error) {
	post, err := p.GetPostByID(ctx, postID)
	if err != nil {
		return domain.Post{}, err
	}

	p.mu.Lock()
	for idx, curPost := range p.data {
		if curPost.ID == postID {
			for commentIdx, comment := range p.data[idx].Comments {
				if comment.ID == commentID {
					p.data[idx].Comments = append(p.data[idx].Comments[:commentIdx], p.data[idx].Comments[commentIdx+1:]...)
				}
			}
			post = p.data[idx]
		}
	}
	p.mu.Unlock()

	return post, nil
}

func (p *PostsRepo) GetPostsByUsername(ctx context.Context, username string) ([]domain.Post, error) {
	var posts []domain.Post
	for _, post := range p.data {
		if post.Author.Username == username {
			posts = append(posts, post)
		}
	}

	return posts, nil
}
