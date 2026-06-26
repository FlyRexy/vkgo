package repository

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"server/internal/pkg/domain"
	"server/internal/pkg/lib/external"
	"server/internal/pkg/lib/log"

	"go.uber.org/zap"
)

type repository struct {
	logger *zap.SugaredLogger
}

var (
	client = external.NewClient("thread")
)

func (r repository) Create(thread domain.Thread) error {
	reqBody, err := json.Marshal(thread)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:15000/thread", bytes.NewBuffer(reqBody))
	if err != nil {
		log.LogExternalRequest(r.logger, req, nil)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.LogExternalRequest(r.logger, req, resp)
		return err
	}

	log.LogExternalRequest(r.logger, req, resp)

	if resp.StatusCode != 200 {
		return errors.New("failed to create thread remotely")
	}

	return nil
}

func (r repository) Get(id string) (domain.Thread, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:15000/thread?id=%s", id), nil)
	if err != nil {
		log.LogExternalRequest(r.logger, req, nil)
		return domain.Thread{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.LogExternalRequest(r.logger, req, resp)
		return domain.Thread{}, err
	}

	if resp.StatusCode != 200 {
		log.LogExternalRequest(r.logger, req, resp)
		return domain.Thread{}, errors.New("failed to fetch thread remotely")
	}

	var thread domain.Thread
	err = json.NewDecoder(resp.Body).Decode(&thread)
	if err != nil {
		log.LogExternalRequest(r.logger, req, resp)
		return domain.Thread{}, err
	}

	log.LogExternalRequest(r.logger, req, resp)

	return thread, nil
}

func NewRepository(logger *zap.SugaredLogger) domain.ThreadRepository {
	return repository{
		logger: logger,
	}
}
