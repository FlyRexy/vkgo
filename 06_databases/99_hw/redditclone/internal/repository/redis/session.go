package redisRepo

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/domain"
)

type SessionRepository struct {
	redis *redis.Client
}

const (
	sessionKey = "sessions:"
)

func NewSessionRepository(r *redis.Client) *SessionRepository {
	return &SessionRepository{
		redis: r,
	}
}

func (r *SessionRepository) Create(ctx context.Context, user domain.User) (string, error) {
	keyWithID := sessionKey + strconv.Itoa(int(user.ID))
	jsonUser, err := json.Marshal(&user)
	if err != nil {
		return "", err
	}
	_, err = r.redis.Set(ctx, keyWithID, jsonUser, 86400).Result()

	if err != nil {
		return "", err
	}

	return keyWithID, nil
}
