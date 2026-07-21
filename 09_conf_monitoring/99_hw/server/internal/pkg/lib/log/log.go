package log

import (
	"net/http"

	"go.uber.org/zap"
)

func LogExternalRequest(logger *zap.SugaredLogger, req *http.Request, resp *http.Response) {
	var statusCode int
	if resp == nil {
		statusCode = 500
	} else {
		statusCode = resp.StatusCode
	}
	if statusCode == http.StatusOK {
		logger.Infow(
			"external request finished",
			"host", req.Host,
			"path", req.URL.Path,
			"status", resp.StatusCode,
		)
	} else {
		logger.Errorw(
			"external request finished",
			"host", req.Host,
			"path", req.URL.Path,
			"status", resp.StatusCode,
		)
	}
}
