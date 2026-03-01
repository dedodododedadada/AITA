package repository

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
)

type SessionStore interface {
	Create(ctx context.Context, session *models.Session) (*models.Session, error)
	Get(ctx context.Context, tokenHash string) (*models.Session, error) 
	Update(ctx context.Context, session *models.Session) error
	Delete(ctx context.Context, session *models.Session) error
	DeleteByUserID(ctx context.Context, userID int64) error
}

type sessionRepository struct {
	sessionStore  SessionStore
}

func NewSessionRepository(ss SessionStore) *sessionRepository {
	return &sessionRepository{
		sessionStore: ss,
	}
}

func (r *sessionRepository) Create(ctx context.Context, sr *dto.SessionRecord) (*dto.SessionRecord, error) {
	session := sr.ToModel()
	newSession, err := r.sessionStore.Create(ctx, session)

	if err != nil {
		return nil, err
	}

	if newSession == nil {
		return nil, errcode.ErrInternal
	}

	return dto.ToSessionRecord(newSession), nil
}

func (r *sessionRepository) Get(ctx context.Context, tokenHash string) (*dto.SessionRecord, error) {
	session, err := r.sessionStore.Get(ctx, tokenHash)

	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, errcode.ErrInternal
	}

	return dto.ToSessionRecord(session), nil
}

func(r * sessionRepository) Update(ctx context.Context, sr *dto.SessionRecord) error {
	session := sr.ToModel()

	err := r.sessionStore.Update(ctx, session)
	if err != nil {
		return err
	}

	return nil
}


func(r *sessionRepository) Delete(ctx context.Context, sr *dto.SessionRecord) error {
	session := sr.ToModel()

	err := r.sessionStore.Delete(ctx, session)
	if err != nil {
		return err
	}
	return nil
}