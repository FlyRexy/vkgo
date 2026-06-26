package lib

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/domain"
)

type Token string

func CreateSessionToken(user domain.User) Token {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]interface{}{
			"username": user.Username,
			"id":       user.ID,
		},
	})
	tokenString, _ := token.SignedString([]byte("secretkey"))

	return Token(tokenString)
}

func GetToken(r *http.Request) (string, error) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 || authHeader[1] == "" {
		return "", errors.New("no auth token provided")
	}

	return authHeader[1], nil
}

func GetUserFromToken(token string) (*domain.User, error) {
	authToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) { return []byte("secretkey"), nil })
	if err != nil {
		return nil, errors.Join(err, errors.New("jwt parsing error"))
	}
	userMap, ok := authToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("incorrect jwt token")
	}

	user, ok := userMap["user"].(map[string]interface{})
	if !ok {
		return nil, errors.New("cant extract user from context in auth middleware")
	}
	name, ok := user["username"].(string)
	if !ok {
		return nil, errors.New("cant extract username from context in auth middleware")
	}
	id, ok := user["id"].(float64)
	if !ok {
		return nil, errors.New("cant extract id from context in auth middleware")
	}

	return &domain.User{
		Username: name,
		ID:       int64(id),
	}, nil
}
