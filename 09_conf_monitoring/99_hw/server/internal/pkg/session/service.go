package session

import (
	"errors"
	"net/http"
	"server/internal/pkg/domain"
	"server/internal/pkg/lib/external"
	"server/internal/pkg/lib/log"

	"go.uber.org/zap"
)

type service struct {
	logger *zap.SugaredLogger
}

func NewService(logger *zap.SugaredLogger) domain.SessionService {
	return service{
		logger: logger,
	}
}

var (
	client = external.NewClient("session")
)

func (s service) CheckSession(headers http.Header) (domain.Session, error) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:17000/int/CheckSession", nil)
	if err != nil {
		log.LogExternalRequest(s.logger, req, nil)
		return domain.Session{}, err
	}

	req.Header = headers

	resp, err := client.Do(req)
	if err != nil {
		log.LogExternalRequest(s.logger, req, resp)
		return domain.Session{}, err
	}

	log.LogExternalRequest(s.logger, req, resp)

	switch resp.StatusCode {
	case 500:
		return domain.Session{}, errors.New("failed to request check session")
	case 200:
		return domain.Session{}, nil
	default:
		return domain.Session{}, domain.ErrNoSession
	}
}
