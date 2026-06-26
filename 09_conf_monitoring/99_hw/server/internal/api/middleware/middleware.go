package middleware

import (
	"server/internal/pkg/domain"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func AuthEchoMiddleware(service domain.SessionService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			_, err := service.CheckSession(context.Request().Header)
			if err != nil {
				return context.NoContent(401)
			}

			return next(context)
		}
	}
}

var (
	hitsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_hits",
		},
		[]string{"path", "status"},
	)
	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
	}, []string{"path"})
)

func init() {
	prometheus.MustRegister(hitsCounter, requestDuration)
}

func ObserveMiddleware(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			reqID := uuid.NewString()
			context.Set("requestID", reqID)
			start := time.Now()

			err := next(context)

			if err != nil {
				logger.Error(
					zap.Int("status", context.Response().Status),
					zap.Error(err),
				)
			} else {
				logger.Info(
					zap.String("requestID", reqID),
					zap.String("path", context.Path()),
					zap.Int("status", context.Response().Status),
					zap.String("tenant", context.Request().RemoteAddr),
					zap.Int64("duration", time.Since(start).Milliseconds()),
				)
			}

			hitsCounter.WithLabelValues(context.Path(), strconv.Itoa(context.Response().Status)).Inc()
			requestDuration.WithLabelValues(context.Path()).Observe(float64(time.Since(start).Seconds()))

			return err
		}
	}
}
