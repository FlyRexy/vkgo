package lib

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/domain"
)

func WriteError(w http.ResponseWriter, err error) {
	resp, _ := json.Marshal(map[string]string{"error": err.Error()})
	w.WriteHeader(http.StatusBadRequest)
	w.Write(resp)
}

const (
	UserSessionContextKey = "secretsession"
)

func SaveUserToContext(parentCtx context.Context, user any) context.Context {
	ctx := context.WithValue(parentCtx, UserSessionContextKey, user)
	return ctx
}

func GetUserFromContext(ctx context.Context) (domain.User, error) {
	value := ctx.Value(UserSessionContextKey)
	user, ok := value.(domain.User)
	if !ok {
		return domain.User{}, errors.New("failed to extract session from context")
	}

	return user, nil
}
