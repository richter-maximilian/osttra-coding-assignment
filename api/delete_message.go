package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/RichterMaximilian/osttra-coding-assignment/model"
	"go.uber.org/zap"
)

func (h *handler) deleteMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqBody := struct {
		MessageIDs []string `json:"message_ids"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondJSONStatus(w, &HTTPError{Message: fmt.Sprintf("decode request body: %v", err)}, http.StatusBadRequest)
		return
	}

	if len(reqBody.MessageIDs) == 0 {
		respondJSONStatus(w, &HTTPError{Message: "message_ids is required"}, http.StatusBadRequest)
		return
	}

	err := h.service.DeleteMessages(ctx, reqBody.MessageIDs)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondJSONStatus(w, &HTTPError{Message: "message not found"}, http.StatusNotFound)
			return
		}
		h.logger.Error("error deleting messages", zap.Error(err))
		respondJSONStatus(w, &HTTPError{Message: "error deleting messages"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
