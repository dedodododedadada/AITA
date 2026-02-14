package service

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"fmt"
	"time"
)

type SessionStore interface {
	Create(ctx context.Context, session *models.Session) (*models.Session, error)
	GetByHash(ctx context.Context, tokenHash string) (*models.Session, error)
	UpdateExpiresAt(ctx context.Context, expiresAt time.Time, id int64) error
    DeleteBySessionID(ctx context.Context, sessionID int64) error
}

type UserInfoProvider interface {
    ToMyPage(ctx context.Context, userID int64) (*models.User, error)
}

type TokenManager interface {
    Generate(length int) (string, error)
    Hash(token string) string
}


type sessionService struct {
    sessionStore SessionStore
    userService  UserInfoProvider
    tokenManager TokenManager
}

func NewSessionService(ss SessionStore, usvc UserInfoProvider, tm TokenManager) *sessionService {
    return &sessionService{
        sessionStore: ss,
        userService: usvc,
        tokenManager: tm,
    }
} 

func (s *sessionService) validateAndHash(token string) (string, error) {
    if token == ""  {
        return "", errcode.ErrSessionNotFound
    }
    if len(token) < 32 || len(token) > 255 {
        return "", errcode.ErrInvalidTokenFormat
    }
    return s.tokenManager.Hash(token), nil
}

func (s *sessionService) Issue(ctx context.Context, userID int64) (string, error) {
    if userID <= 0 {
        return "", errcode.ErrRequiredFieldMissing
    }

    token, err := s.tokenManager.Generate(32)
    if err != nil {
        return "", fmt.Errorf("トークンの生成に失敗しました: %w", err)
    }

    sessionData := &models.Session{
        UserID:    userID,
        TokenHash: s.tokenManager.Hash(token),
        ExpiresAt: time.Now().Add(24 * time.Hour).UTC(), 
    }

    _, err = s.sessionStore.Create(ctx, sessionData)
    if err != nil {
        return "", fmt.Errorf("発行に失敗しました: %w", err)
    }
    
    return token, nil
}

func (s *sessionService) authenticate(ctx context.Context, token string) (*models.Session, error) {
    tokenHash, err := s.validateAndHash(token)
    if err != nil {
        return nil, err
    }
    session, err := s.sessionStore.GetByHash(ctx, tokenHash)
    if err != nil {
        return nil, fmt.Errorf("セッションの取得に失敗しました: %w", err)
    }

    if session.IsExpired() {
        return nil, errcode.ErrSessionExpired
    }
    return session, nil
}

func (s *sessionService) Validate(ctx context.Context, token string) (*models.Session, error){
    session, err := s.authenticate(ctx, token)
    if err != nil {
        return nil, err
    }

    if _, err := s.userService.ToMyPage(ctx, session.UserID); err != nil {
        return nil, err
    }

    if session.ShouldRefresh() {
        if err := s.refreshSession(ctx, session); err != nil {
            return nil, err 
        }
    }
    return session, nil
}

func (s *sessionService) refreshSession(ctx context.Context, session *models.Session) error {
    newExpiry := time.Now().Add(models.SessionDuration).UTC()

    if err := s.sessionStore.UpdateExpiresAt(ctx, newExpiry, session.ID); err != nil {
        return fmt.Errorf("セッション期限の更新に失敗しました: %w", err)
    }

    session.ExpiresAt = newExpiry
    return nil
}

func (s *sessionService) Revoke(ctx context.Context, sessionID int64) error {
    if sessionID <= 0 {
        return errcode.ErrInvalidSessionID
    }
    err := s.sessionStore.DeleteBySessionID(ctx, sessionID)
    if err != nil {
        return fmt.Errorf("セッションの削除に失敗しました: %w", err)
    }

    return nil
}