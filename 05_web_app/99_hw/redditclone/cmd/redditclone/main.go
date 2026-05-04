package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/delivery"
	"gitlab.vk-golang.com/vk-golang/lectures/05_web_app/99_hw/redditclone/internal/repository"
	"go.uber.org/zap"
)

func ServeMain(w http.ResponseWriter, r *http.Request) {
	file, _ := os.ReadFile("./static/html/index.html")
	w.Write(file)
}

func ServeStatic(w http.ResponseWriter, r *http.Request) {
	file, _ := os.ReadFile("." + r.URL.Path)
	w.Write(file)
}

const (
	addr = ":8080"
)


func main() {
	r := mux.NewRouter()
	zapLogger, _ := zap.NewProduction()
	sugared := zapLogger.Sugar()
	defer zapLogger.Sync()

	repo := repository.NewRepository()
	r.PathPrefix("/api/").Handler(delivery.SetupAPI(repo, sugared))

	r.PathPrefix("/static/").HandlerFunc(ServeStatic)
	r.PathPrefix("/").HandlerFunc(ServeMain)

	zapLogger.Info("starting server",
		zap.String("port", addr),
	)
	http.ListenAndServe(addr, r)
}
