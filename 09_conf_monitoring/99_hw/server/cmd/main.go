package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"server/internal/api/middleware"
	"server/internal/pkg/comment/handler"
	commentrepo "server/internal/pkg/comment/repository"
	commentsvc "server/internal/pkg/comment/service"
	"server/internal/pkg/session"
	threadhttp "server/internal/pkg/thread/handler"
	threadrepo "server/internal/pkg/thread/repository"
	threadsvc "server/internal/pkg/thread/service"
	"time"

	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	defer logger.Sync()
	if err != nil {
		fmt.Println("zap logger is not available")
	}
	sugarLogger := logger.Sugar()

	e := echo.New()
	sessionSvc := session.NewService(sugarLogger)

	threadRepo := threadrepo.NewRepository(sugarLogger)
	threadSvc := threadsvc.NewService(threadRepo)
	threadHandler := threadhttp.Handler{ThreadSvc: threadSvc}

	commentRepo := commentrepo.NewRepository(sugarLogger)
	commentSvc := commentsvc.NewService(commentRepo, threadRepo)
	commentHandler := handler.Handler{CommentSvc: commentSvc}

	api := e.Group("")
	api.Use(middleware.ObserveMiddleware(sugarLogger))
	api.Use(middleware.AuthEchoMiddleware(sessionSvc))

	api.GET("/thread/:id", threadHandler.GetThread)
	api.POST("/thread", threadHandler.CreateThread)
	api.POST("/thread/:tid/comment", commentHandler.Create)
	api.POST("/thread/:tid/comment/:cid/like", commentHandler.Like)
	e.GET("/metrics", func(ctx echo.Context) error {
		sugarLogger.Info("collecting metrics")
		promhttp.Handler().ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	})
	go startShooting()

	fmt.Print(e.Start(":8080"))
}

func startShooting() {
	for {
		reqBody, _ := json.Marshal(struct{}{})

		req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/thread", bytes.NewBuffer(reqBody))
		req.Header.Add("Cookie", "user=user1")
		req.Header.Add("Content-Type", "application/json")
		http.DefaultClient.Do(req)

		req, _ = http.NewRequest(http.MethodPost, "http://localhost:8080/thread/123/comment", bytes.NewBuffer(reqBody))
		req.Header.Add("Cookie", "user=user1")
		req.Header.Add("Content-Type", "application/json")
		http.DefaultClient.Do(req)
		time.Sleep(50 * time.Millisecond)
	}
}
