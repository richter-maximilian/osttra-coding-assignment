package api_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/RichterMaximilian/osttra-coding-assignment/api"
	"github.com/RichterMaximilian/osttra-coding-assignment/mock"
	"github.com/RichterMaximilian/osttra-coding-assignment/model"
	"go.uber.org/zap"
)

func TestHandler_DeleteMessages(t *testing.T) {
	t.Run("should forward message ids to service", func(t *testing.T) {
		wantMessageIDs := []string{"message-id-1", "message-id-2"}

		gotDeleteMessagesCalled := false
		service := &mock.Service{
			DeleteMessagesFunc: func(ctx context.Context, messageIDs []string) error {
				gotDeleteMessagesCalled = true
				if got, want := messageIDs, wantMessageIDs; !reflect.DeepEqual(got, want) {
					t.Errorf("got message ids %v, want %v", got, want)
				}
				return nil
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
		reqBody := fmt.Sprintf(`{"message_ids": ["%s", "%s"]}`, wantMessageIDs[0], wantMessageIDs[1])
		req, err := http.NewRequest(http.MethodDelete, url, strings.NewReader(reqBody))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := testServer.Client().Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := resp.StatusCode, http.StatusNoContent; got != want {
			t.Errorf("got HTTP status %d, want %d", got, want)
		}

		wantDeleteMessagesCalled := true
		if got, want := gotDeleteMessagesCalled, wantDeleteMessagesCalled; got != want {
			t.Errorf("got delete messages called %t, want %t", got, want)
		}
	})

	t.Run("should return 400 if request body is invalid", func(t *testing.T) {
		service := &mock.Service{}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
		req, err := http.NewRequest(http.MethodDelete, url, strings.NewReader("{"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := testServer.Client().Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := resp.StatusCode, http.StatusBadRequest; got != want {
			t.Errorf("got HTTP status %d, want %d", got, want)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantRespBody := "{\"message\":\"decode request body: unexpected EOF\"}\n"
		if got, want := string(respBody), wantRespBody; got != want {
			t.Errorf("got response body %q, want %q", got, want)
		}
	})

	t.Run("should return 400 if no message IDs provided", func(t *testing.T) {
		service := &mock.Service{}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
		req, err := http.NewRequest(http.MethodDelete, url, strings.NewReader("{}"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := testServer.Client().Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := resp.StatusCode, http.StatusBadRequest; got != want {
			t.Errorf("got HTTP status %d, want %d", got, want)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantRespBody := "{\"message\":\"message_ids is required\"}\n"
		if got, want := string(respBody), wantRespBody; got != want {
			t.Errorf("got response body %q, want %q", got, want)
		}
	})

	t.Run("should return 404 if message IDs are not found", func(t *testing.T) {
		service := &mock.Service{
			DeleteMessagesFunc: func(ctx context.Context, messageIDs []string) error {
				return model.ErrNotFound
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
		req, err := http.NewRequest(http.MethodDelete, url, strings.NewReader(`{"message_ids": ["message-id-1"]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := testServer.Client().Do(req)
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

	t.Run("should return 500 if service returns an error", func(t *testing.T) {
		service := &mock.Service{
			DeleteMessagesFunc: func(ctx context.Context, messageIDs []string) error {
				return errors.New("some error")
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
		req, err := http.NewRequest(http.MethodDelete, url, strings.NewReader(`{"message_ids": ["message-id-1"]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := testServer.Client().Do(req)
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

		wantRespBody := "{\"message\":\"error deleting messages\"}\n"
		if got, want := string(respBody), wantRespBody; got != want {
			t.Errorf("got response body %q, want %q", got, want)
		}
	})
}
