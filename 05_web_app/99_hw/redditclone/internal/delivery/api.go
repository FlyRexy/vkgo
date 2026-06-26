package delivery

import (
	"github.com/gorilla/mux"
	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/di"
	"go.uber.org/zap"
)

type Router interface {
	SetupRoutes(pub, pr *mux.Router) (*mux.Router, *mux.Router)
}

func SetupAPI(sInj di.ServiceInjector, logger *zap.SugaredLogger) *mux.Router {
	r := mux.NewRouter()
	public := r.PathPrefix("/api").Subrouter()
	private := r.PathPrefix("/api").Subrouter()

	var routes []Router
	routes = append(routes, NewUserHandler(sInj.AuthService, logger))
	routes = append(routes, NewPostsHandler(sInj.PostsService, logger))

	for _, route := range routes {
		public, private = route.SetupRoutes(public, private)
	}
	private.Use(AuthMiddleware)

	return r
}
