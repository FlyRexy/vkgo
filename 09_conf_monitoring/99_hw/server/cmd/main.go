package main

import (
	"fmt"
	"server/internal/api/middleware"
	"server/internal/pkg/comment/handler"
	commentrepo "server/internal/pkg/comment/repository"
	commentsvc "server/internal/pkg/comment/service"
	"server/internal/pkg/session"
	threadhttp "server/internal/pkg/thread/handler"
	threadrepo "server/internal/pkg/thread/repository"
	threadsvc "server/internal/pkg/thread/service"

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
	e.Use(middleware.AuthEchoMiddleware(sessionSvc))

	e.Use(middleware.ObserveMiddleware(sugarLogger))

	threadRepo := threadrepo.NewRepository(sugarLogger)
	threadSvc := threadsvc.NewService(threadRepo)
	threadHandler := threadhttp.Handler{ThreadSvc: threadSvc}

	commentRepo := commentrepo.NewRepository(sugarLogger)
	commentSvc := commentsvc.NewService(commentRepo, threadRepo)
	commentHandler := handler.Handler{CommentSvc: commentSvc}

	e.GET("/thread/:id", threadHandler.GetThread)
	e.POST("/thread", threadHandler.CreateThread)
	e.POST("/thread/:tid/comment", commentHandler.Create)
	e.POST("/thread/:tid/comment/:cid/like", commentHandler.Like)
	e.GET("/metrics", func(ctx echo.Context) error {
		sugarLogger.Info("collecting metrics")
		promhttp.Handler().ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	})

	fmt.Print(e.Start(":8080"))
}
