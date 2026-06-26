package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/delivery"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/di"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/repository"
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

	db := initDB(sugared)
	redisConn := initRedis(sugared)

	// repo := memory.NewRepository()
	pgRepo := repository.NewRepositories(db, redisConn)
	sInj := di.NewServiceInjector(pgRepo)
	r.PathPrefix("/api/").Handler(delivery.SetupAPI(sInj, sugared))

	r.PathPrefix("/static/").HandlerFunc(ServeStatic)
	r.PathPrefix("/").HandlerFunc(ServeMain)

	zapLogger.Info("starting server",
		zap.String("port", addr),
	)
	http.ListenAndServe(addr, r)
}

func initDB(logger *zap.SugaredLogger) *pgxpool.Pool {
	db, err := pgxpool.New(context.Background(), "postgres://myuser:mypass@localhost:5432/sreddit?sslmode=disable")
	if err != nil {
		logger.Fatalf("cannot init postgres database")
	}

	if err = db.Ping(context.Background()); err != nil {
		logger.Fatalf("can't get to db")
	}
	logger.Info("db connection completed succesfully")

	return db
}

func initRedis(logger *zap.SugaredLogger) *redis.Client {
	redisConn := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	pong, err := redisConn.Ping(context.Background()).Result()

	if err != nil {
		logger.Fatalf("redis connection err %+v", err)
	}

	logger.Info("successful redis connection", pong)

	return redisConn
}
