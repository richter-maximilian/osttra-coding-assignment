package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

func (h *handler) postMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqBody := struct {
		RecipientUserName string `json:"recipient_user_name"`
		Content           string `json:"content"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondJSONStatus(w, &HTTPError{Message: fmt.Sprintf("decode request body: %v", err)}, http.StatusBadRequest)
		return
	}

	if reqBody.RecipientUserName == "" {
		respondJSONStatus(w, &HTTPError{Message: "user_name is required"}, http.StatusBadRequest)
		return
	}

	messageID, err := h.service.SubmitMessage(ctx, reqBody.RecipientUserName, reqBody.Content)
	if err != nil {
		h.logger.Error("error submitting message", zap.Error(err))
		respondJSONStatus(w, &HTTPError{Message: "error submitting message"}, http.StatusInternalServerError)
		return
	}

	respBody := struct {
		MessageID string `json:"message_id"`
	}{
		MessageID: messageID,
	}

	respondJSONStatus(w, &respBody, http.StatusOK)
}
