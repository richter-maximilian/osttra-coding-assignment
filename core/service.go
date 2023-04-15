package core

import (
	"context"
	"fmt"
	"time"

	"github.com/RichterMaximilian/osttra-coding-assignment/model"
	"github.com/google/uuid"
)

type Repository interface {
	InsertMessage(ctx context.Context, message model.Message) error
	GetNewMessages(ctx context.Context) ([]model.Message, error)
	DeleteMessages(ctx context.Context, messageIDs []string) error
	GetAllMessages(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error)
}

type Service struct {
	repo Repository
	now  nowFunc
	uuid uuidFunc
}

type nowFunc func() time.Time

type uuidFunc func() string

type serviceOptsFunc func(s *Service)

func NewService(repo Repository, opts ...serviceOptsFunc) *Service {
	s := &Service{
		repo: repo,
		now:  time.Now,
		uuid: uuid.NewString,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func ServiceNow(now nowFunc) serviceOptsFunc {
	return func(s *Service) {
		s.now = now
	}
}

func ServiceUUID(uuid uuidFunc) serviceOptsFunc {
	return func(s *Service) {
		s.uuid = uuid
	}
}

func (s *Service) SubmitMessage(ctx context.Context, recipientUserName, messageContent string) (string, error) {
	message := model.Message{
		ID:                s.uuid(),
		RecipientUserName: recipientUserName,
		Content:           messageContent,
		SentAt:            s.now(),
	}

	if err := s.repo.InsertMessage(ctx, message); err != nil {
		return "", fmt.Errorf("insert message: %w", err)
	}

	return message.ID, nil
}

func (s *Service) FetchNewMessages(ctx context.Context) ([]model.Message, error) {
	messages, err := s.repo.GetNewMessages(ctx)
	if err != nil {
		return nil, fmt.Errorf("get new messages: %w", err)
	}

	return messages, nil
}

func (s *Service) DeleteMessages(ctx context.Context, messageIDs []string) error {
	if err := s.repo.DeleteMessages(ctx, messageIDs); err != nil {
		return fmt.Errorf("delete messages: %w", err)
	}

	return nil
}

func (s *Service) GetAllMessages(ctx context.Context, startCursor, endCursor *string) ([]model.Message, error) {
	messages, err := s.repo.GetAllMessages(ctx, startCursor, endCursor)
	if err != nil {
		return nil, fmt.Errorf("get all messages: %w", err)
	}

	return messages, nil
}
