package delivery

import (
	"net/http"

	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/lib"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := lib.GetToken(r)
		if err != nil {
			lib.WriteError(w, err)
			return
		}
		user, err := lib.GetUserFromToken(token)
		if err != nil {
			lib.WriteError(w, err)
			return
		}

		ctx := lib.SaveUserToContext(r.Context(), *user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
