package log

import (
	"net/http"

	"go.uber.org/zap"
)

func LogExternalRequest(logger *zap.SugaredLogger, req *http.Request, resp *http.Response) {
	if resp.StatusCode == http.StatusOK {
		logger.Info(zap.String("host", req.Host), zap.String("path", req.URL.Path), zap.Int("status", resp.StatusCode))
	} else {
		logger.Error(zap.String("host", req.Host), zap.String("path", req.URL.Path), zap.Int("status", resp.StatusCode))
	}
}
