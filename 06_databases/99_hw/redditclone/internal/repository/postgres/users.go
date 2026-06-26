package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/domain"
)

type UsersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(db *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{
		db,
	}
}

func (r *UsersRepository) Login(ctx context.Context, user domain.User) (domain.User, error) {
	err := r.db.QueryRow(ctx,
		`SELECT id FROM users
	 	 WHERE username = $1 AND password_hash = $2`, user.Username, user.Password).Scan(&user.ID)

	return user, err
}

func (r *UsersRepository) Signup(ctx context.Context, user domain.User) (domain.User, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`, user.Username, user.Password).Scan(&user.ID)

	return user, err
}
