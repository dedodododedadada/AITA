package dto

import (
	"aita/internal/models"
	"time"
)

type AuthContext struct {
	UserID    int64
	SessionID int64
	Role      string
}

type SessionRecord struct {
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
    CreatedAt time.Time
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