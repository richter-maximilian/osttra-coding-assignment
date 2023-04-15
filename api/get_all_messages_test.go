package api_test

import (
	"context"
	"encoding/json"
	"errors"
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

func TestHandler_GetAllMessages(t *testing.T) {
	t.Run("should forward start and end cursor to service", func(t *testing.T) {
		wantStartCursor := "start-cursor"
		wantEndCursor := "end-cursor"

		gotGetAllMessagesCalled := false
		service := &mock.Service{
			GetAllMessagesFunc: func(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error) {
				gotGetAllMessagesCalled = true
				if diff := cmp.Diff(&wantStartCursor, startCursor); diff != "" {
					t.Errorf("start cursor mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(&wantEndCursor, endCursor); diff != "" {
					t.Errorf("end cursor mismatch (-want +got):\n%s", diff)
				}
				return nil, nil
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages?start_cursor=%s&end_cursor=%s", testServer.URL, wantStartCursor, wantEndCursor)
		resp, err := testServer.Client().Get(url)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := resp.StatusCode, http.StatusOK; got != want {
			t.Errorf("got HTTP status %d, want %d", got, want)
		}

		wantGetAllMessagesCalled := true
		if got, want := gotGetAllMessagesCalled, wantGetAllMessagesCalled; got != want {
			t.Errorf("got GetAllMessages called %t, want %t", got, want)
		}
	})

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
			GetAllMessagesFunc: func(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error) {
				return wantMessages, nil
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
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
			t.Errorf("got messages %v, want %v, diff %s", gotMessages, wantMessages, diff)
		}
	})

	t.Run("should return empty array if no messages are found", func(t *testing.T) {
		service := &mock.Service{
			GetAllMessagesFunc: func(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error) {
				return nil, nil
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
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

	t.Run("should return 404 if no messages are found", func(t *testing.T) {
		service := &mock.Service{
			GetAllMessagesFunc: func(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error) {
				return nil, model.ErrNotFound
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
		resp, err := testServer.Client().Get(url)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := resp.StatusCode, http.StatusNotFound; got != want {
			t.Errorf("got HTTP status %d, want %d", got, want)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantRespBody := "{\"message\":\"message not found\"}\n"
		if got, want := string(respBody), wantRespBody; got != want {
			t.Errorf("got response body %q, want %q", got, want)
		}
	})

	t.Run("should return 500 if service returns error", func(t *testing.T) {
		service := &mock.Service{
			GetAllMessagesFunc: func(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error) {
				return nil, errors.New("some error")
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
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

		wantRespBody := "{\"message\":\"error fetching all messages\"}\n"
		if got, want := string(respBody), wantRespBody; got != want {
			t.Errorf("got response body %q, want %q", got, want)
		}
	})
}
