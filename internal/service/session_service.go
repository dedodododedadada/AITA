package service

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"fmt"
	"time"
)

type SessionRepository interface {
	Create(ctx context.Context, sr *dto.SessionRecord) (*dto.SessionRecord, error)
	Get(ctx context.Context, tokenHash string) (*dto.SessionRecord, error) 
	Update(ctx context.Context, sr *dto.SessionRecord) error
    Delete(ctx context.Context, sr *dto.SessionRecord) error
}

type UserInfoProvider interface {
    ToMyPage(ctx context.Context, userID int64) (*models.User, error)
}

type TokenManager interface {
    Generate(length int) (string, error)
    Hash(token string) string
}


type sessionService struct {
    sessionRepository SessionRepository
    userService  UserInfoProvider
    tokenManager TokenManager
}

func NewSessionService(sr SessionRepository, usvc UserInfoProvider, tm TokenManager) *sessionService {
    return &sessionService{
        sessionRepository: sr,
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

func(s *sessionService) IsExpired(sr *dto.SessionRecord) bool {
    if sr == nil {
        return true
    }

    now  := time.Now().UTC()
    if now.After(sr.ExpiresAt.UTC()) {
        return true
    }

    if now.After(sr.CreatedAt.UTC().Add(MaxSessionLife)) {
        return true
    }

    return false
}

func (s *sessionService) ShouldRefresh(sr *dto.SessionRecord) bool {
    if sr == nil || s.IsExpired(sr) {
        return false
    }

    totalDuration := sr.ExpiresAt.UTC().Sub(sr.CreatedAt.UTC())
    remaining := sr.ExpiresAt.UTC().Sub(time.Now().UTC())
    if totalDuration <= 0 {
        return true
    }

    return remaining < totalDuration / 4
}

func (s *sessionService) Issue(ctx context.Context, userID int64) (*dto.SessionResponse, error) {
    if userID <= 0 {
        return nil, errcode.ErrRequiredFieldMissing
    }

    token, err := s.tokenManager.Generate(32)
    if err != nil {
        return nil, fmt.Errorf("トークンの生成に失敗しました: %w", err)
    }

    data := &dto.SessionRecord{
        UserID:    userID,
        TokenHash: s.tokenManager.Hash(token),
        ExpiresAt: time.Now().Add(MaxSessionLife).UTC(), 
        CreatedAt: time.Now().UTC(),
    }

    record, err := s.sessionRepository.Create(ctx, data)
    if err != nil || record == nil {
        return nil, fmt.Errorf("発行に失敗しました: %w", err)
    }
    

    return dto.ToSessionResponse(record, token), nil
}

func (s *sessionService) authenticate(ctx context.Context, token string) (*dto.SessionRecord, error) {
    tokenHash, err := s.validateAndHash(token)
    if err != nil {
        return nil, err
    }
    record, err := s.sessionRepository.Get(ctx, tokenHash)
    if err != nil || record == nil {
        return nil, fmt.Errorf("セッションの取得に失敗しました: %w", err)
    }

    if s.IsExpired(record) {
        return nil, errcode.ErrSessionExpired
    }

    return record, nil
}
// Validate verifies the token's validity and checks the user's status.
//
// NOTE: Architectural Decoupling in progress.
// 1. [ ] REMOVE internal calls to ShouldRefresh and refreshSession.
// 2. [ ] REASON: To follow the "Idempotency" principle. Validate should be a Read-Only operation.
// 3. [ ] STRATEGY: Move async refresh logic to the Middleware/API layer to achieve 
//    non-blocking Sliding Expiration and improve response latency.
// 4. [ ] TESTING: This decoupling simplifies Unit Testing by removing the need to mock 
//    database updates during simple validation checks.

func (s *sessionService) Validate(ctx context.Context, token string) (*dto.SessionResponse, error) {
    record, err := s.authenticate(ctx, token)
    if err != nil {
        return nil, err
    }

    if _, err := s.userService.ToMyPage(ctx, record.UserID); err != nil {
        return nil, err
    }

    if s.ShouldRefresh(record) {
        if err := s.refreshSession(ctx, record); err != nil {
            return nil, err 
        }
    }
    return dto.ToSessionResponse(record, token), nil
}

func (s *sessionService) refreshSession(ctx context.Context, sr *dto.SessionRecord) error {
    newExpiry := time.Now().Add(SessionDuration).UTC()
    sr.ExpiresAt = newExpiry
    err := s.sessionRepository.Update(ctx, sr)
    if err != nil {
        return fmt.Errorf("セッション期限の更新に失敗しました: %w", err)
    }

    return nil
}

func (s *sessionService) Revoke(ctx context.Context, userID int64, token string) error {
    if userID <= 0 {
        return errcode.ErrRequiredFieldMissing
    }

    check, err := s.authenticate(ctx, token) 
    if err != nil  {
        return err
    }
    if check.UserID != userID {
        return errcode.ErrForbidden
    }

    data := &dto.SessionRecord{
        UserID: userID,
        TokenHash: check.TokenHash,
    }

    err = s.sessionRepository.Delete(ctx, data)
    if err != nil {
        return fmt.Errorf("セッションの削除に失敗しました: %w", err)
    }

    return nil
}