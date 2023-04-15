package api

import (
	"net/http"

	"github.com/RichterMaximilian/osttra-coding-assignment/model"
	"go.uber.org/zap"
)

func (h *handler) getNewMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	messages, err := h.service.FetchNewMessages(ctx)
	if err != nil {
		h.logger.Error("error fetching new messages", zap.Error(err))
		respondJSONStatus(w, &HTTPError{Message: "error fetching new messages"}, http.StatusInternalServerError)
		return
	}

	if messages == nil {
		messages = []model.Message{}
	}

	respondJSONStatus(w, &messages, http.StatusOK)
}
