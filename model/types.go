package model

import "time"

type Message struct {
	ID                string     `json:"id"`
	RecipientUserName string     `json:"recipient_user_name"`
	Content           string     `json:"content"`
	SentAt            time.Time  `json:"sent_at"`
	FetchedAt         *time.Time `json:"fetched_at,omitempty"`
}
