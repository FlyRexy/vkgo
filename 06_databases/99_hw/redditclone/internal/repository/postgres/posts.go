package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/domain"
)

type PostsRepository struct {
	db *pgxpool.Pool
}

func NewPostsPGRepository(db *pgxpool.Pool) *PostsRepository {
	return &PostsRepository{
		db,
	}
}

func (r *PostsRepository) CreatePost(ctx context.Context, post domain.Post) (domain.Post, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO posts (author_id, title, body, type, category, url) 
		 VALUES ($1, $2, $3, $4, $5, $6) 
		 RETURNING id, created_at`,
		post.Author.ID, post.Title, post.Text, post.Type, post.Category, post.URL,
	).Scan(&post.ID, &post.CreatedAt)

	return post, err
}

func (r *PostsRepository) GetAllPosts(ctx context.Context) ([]domain.Post, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.id, p.created_at, p.score, p.upvote_percentage, p.title, 
						p.body, p.type, p.category, p.url, u.username, u.id 
		 FROM posts p 
		 JOIN users u ON u.id = p.author_id 
		 ORDER BY p.created_at DESC
		 `,
	)
	if err != nil {
		return []domain.Post{}, err
	}
	defer rows.Close()

	return processPosts(rows)
}

func (r *PostsRepository) GetPostByCategory(ctx context.Context, category string) ([]domain.Post, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.id, p.created_at, p.score, p.upvote_percentage, p.title, 
						p.body, p.type, p.category, p.url, u.username, u.id 
		 FROM posts p
		 JOIN users u ON u.id = p.author_id 
		 WHERE p.category = $1
		 ORDER BY p.created_at DESC
		 `,
		category,
	)
	if err != nil {
		return []domain.Post{}, err
	}
	defer rows.Close()

	return processPosts(rows)
}

func (r *PostsRepository) GetPostByID(ctx context.Context, postID uint64) (domain.Post, error) {
	var post domain.Post
	var author domain.User
	batch := &pgx.Batch{}
	batch.Queue(
		`SELECT p.id, p.created_at, p.score, p.upvote_percentage, p.title, 
						p.body, p.type, p.category, p.url, u.username, u.id
		 FROM posts p
		 JOIN users u ON u.id = p.author_id
		 WHERE p.id = $1
		 `, postID,
	)

	batch.Queue(
		`SELECT v.vote, v.user_id
		 FROM votes v
		 WHERE v.post_id = $1
		`, postID,
	)

	batch.Queue(
		`SELECT c.body, c.id, c.created_at, u.id, u.username
		 FROM comments c
		 JOIN users u ON u.id = c.author_id
		 WHERE c.post_id = $1
		 ORDER BY c.created_at DESC
		`, postID,
	)

	results := r.db.SendBatch(ctx, batch)

	if err := results.QueryRow().Scan(&post.ID, &post.CreatedAt, &post.Score, &post.UpvotePercentage, &post.Title,
		&post.Text, &post.Type, &post.Category, &post.URL, &author.Username, &author.ID); err != nil {
		return domain.Post{}, err
	}

	votesRows, err := results.Query()
	if err != nil {
		return domain.Post{}, err
	}
	post.Votes, err = processVotes(votesRows)
	if err != nil {
		return domain.Post{}, err
	}

	commentsRows, err := results.Query()
	if err != nil {
		return domain.Post{}, err
	}
	post.Comments, err = processComments(commentsRows)
	if err != nil {
		return domain.Post{}, err
	}

	return post, nil
}

func (r *PostsRepository) GetPostsByUsername(ctx context.Context, username string) ([]domain.Post, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.id, p.created_at, p.score, p.upvote_percentage, p.title, 
						p.body, p.type, p.category, p.url, u.username, u.id 
		 FROM posts p
		 JOIN users u ON u.id = p.author_id 
		 WHERE u.username = $1
		 ORDER BY p.created_at DESC
		 `,
		username,
	)
	if err != nil {
		return []domain.Post{}, err
	}
	defer rows.Close()

	return processPosts(rows)
}

func (r *PostsRepository) Vote(ctx context.Context, postID uint64, userID int64, amount int8) (domain.Post, error) {
	tx, err := r.db.Begin(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		return domain.Post{}, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO votes (post_id, user_id, vote) VALUES ($1, $2, $3) 
		ON CONFLICT (post_id, user_id) DO UPDATE SET vote = EXCLUDED.vote
	`, postID, userID, amount)
	if err != nil {
		return domain.Post{}, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE posts SET score = (
			SELECT COALESCE(SUM(vote), 0) FROM votes WHERE post_id = $1
		)
		WHERE id = $1
	`, postID)
	if err != nil {
		return domain.Post{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.Post{}, err
	}

	post, err := r.GetPostByID(ctx, postID)
	if err != nil {
		return domain.Post{}, err
	}

	return post, nil
}

func (r *PostsRepository) DiscardVote(ctx context.Context, postID uint64, userID int64) (domain.Post, error) {
	tx, err := r.db.Begin(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		return domain.Post{}, err
	}

	_, err = tx.Exec(ctx, `DELETE FROM votes WHERE post_id = $1 AND user_id = $2`, postID, userID)
	if err != nil {
		return domain.Post{}, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE posts SET score = (
			SELECT COALESCE(SUM(vote), 0) FROM votes WHERE post_id = $1
		)
		WHERE id = $1
	`, postID)
	if err != nil {
		return domain.Post{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.Post{}, err
	}

	post, err := r.GetPostByID(ctx, postID)
	if err != nil {
		return domain.Post{}, err
	}

	return post, nil
}

func (r *PostsRepository) AddCommentToPost(ctx context.Context, postID uint64, comment domain.Comment) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO comments (body, author_id, post_id)
		VALUES ($1, $2, $3)
	`, comment.Body, comment.Author.ID, postID)

	return err
}

func (r *PostsRepository) DeleteCommentFromPost(ctx context.Context, postID, commentID, userID uint64) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM comments WHERE id = $1 AND author_id = $2 AND post_id = $3
	`, commentID, userID, postID)

	return err
}

func processPosts(rows pgx.Rows) ([]domain.Post, error) {
	var posts []domain.Post

	for rows.Next() {
		var post domain.Post
		var author domain.User
		if err := rows.Scan(&post.ID, &post.CreatedAt, &post.Score, &post.UpvotePercentage, &post.Title,
			&post.Text, &post.Type, &post.Category, &post.URL, &author.Username, &author.ID); err != nil {
			return []domain.Post{}, err
		}

		post.Author = author
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func processVotes(rows pgx.Rows) ([]domain.Vote, error) {
	var votes []domain.Vote

	for rows.Next() {
		var vote domain.Vote
		if err := rows.Scan(&vote.Vote, &vote.UserID); err != nil {
			return []domain.Vote{}, err
		}

		votes = append(votes, vote)
	}

	return votes, rows.Err()
}

func processComments(rows pgx.Rows) ([]domain.Comment, error) {
	var comments []domain.Comment

	for rows.Next() {
		var comment domain.Comment
		var author domain.User
		if err := rows.Scan(&comment.Body, &comment.ID, &comment.CreatedAt, &author.ID, &author.Username); err != nil {
			return []domain.Comment{}, err
		}

		comment.Author = author
		comments = append(comments, comment)
	}

	return comments, rows.Err()
}
