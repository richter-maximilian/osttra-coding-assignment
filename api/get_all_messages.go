package api

import (
	"errors"
	"net/http"

	"github.com/RichterMaximilian/osttra-coding-assignment/model"
	"go.uber.org/zap"
)

func (h *handler) getAllMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	startCursorRaw := r.URL.Query().Get("start_cursor")
	var startCursor *string
	if startCursorRaw != "" {
		startCursor = &startCursorRaw
	}

	endCursorRaw := r.URL.Query().Get("end_cursor")
	var endCursor *string
	if endCursorRaw != "" {
		endCursor = &endCursorRaw
	}

	messages, err := h.service.GetAllMessages(ctx, startCursor, endCursor)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			respondJSONStatus(w, &HTTPError{Message: "message not found"}, http.StatusNotFound)
			return
		}
		h.logger.Error("error fetching all messages", zap.Error(err))
		respondJSONStatus(w, &HTTPError{Message: "error fetching all messages"}, http.StatusInternalServerError)
		return
	}

	if messages == nil {
		messages = []model.Message{}
	}

	respondJSONStatus(w, &messages, http.StatusOK)
}
