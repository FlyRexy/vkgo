package delivery

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/repository"
	"go.uber.org/zap"
)

type UserHandler struct {
	repo   *repository.Repository
	logger *zap.SugaredLogger
}

func NewUserHandler(repo *repository.Repository, logger *zap.SugaredLogger) *UserHandler {
	return &UserHandler{
		repo,
		logger,
	}
}

func (uh *UserHandler) SetupRoutes(r *mux.Router) *mux.Router {
	r.HandleFunc("/api/login", uh.Login)
	r.HandleFunc("/api/register", uh.Signup).Methods(http.MethodPost)

	return r
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	user := new(repository.User)
	err = json.Unmarshal(body, user)
	userFromRepo, err := uh.repo.UserRepo.CheckUserExist(user.Username, user.Password)
	uh.logger.Infoln(uh.repo.UserRepo)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": err.Error()})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	token := uh.repo.SessionRepo.CreateSession(userFromRepo)

	success, _ := json.Marshal(map[string]string{"token": token})
	w.Write(success)
}

func (uh *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	user := new(repository.User)
	err = json.Unmarshal(body, user)

	userFromRepo, err := uh.repo.UserRepo.CreateUser(user.Username, user.Password)
	uh.logger.Infoln(userFromRepo)
	if err != nil {
		resp, _ := json.Marshal(map[string]string{"error": err.Error()})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	token := uh.repo.SessionRepo.CreateSession(userFromRepo)

	success, _ := json.Marshal(map[string]string{"token": token})
	w.Write(success)
}
