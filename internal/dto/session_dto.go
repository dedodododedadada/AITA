package dto

import (
	"aita/internal/models"
	"time"
)

type SessionRecord struct {
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
    CreatedAt time.Time
}

type SessionResponse struct {
    UserID    int64
	Token     string
	ExpiresAt time.Time
    CreatedAt time.Time
}

type AuthContext struct {
	UserID    int64
	Token     string
	Role      string
}

func (sr *SessionRecord) ToModel() *models.Session {
    if  sr == nil {
        return nil
    }

    return &models.Session{
        UserID: sr.UserID,
        TokenHash: sr.TokenHash,
        ExpiresAt: sr.ExpiresAt,
        CreatedAt: sr.CreatedAt,
    }
}

func ToSessionRecord(ms *models.Session) *SessionRecord {
    if ms == nil {
        return nil
    }

    return &SessionRecord{
        UserID: ms.UserID,
        TokenHash: ms.TokenHash,
        ExpiresAt: ms.ExpiresAt,
        CreatedAt: ms.CreatedAt,
    }
}

func ToSessionResponse(s *SessionRecord, token string) *SessionResponse {
    if s == nil {
        return nil
    }

    return &SessionResponse{
        UserID: s.UserID,
        Token: token,
        ExpiresAt: s.ExpiresAt,
        CreatedAt: s.CreatedAt,
    }
}