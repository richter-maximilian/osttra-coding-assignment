package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RichterMaximilian/osttra-coding-assignment/api"
	"github.com/RichterMaximilian/osttra-coding-assignment/mock"
	"github.com/RichterMaximilian/osttra-coding-assignment/model"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

func TestHandler_GetNewMessages(t *testing.T) {
	t.Run("should return messages from service", func(t *testing.T) {
		wantMessages := []model.Message{
			{
				ID:                "id-1",
				RecipientUserName: "recipient-1",
				Content:           "content-1",
				SentAt:            time.Now(),
			},
			{
				ID:                "id-2",
				RecipientUserName: "recipient-2",
				Content:           "content-2",
				SentAt:            time.Now(),
			},
		}

		service := &mock.Service{
			FetchNewMessagesFunc: func(ctx context.Context) ([]model.Message, error) {
				return wantMessages, nil
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages/new", testServer.URL)
		resp, err := testServer.Client().Get(url)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := resp.StatusCode, http.StatusOK; got != want {
			t.Errorf("got HTTP status %d, want %d", got, want)
		}

		var gotMessages []model.Message
		if err := json.NewDecoder(resp.Body).Decode(&gotMessages); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(gotMessages, wantMessages); diff != "" {
			t.Errorf("messages mismatch (-got +want): %s", diff)
		}
	})

	t.Run("should return empty array if no messages are available", func(t *testing.T) {
		service := &mock.Service{
			FetchNewMessagesFunc: func(ctx context.Context) ([]model.Message, error) {
				return nil, nil
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages/new", testServer.URL)
		resp, err := testServer.Client().Get(url)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := resp.StatusCode, http.StatusOK; got != want {
			t.Errorf("got HTTP status %d, want %d", got, want)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantRespBody := "[]\n"
		if got, want := string(respBody), wantRespBody; got != want {
			t.Errorf("got response body %q, want %q", got, want)
		}
	})

	t.Run("should return 500 if service returns error", func(t *testing.T) {
		service := &mock.Service{
			FetchNewMessagesFunc: func(ctx context.Context) ([]model.Message, error) {
				return nil, fmt.Errorf("service error")
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages/new", testServer.URL)
		resp, err := testServer.Client().Get(url)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := resp.StatusCode, http.StatusInternalServerError; got != want {
			t.Errorf("got HTTP status %d, want %d", got, want)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantRespBody := "{\"message\":\"error fetching new messages\"}\n"
		if got, want := string(respBody), wantRespBody; got != want {
			t.Errorf("got response body %q, want %q", got, want)
		}
	})
}
