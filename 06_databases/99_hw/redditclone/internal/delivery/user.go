package delivery

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/di"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/lib"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/service"
	"go.uber.org/zap"
)

type UserHandler struct {
	authService di.AuthService
	logger      *zap.SugaredLogger
}

func NewUserHandler(authService di.AuthService, logger *zap.SugaredLogger) *UserHandler {
	return &UserHandler{
		authService,
		logger,
	}
}

func (uh *UserHandler) SetupRoutes(pub, pr *mux.Router) (*mux.Router, *mux.Router) {
	pub.HandleFunc("/login", uh.Login).Methods(http.MethodPost)
	pub.HandleFunc("/register", uh.Signup).Methods(http.MethodPost)

	return pub, pr
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req UserDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, err)
		return
	}

	_, token, err := uh.authService.Login(r.Context(), service.UserDTO{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	success, _ := json.Marshal(map[string]string{"token": string(token)})
	w.Write(success)
}

func (uh *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req UserDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, err)
		return
	}

	_, token, err := uh.authService.Signup(r.Context(), service.UserDTO{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	success, _ := json.Marshal(map[string]string{"token": string(token)})
	w.Write(success)
}
