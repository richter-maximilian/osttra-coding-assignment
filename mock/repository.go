package mock

import (
	"context"

	"github.com/RichterMaximilian/osttra-coding-assignment/model"
)

type Repository struct {
	InsertMessageFunc  func(ctx context.Context, message model.Message) error
	GetNewMessagesFunc func(ctx context.Context) ([]model.Message, error)
	DeleteMessagesFunc func(ctx context.Context, messageIDs []string) error
	GetAllMessagesFunc func(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error)
}

func (r *Repository) InsertMessage(ctx context.Context, message model.Message) error {
	return r.InsertMessageFunc(ctx, message)
}

func (r *Repository) GetNewMessages(ctx context.Context) ([]model.Message, error) {
	return r.GetNewMessagesFunc(ctx)
}

func (r *Repository) DeleteMessages(ctx context.Context, messageIDs []string) error {
	return r.DeleteMessagesFunc(ctx, messageIDs)
}

func (r *Repository) GetAllMessages(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error) {
	return r.GetAllMessagesFunc(ctx, startCursor, endCursor)
}
