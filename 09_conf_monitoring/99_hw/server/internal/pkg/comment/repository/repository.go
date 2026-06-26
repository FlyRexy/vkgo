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

func NewRepository(logger *zap.SugaredLogger) domain.CommentRepository {
	return repository{logger: logger}
}

var (
	client = external.NewClient("comment")
)

func (r repository) Create(comment domain.Comment) error {
	reqBody, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:16000/comment", bytes.NewBuffer(reqBody))
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
		return errors.New("failed to create comment remotely")
	}

	return nil
}

func (r repository) Like(commentID string) error {
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://localhost:16000/comment/like?cid=%s", commentID),
		nil,
	)
	if err != nil {
		log.LogExternalRequest(r.logger, req, nil)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.LogExternalRequest(r.logger, req, resp)
		return err
	}

	log.LogExternalRequest(r.logger, req, resp)

	if resp.StatusCode != 200 {
		return errors.New("failed to like comment remotely")
	}

	return nil
}
