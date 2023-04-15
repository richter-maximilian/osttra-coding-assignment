package mock

import (
	"context"

	"github.com/RichterMaximilian/osttra-coding-assignment/model"
)

type Service struct {
	SubmitMessageFunc    func(ctx context.Context, recipientUserName, messageContent string) (string, error)
	FetchNewMessagesFunc func(ctx context.Context) ([]model.Message, error)
	DeleteMessagesFunc   func(ctx context.Context, messageIDs []string) error
	GetAllMessagesFunc   func(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error)
}

func (s *Service) SubmitMessage(ctx context.Context, recipientUserName, messageContent string) (string, error) {
	return s.SubmitMessageFunc(ctx, recipientUserName, messageContent)
}

func (s *Service) FetchNewMessages(ctx context.Context) ([]model.Message, error) {
	return s.FetchNewMessagesFunc(ctx)
}

func (s *Service) DeleteMessages(ctx context.Context, messageIDs []string) error {
	return s.DeleteMessagesFunc(ctx, messageIDs)
}

func (s *Service) GetAllMessages(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error) {
	return s.GetAllMessagesFunc(ctx, startCursor, endCursor)
}
