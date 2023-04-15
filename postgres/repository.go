package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RichterMaximilian/osttra-coding-assignment/model"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) InsertMessage(ctx context.Context, message model.Message) error {
	if _, err := r.pool.Exec(ctx, `
		INSERT INTO messages (
			id,
			user_name,
			content,
			sent_at
		) VALUES (
			$1::text,
			$2::text,
			$3::text,
			$4::timestamptz
		)
	`,
		message.ID,
		message.RecipientUserName,
		message.Content,
		message.SentAt,
	); err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	return nil
}

func (r *Repository) GetNewMessages(ctx context.Context) ([]model.Message, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			user_name,
			content,
			sent_at
		FROM messages
		WHERE fetched_at IS NULL
		ORDER BY sent_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("select messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	var messageIDs []string
	for rows.Next() {
		var message model.Message
		if err := rows.Scan(
			&message.ID,
			&message.RecipientUserName,
			&message.Content,
			&message.SentAt,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		messages = append(messages, message)
		messageIDs = append(messageIDs, message.ID)
	}

	if _, err := r.pool.Exec(ctx, `
		UPDATE messages
		SET fetched_at = NOW()
		WHERE id = ANY($1::text[])
	`, messageIDs); err != nil {
		return nil, fmt.Errorf("update fetched_at: %w", err)
	}

	return messages, nil
}

func (r *Repository) DeleteMessages(ctx context.Context, messageIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if cmdTag, err := tx.Exec(ctx, `
		DELETE FROM messages
		WHERE id = ANY($1::text[])
	`, messageIDs); err != nil {
		return fmt.Errorf("delete messages: %w", err)
	} else if cmdTag.RowsAffected() != int64(len(messageIDs)) {
		return fmt.Errorf("delete messages: %w", model.ErrNotFound)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) GetAllMessages(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error) {
	var startAt, endAt *time.Time
	if startCursor != nil {
		cursor, err := r.getSentAt(ctx, *startCursor)
		if err != nil {
			return nil, fmt.Errorf("get start at: %w", err)
		}
		startAt = &cursor
	}

	if endCursor != nil {
		cursor, err := r.getSentAt(ctx, *endCursor)
		if err != nil {
			return nil, fmt.Errorf("get end at: %w", err)
		}
		endAt = &cursor
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			user_name,
			content,
			sent_at,
			fetched_at
		FROM messages
		WHERE ($1::timestamptz IS NULL OR sent_at >= $1::timestamptz)
		AND ($2::timestamptz IS NULL OR sent_at <= $2::timestamptz)
		ORDER BY sent_at ASC
	`, startAt, endAt)
	if err != nil {
		return nil, fmt.Errorf("select messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var message model.Message
		if err := rows.Scan(
			&message.ID,
			&message.RecipientUserName,
			&message.Content,
			&message.SentAt,
			&message.FetchedAt,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func (r *Repository) getSentAt(ctx context.Context, messageID string) (time.Time, error) {
	var sentAt time.Time
	if err := r.pool.QueryRow(ctx, `
		SELECT sent_at
		FROM messages
		WHERE id = $1
	`, messageID).Scan(&sentAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, fmt.Errorf("message %q: %w", messageID, model.ErrNotFound)
		}
		return time.Time{}, fmt.Errorf("select sent_at: %w", err)
	}

	return sentAt, nil
}
