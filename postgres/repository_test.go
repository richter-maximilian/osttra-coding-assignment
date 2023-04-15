package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RichterMaximilian/osttra-coding-assignment/model"
	"github.com/RichterMaximilian/osttra-coding-assignment/postgres"
	"github.com/RichterMaximilian/osttra-coding-assignment/testhelpers"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v4/pgxpool"
)

const migrationsPath = "file://../migrations"

func TestRepository_InsertMessage(t *testing.T) {
	t.Run("should insert a message", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		wantMessage := model.Message{
			ID:                "id",
			RecipientUserName: "recipient",
			Content:           "content",
			SentAt:            time.Now(),
		}

		if err := r.InsertMessage(ctx, wantMessage); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var gotMessage model.Message
		if err := pool.QueryRow(ctx, `
			SELECT
				id,
				user_name,
				content,
				sent_at,
				fetched_at
			FROM messages
		`).Scan(
			&gotMessage.ID,
			&gotMessage.RecipientUserName,
			&gotMessage.Content,
			&gotMessage.SentAt,
			&gotMessage.FetchedAt,
		); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(wantMessage, gotMessage); diff != "" {
			t.Fatalf("message mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("should return an error if the message ID already exists", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		message := model.Message{
			ID:                "id",
			RecipientUserName: "recipient",
			Content:           "content",
			SentAt:            time.Now(),
		}

		if err := r.InsertMessage(ctx, message); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err := r.InsertMessage(ctx, message)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestRepository_GetNewMessages(t *testing.T) {
	t.Run("should return non-fetched and non-deleted messages and set fetched_at", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		now := time.Now()

		messages := []model.Message{
			{
				ID:                "id1",
				RecipientUserName: "recipient1",
				Content:           "content1",
				SentAt:            now.Add(time.Hour),
				FetchedAt:         nil,
			},
			{
				ID:                "id2",
				RecipientUserName: "recipient2",
				Content:           "content2",
				SentAt:            now,
				FetchedAt:         &now,
			},
		}

		for _, message := range messages {
			insertMessage(ctx, t, pool.Pool, message)
		}

		gotMessages, err := r.GetNewMessages(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantMessages := []model.Message{messages[0]}

		if diff := cmp.Diff(wantMessages, gotMessages); diff != "" {
			t.Fatalf("messages mismatch (-want +got):\n%s", diff)
		}

		var gotFetchedAt *time.Time
		if err := pool.QueryRow(ctx, `
			SELECT fetched_at FROM messages WHERE id = $1::text
		`, messages[0].ID).Scan(&gotFetchedAt); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if gotFetchedAt == nil {
			t.Error("fetched at should not be nil")
		}
	})

	t.Run("should be ordered by sent_at", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		now := time.Now()

		messages := []model.Message{
			{
				ID:                "id1",
				RecipientUserName: "recipient1",
				Content:           "content1",
				SentAt:            now.Add(time.Hour),
				FetchedAt:         nil,
			},
			{
				ID:                "id2",
				RecipientUserName: "recipient2",
				Content:           "content2",
				SentAt:            now,
				FetchedAt:         nil,
			},
		}

		for _, message := range messages {
			insertMessage(ctx, t, pool.Pool, message)
		}

		gotMessages, err := r.GetNewMessages(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantMessages := []model.Message{messages[1], messages[0]}

		if diff := cmp.Diff(wantMessages, gotMessages); diff != "" {
			t.Fatalf("messages mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestRepository_DeleteMessages(t *testing.T) {
	t.Run("should delete messages", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		now := time.Now()

		messages := []model.Message{
			{
				ID:                "id1",
				RecipientUserName: "recipient1",
				Content:           "content1",
				SentAt:            now,
				FetchedAt:         nil,
			},
			{
				ID:                "id2",
				RecipientUserName: "recipient2",
				Content:           "content2",
				SentAt:            now,
				FetchedAt:         nil,
			},
		}

		messageIDs := make([]string, 0, len(messages))
		for _, message := range messages {
			insertMessage(ctx, t, pool.Pool, message)
			messageIDs = append(messageIDs, message.ID)
		}

		if err := r.DeleteMessages(ctx, messageIDs); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var count int
		if err := pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM messages
		`).Scan(&count); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := count, 0; got != want {
			t.Errorf("got %d messages, want %d", got, want)
		}
	})

	t.Run("should not delete messages if not all message ids are found", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		now := time.Now()

		messages := []model.Message{
			{
				ID:                "id1",
				RecipientUserName: "recipient1",
				Content:           "content1",
				SentAt:            now,
				FetchedAt:         nil,
			},
			{
				ID:                "id2",
				RecipientUserName: "recipient2",
				Content:           "content2",
				SentAt:            now,
				FetchedAt:         nil,
			},
		}

		messageIDs := make([]string, 0, len(messages)+1)
		for _, message := range messages {
			insertMessage(ctx, t, pool.Pool, message)
			messageIDs = append(messageIDs, message.ID)
		}

		messageIDs = append(messageIDs, "id3")

		err := r.DeleteMessages(ctx, messageIDs)
		if err == nil {
			t.Fatal("expected error")
		}

		if got, want := err, model.ErrNotFound; !errors.Is(got, want) {
			t.Errorf("got error %v, want %v", got, want)
		}

		var count int
		if err := pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM messages
		`).Scan(&count); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got, want := count, 2; got != want {
			t.Errorf("got %d messages, want %d", got, want)
		}
	})
}

func TestRepository_GetAllMessages(t *testing.T) {
	t.Run("should return all messages", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		now := time.Now()

		messages := []model.Message{
			{
				ID:                "id1",
				RecipientUserName: "recipient1",
				Content:           "content1",
				SentAt:            now,
				FetchedAt:         nil,
			},
			{
				ID:                "id2",
				RecipientUserName: "recipient2",
				Content:           "content2",
				SentAt:            now,
				FetchedAt:         &now,
			},
		}

		for _, message := range messages {
			insertMessage(ctx, t, pool.Pool, message)
		}

		gotMessages, err := r.GetAllMessages(ctx, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(messages, gotMessages); diff != "" {
			t.Fatalf("messages mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("should be ordered by sent_at", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		now := time.Now()

		messages := []model.Message{
			{
				ID:                "id1",
				RecipientUserName: "recipient1",
				Content:           "content1",
				SentAt:            now.Add(time.Hour),
				FetchedAt:         nil,
			},
			{
				ID:                "id2",
				RecipientUserName: "recipient2",
				Content:           "content2",
				SentAt:            now,
				FetchedAt:         nil,
			},
		}

		for _, message := range messages {
			insertMessage(ctx, t, pool.Pool, message)
		}

		gotMessages, err := r.GetAllMessages(ctx, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantMessages := []model.Message{messages[1], messages[0]}

		if diff := cmp.Diff(wantMessages, gotMessages); diff != "" {
			t.Fatalf("messages mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("should return messages later than startCursor", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		now := time.Now()

		messages := []model.Message{
			{
				ID:                "id1",
				RecipientUserName: "recipient1",
				Content:           "content1",
				SentAt:            now.Add(time.Hour),
				FetchedAt:         nil,
			},
			{
				ID:                "id2",
				RecipientUserName: "recipient2",
				Content:           "content2",
				SentAt:            now,
				FetchedAt:         nil,
			},
			{
				ID:                "id3",
				RecipientUserName: "recipient3",
				Content:           "content3",
				SentAt:            now.Add(2 * time.Hour),
				FetchedAt:         nil,
			},
		}

		for _, message := range messages {
			insertMessage(ctx, t, pool.Pool, message)
		}

		gotMessages, err := r.GetAllMessages(ctx, &messages[0].ID, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantMessages := []model.Message{messages[0], messages[2]}

		if diff := cmp.Diff(wantMessages, gotMessages); diff != "" {
			t.Fatalf("messages mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("should return messages earlier than endCursor", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		now := time.Now()

		messages := []model.Message{
			{
				ID:                "id1",
				RecipientUserName: "recipient1",
				Content:           "content1",
				SentAt:            now.Add(time.Hour),
				FetchedAt:         nil,
			},
			{
				ID:                "id2",
				RecipientUserName: "recipient2",
				Content:           "content2",
				SentAt:            now,
				FetchedAt:         nil,
			},
			{
				ID:                "id3",
				RecipientUserName: "recipient3",
				Content:           "content3",
				SentAt:            now.Add(2 * time.Hour),
				FetchedAt:         nil,
			},
		}

		for _, message := range messages {
			insertMessage(ctx, t, pool.Pool, message)
		}

		gotMessages, err := r.GetAllMessages(ctx, nil, &messages[0].ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantMessages := []model.Message{messages[1], messages[0]}

		if diff := cmp.Diff(wantMessages, gotMessages); diff != "" {
			t.Fatalf("messages mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("should return error if startCursor is not found", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		invalidID := "invalidID"

		_, err := r.GetAllMessages(ctx, &invalidID, nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if got, want := err, model.ErrNotFound; !errors.Is(got, want) {
			t.Errorf("got error %v, want %v", got, want)
		}
	})

	t.Run("should return error if endCursor is not found", func(t *testing.T) {
		pool := testhelpers.GetMigratedDBPool(context.Background(), migrationsPath)
		defer pool.Close()
		ctx := context.Background()

		r := postgres.NewRepository(pool.Pool)

		invalidID := "invalidID"

		_, err := r.GetAllMessages(ctx, nil, &invalidID)
		if err == nil {
			t.Fatal("expected error")
		}

		if got, want := err, model.ErrNotFound; !errors.Is(got, want) {
			t.Errorf("got error %v, want %v", got, want)
		}
	})
}

func insertMessage(ctx context.Context, t *testing.T, pool *pgxpool.Pool, message model.Message) {
	if _, err := pool.Exec(ctx, `
		INSERT INTO messages (
			id,
			user_name,
			content,
			sent_at,
			fetched_at
		) VALUES (
			$1::text,
			$2::text,
			$3::text,
			$4::timestamptz,
			$5::timestamptz
		)
	`,
		message.ID,
		message.RecipientUserName,
		message.Content,
		message.SentAt,
		message.FetchedAt,
	); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
