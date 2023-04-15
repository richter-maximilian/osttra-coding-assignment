package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RichterMaximilian/osttra-coding-assignment/model"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Service interface {
	SubmitMessage(ctx context.Context, recipientUserName, messageContent string) (string, error)
	FetchNewMessages(ctx context.Context) ([]model.Message, error)
	DeleteMessages(ctx context.Context, messageIDs []string) error
	GetAllMessages(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error)
}

type handler struct {
	service Service
	logger  *zap.Logger
}

func NewRouter(service Service, logger *zap.Logger) *chi.Mux {
	r := chi.NewRouter()
	h := &handler{
		service: service,
		logger:  logger,
	}

	r.Post("/messages", h.postMessage)
	r.Get("/messages/new", h.getNewMessages)
	r.Delete("/messages", h.deleteMessages)
	r.Get("/messages", h.getAllMessages)

	return r
}

type HTTPError struct {
	Message string `json:"message"`
}

func respondJSONStatus(w http.ResponseWriter, v interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("failed to encode json response: %w", err)
	}
	return nil
}
