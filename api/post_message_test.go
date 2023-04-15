package api_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RichterMaximilian/osttra-coding-assignment/api"
	"github.com/RichterMaximilian/osttra-coding-assignment/mock"
	"go.uber.org/zap"
)

func TestHandler_PostMessage(t *testing.T) {
	t.Run("should forward message to service", func(t *testing.T) {
		wantRecipientUserName, wantContent := "recipient", "content"

		gotSubmitMessageCalled := false
		service := &mock.Service{
			SubmitMessageFunc: func(ctx context.Context, recipientUserName, content string) (string, error) {
				gotSubmitMessageCalled = true
				if got, want := recipientUserName, wantRecipientUserName; got != want {
					t.Errorf("got recipient user name %q, want %q", got, want)
				}
				if got, want := content, wantContent; got != want {
					t.Errorf("got content %q, want %q", got, want)
				}
				return "", nil
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
		reqBody := fmt.Sprintf(`{"recipient_user_name": "%s", "content": "%s"}`, wantRecipientUserName, wantContent)
		resp, err := testServer.Client().Post(url, "application/json", strings.NewReader(reqBody))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := resp.StatusCode, http.StatusOK; got != want {
			t.Errorf("got HTTP status %d, want %d", got, want)
		}

		wantSubmitMessageCalled := true
		if got, want := gotSubmitMessageCalled, wantSubmitMessageCalled; got != want {
			t.Errorf("got submit message called %t, want %t", got, want)
		}
	})

	t.Run("should return message id in response", func(t *testing.T) {
		wantMessageID := "message-id"

		service := &mock.Service{
			SubmitMessageFunc: func(ctx context.Context, recipientUserName, content string) (string, error) {
				return wantMessageID, nil
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
		reqBody := `{"recipient_user_name": "recipient", "content": "content"}`
		resp, err := testServer.Client().Post(url, "application/json", strings.NewReader(reqBody))
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

		wantRespBody := fmt.Sprintf("{\"message_id\":\"%s\"}\n", wantMessageID)
		if got, want := string(respBody), wantRespBody; got != want {
			t.Errorf("got response body %q, want %q", got, want)
		}
	})

	t.Run("should return 400 if request body is invalid", func(t *testing.T) {
		service := &mock.Service{}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
		resp, err := testServer.Client().Post(url, "application/json", strings.NewReader("{"))
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

	t.Run("should return 400 if recipient user name is missing", func(t *testing.T) {
		service := &mock.Service{}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
		reqBody := `{"content": "content"}`
		resp, err := testServer.Client().Post(url, "application/json", strings.NewReader(reqBody))
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

		wantRespBody := "{\"message\":\"user_name is required\"}\n"
		if got, want := string(respBody), wantRespBody; got != want {
			t.Errorf("got response body %q, want %q", got, want)
		}
	})

	t.Run("should return 500 if service returns error", func(t *testing.T) {
		service := &mock.Service{
			SubmitMessageFunc: func(ctx context.Context, recipientUserName, content string) (string, error) {
				return "", errors.New("service error")
			},
		}

		testServer := httptest.NewServer(api.NewRouter(service, zap.NewNop()))

		url := fmt.Sprintf("%s/messages", testServer.URL)
		reqBody := `{"recipient_user_name": "recipient", "content": "content"}`
		resp, err := testServer.Client().Post(url, "application/json", strings.NewReader(reqBody))
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

		wantRespBody := "{\"message\":\"error submitting message\"}\n"
		if got, want := string(respBody), wantRespBody; got != want {
			t.Errorf("got response body %q, want %q", got, want)
		}
	})
}
