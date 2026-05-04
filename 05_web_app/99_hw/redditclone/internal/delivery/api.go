package delivery

import (
	"github.com/gorilla/mux"
	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/repository"
	"go.uber.org/zap"
)

type Router interface {
	SetupRoutes(r *mux.Router) *mux.Router
}

func SetupAPI(repo *repository.Repository, logger *zap.SugaredLogger) *mux.Router {
	r := mux.NewRouter()
	r = SetupPosts(r)
	var routes []Router
	routes = append(routes, NewUserHandler(repo, logger))

	for _, route := range routes {
		r = route.SetupRoutes(r)
	}

	return r
}
