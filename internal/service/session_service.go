package service

import (
	"aita/internal/models"
	"context"
	"fmt"
	"strings"
	"time"
)

type SessionStore interface {
	Create(ctx context.Context, session *models.Session) (*models.Session, error)
	GetByHash(ctx context.Context, tokenHash string) (*models.Session, error)
	UpdateExpiresAt(ctx context.Context, expiresAt time.Time, id int64) error
    DeleteByHash(ctx context.Context, tokenHash string) error 
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
    if token == "" || len(token) < 32 || len(token) > 255 {
        return "", models.ErrSessionNotFound
    }
    return s.tokenManager.Hash(token), nil
}

func (h *sessionService) extractBearerToken(header string) string {
    if header == "" {
        return ""
    }
    parts := strings.SplitN(header, " ", 2)
    if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
        return strings.TrimSpace(parts[1])
    }
    return ""
}

func (s *sessionService) Issue(ctx context.Context, userID int64) (*models.IssueResponse, error) {
    if userID <= 0 {
        return nil, models.ErrRequiredFieldMissing
    }

    token, err := s.tokenManager.Generate(32)
    if err != nil {
        return nil, fmt.Errorf("トークンの生成に失敗しました: %w", err)
    }

    sessionData := &models.Session{
        UserID:    userID,
        TokenHash: s.tokenManager.Hash(token),
        ExpiresAt: time.Now().Add(24 * time.Hour).UTC(), 
    }

    session, err := s.sessionStore.Create(ctx, sessionData)
    if err != nil {
        return nil, fmt.Errorf("発行に失敗しました: %w", err)
    }
    
    return &models.IssueResponse{
        Session: session,
        Token: token,
    }, nil
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
        return nil, models.ErrSessionExpired
    }
    return session, nil
}

func (s *sessionService) Validate(ctx context.Context, token string) (*models.Session, error){
    finalToken := s.extractBearerToken(token)
    if finalToken == "" {
        return nil, models.ErrSessionNotFound
    }
    session, err := s.authenticate(ctx, finalToken)
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

func (s *sessionService) Revoke(ctx context.Context, token string) error {
    finalToken := s.extractBearerToken(token)
    if finalToken == "" {
        return models.ErrSessionNotFound
    }
    tokenHash, err := s.validateAndHash(finalToken)
    if err != nil {
        return err
    }
    err = s.sessionStore.DeleteByHash(ctx, tokenHash)
    if err != nil {
        return fmt.Errorf("セッションの削除に失敗しました: %w", err)
    }

    return nil
}