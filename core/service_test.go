package core_test

import (
	"context"
	"testing"
	"time"

	"github.com/RichterMaximilian/osttra-coding-assignment/core"
	"github.com/RichterMaximilian/osttra-coding-assignment/mock"
	"github.com/RichterMaximilian/osttra-coding-assignment/model"
	"github.com/RichterMaximilian/osttra-coding-assignment/testhelpers"
	"github.com/google/go-cmp/cmp"
)

func TestService_SubmitMessage(t *testing.T) {
	t.Run("should submit message", func(t *testing.T) {
		fakeTime := testhelpers.NewFakeTime(time.Now())
		wantUUID := "uuid"
		fakeUUID := func() string { return wantUUID }

		wantMessage := model.Message{
			ID:                wantUUID,
			RecipientUserName: "recipient",
			Content:           "content",
			SentAt:            fakeTime.Now(),
			FetchedAt:         nil,
		}

		gotInsertMessageCall := false
		repo := &mock.Repository{
			InsertMessageFunc: func(ctx context.Context, message model.Message) error {
				gotInsertMessageCall = true
				if diff := cmp.Diff(wantMessage, message); diff != "" {
					t.Errorf("message mismatch (-want +got):\n%s", diff)
				}
				return nil
			},
		}

		service := core.NewService(repo, core.ServiceNow(fakeTime.Now), core.ServiceUUID(fakeUUID))

		gotMessageID, err := service.SubmitMessage(context.Background(), wantMessage.RecipientUserName, wantMessage.Content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := gotMessageID, wantMessage.ID; got != want {
			t.Errorf("got message ID %q, want %q", got, want)
		}

		wantInsertMessageCall := true
		if got, want := gotInsertMessageCall, wantInsertMessageCall; got != want {
			t.Errorf("got insert message call %v, want %v", got, want)
		}
	})
}

func TestService_FetchNewMessages(t *testing.T) {
	t.Run("should fetch new messages", func(t *testing.T) {
		wantMessages := []model.Message{
			{
				ID:                "id",
				RecipientUserName: "recipient",
				Content:           "content",
				SentAt:            time.Now(),
				FetchedAt:         nil,
			},
		}

		repo := &mock.Repository{
			GetNewMessagesFunc: func(ctx context.Context) ([]model.Message, error) {
				return wantMessages, nil
			},
		}

		service := core.NewService(repo)

		gotMessages, err := service.FetchNewMessages(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(wantMessages, gotMessages); diff != "" {
			t.Errorf("messages mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestService_DeleteMessages(t *testing.T) {
	t.Run("should delete messages", func(t *testing.T) {
		wantMessageIDs := []string{"id"}

		gotDeleteMessagesCall := false
		repo := &mock.Repository{
			DeleteMessagesFunc: func(ctx context.Context, messageIDs []string) error {
				gotDeleteMessagesCall = true
				if diff := cmp.Diff(wantMessageIDs, messageIDs); diff != "" {
					t.Errorf("message IDs mismatch (-want +got):\n%s", diff)
				}
				return nil
			},
		}

		service := core.NewService(repo)

		if err := service.DeleteMessages(context.Background(), wantMessageIDs); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantDeleteMessagesCall := true
		if got, want := gotDeleteMessagesCall, wantDeleteMessagesCall; got != want {
			t.Errorf("got delete messages call %v, want %v", got, want)
		}
	})
}

func TestService_GetAllMessages(t *testing.T) {
	t.Run("should get all messages", func(t *testing.T) {
		wantMessages := []model.Message{
			{
				ID:                "id",
				RecipientUserName: "recipient",
				Content:           "content",
				SentAt:            time.Now(),
				FetchedAt:         nil,
			},
		}

		repo := &mock.Repository{
			GetAllMessagesFunc: func(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error) {
				return wantMessages, nil
			},
		}

		service := core.NewService(repo)

		gotMessages, err := service.GetAllMessages(context.Background(), nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(wantMessages, gotMessages); diff != "" {
			t.Errorf("messages mismatch (-want +got):\n%s", diff)
		}
	})
}
