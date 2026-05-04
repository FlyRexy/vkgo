package repository

import (
	"sync"

	"github.com/dgrijalva/jwt-go"
)

type SessionRepo struct {
	data map[string]bool
	mu   *sync.RWMutex
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{
		mu:   &sync.RWMutex{},
		data: make(map[string]bool),
	}
}

func (s *SessionRepo) CreateSession(user User) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]interface{}{
			"username": user.Username,
			"id":       user.ID,
		},
	})
	tokenString, _ := token.SignedString([]byte("secretkey"))

	s.mu.Lock()
	s.data[tokenString] = true
	s.mu.Unlock()

	return tokenString
}

func (s *SessionRepo) DeleteSession(token string) {
	s.mu.Lock()
	delete(s.data, token)
	s.mu.Unlock()
}
