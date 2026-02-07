package models

import(
	"time"
)

const (
	MaxSessionLife = 7 * 24 * time.Hour
	SessionDuration = 24 * time.Hour
)

type Session struct{
	ID            int64        `db:"id"`
	UserID        int64        `db:"user_id"`
	TokenHash     string       `db:"token_hash"`
	ExpiresAt     time.Time    `db:"expires_at"`
	CreatedAt     time.Time    `db:"created_at"`
}

type IssueResponse struct {
    Session  *Session
    Token string
}

func(s *Session) IsExpired() bool {
	if s == nil {
		return true
	}

	now := time.Now().UTC()
	if now.After(s.ExpiresAt.UTC()) {
		return true
	}
	maxAge := MaxSessionLife
	if now.After(s.CreatedAt.UTC().Add(maxAge)) {
		return true
	}
	return false
}

func (s *Session) ShouldRefresh() bool {
	if s == nil || s.IsExpired(){
		return false
	}

	totalDuration := s.ExpiresAt.UTC().Sub(s.CreatedAt.UTC())
	remaining := s.ExpiresAt.UTC().Sub(time.Now().UTC())

	if totalDuration <= 0 {
		return true
	}
	return remaining < totalDuration / 4
}