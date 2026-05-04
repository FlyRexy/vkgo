package delivery

import (
	"net/http"

	"github.com/gorilla/mux"
)

func SetupPosts(r *mux.Router) *mux.Router {
	// r.HandleFunc("/api/posts/", GetPosts)
	return r
}

func GetPosts(w http.ResponseWriter, r *http.Request) {

}
