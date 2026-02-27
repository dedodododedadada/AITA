package service

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"fmt"
	"log"
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

func(s *sessionService) expirationCheck(expiresAt time.Time, createdAt time.Time)  error {
    if expiresAt.IsZero() || createdAt.IsZero() {
        return errcode.ErrRequiredFieldMissing
    }

    now  := time.Now().UTC()
    if now.After(expiresAt.UTC()) {
        return errcode.ErrSessionExpired
    }

    if now.After(createdAt.UTC().Add(MaxSessionLife)) {
        return errcode.ErrSessionExpired
    }

    return nil
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

    err = s.expirationCheck(record.ExpiresAt, record.CreatedAt)
    if err != nil {
        return nil, err
    }

    return record, nil
}

func (s *sessionService) executeRefresh(ctx context.Context, token string) error {
    tokenHash, err := s.validateAndHash(token)
    record, err := s.sessionRepository.Get(ctx, tokenHash)
    if err != nil || record == nil {
        return fmt.Errorf("セッションの取得に失敗しました: %w", err)
    }

    newExpiry := time.Now().Add(SessionDuration).UTC()
    maxExpiry := record.CreatedAt.Add(MaxSessionLife).UTC()
    if newExpiry.After(maxExpiry) {
        newExpiry = maxExpiry 
    }
    record.ExpiresAt = newExpiry
    err = s.sessionRepository.Update(ctx, record)
    if err != nil {
        return fmt.Errorf("セッション期限の更新に失敗しました: %w", err)
    }

    return nil
}

func (s *sessionService) RefreshAsync(token string) {
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        defer func() {
            if r := recover(); r != nil {
                log.Printf("RefreshAsync panic: %v", r)
            }
        }()
        _ = s.executeRefresh(ctx, token)
    }()
}
// wil add user validation by userrepo
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
        ExpiresAt: time.Now().Add(SessionDuration).UTC(), 
        CreatedAt: time.Now().UTC(),
    }

    record, err := s.sessionRepository.Create(ctx, data)
    if err != nil || record == nil {
        return nil, fmt.Errorf("発行に失敗しました: %w", err)
    }
    

    return dto.ToSessionResponse(record, token), nil
}

// ToMyPage will be replaced by exist in userrepo
func (s *sessionService) Validate(ctx context.Context, token string) (*dto.SessionResponse, error) {
    record, err := s.authenticate(ctx, token)
    if err != nil {
        return nil, err
    }

    if _, err := s.userService.ToMyPage(ctx, record.UserID); err != nil {
        return nil, err
    }

   
    return dto.ToSessionResponse(record, token), nil
}



func (s *sessionService) ShouldRefresh(expiresAt time.Time, createdAt time.Time) (bool, error) {
    err := s.expirationCheck(expiresAt, createdAt)
    if err != nil {
        return false, err
    }

    totalDuration := expiresAt.UTC().Sub(createdAt.UTC())
    remaining := expiresAt.UTC().Sub(time.Now().UTC())
    if totalDuration <= 0 {
        return false, errcode.ErrSessionExpired
    }

    return remaining < totalDuration / 4, nil
}



func (s *sessionService) Revoke(ctx context.Context, userID int64, token string) error {
    check, err := s.authenticate(ctx, token)
    if err != nil {
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