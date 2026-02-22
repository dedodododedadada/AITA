package db

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type postgresSessionStore struct {
	database *sqlx.DB
}

func NewPostgresSessionStore(db *sqlx.DB) *postgresSessionStore {
	return &postgresSessionStore{database: db}
}

func (s *postgresSessionStore) Create(ctx context.Context, session *models.Session) (*models.Session, error) {
	query := `INSERT INTO sessions(user_id, token_hash, expires_at) 
              VALUES ($1, $2, $3)
              RETURNING id, user_id, token_hash, expires_at, created_at`

	var newSession models.Session
	err := s.database.QueryRowContext(
		ctx,
		query,
		session.UserID,
		session.TokenHash,
		session.ExpiresAt,
	).Scan(
		&newSession.ID,
		&newSession.UserID,
		&newSession.TokenHash,
		&newSession.ExpiresAt,
		&newSession.CreatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case errCodeForeignKeyViolation:
				if pqErr.Constraint == constraintSessionUserFK {
					return nil, errcode.ErrUserNotFound
				}
			case errCodeUniqueViolation:
				if pqErr.Constraint == constraintTokenHashUnique {
					return nil, errcode.ErrTokenConflict
				}
			case errCodeStringDataRightTruncation:
				return nil, errcode.ErrValueTooLong
			}
		}
		return nil, fmt.Errorf("セッションの生成に失敗しました: %w", err)
	}

	newSession.ExpiresAt = newSession.ExpiresAt.UTC()
	newSession.CreatedAt = newSession.CreatedAt.UTC()
	return &newSession, nil
}

func (s *postgresSessionStore) GetByHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	query := `SELECT id, user_id, token_hash, expires_at, created_at FROM sessions WHERE token_hash = $1`

	var newSession models.Session
	err := s.database.GetContext(ctx, &newSession, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errcode.ErrSessionNotFound
		}
		return nil, fmt.Errorf("tokenhashによるセッション取得に失敗しました: %w", err)
	}

	newSession.ExpiresAt = newSession.ExpiresAt.UTC()
	newSession.CreatedAt = newSession.CreatedAt.UTC()
	return &newSession, nil
}

func (s *postgresSessionStore) UpdateExpiresAt(ctx context.Context, expiresAt time.Time, id int64) error {
	query := `UPDATE sessions SET expires_at = $1 WHERE id = $2`

	result, err := s.database.ExecContext(ctx, query, expiresAt, id)

	if err != nil {
		return fmt.Errorf("セッション期限の更新に失敗しました: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗しました: %w", err)
	}

	if rows == 0 {
		return errcode.ErrSessionNotFound
	}

	return nil
}

func (s *postgresSessionStore) DeleteBySessionID(ctx context.Context, sessionID int64) error {
	query := `DELETE FROM sessions WHERE id = $1`
	result, err := s.database.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("セッションの削除に失敗しました: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗しました: %w", err)
	}
	if rows == 0 {
		return errcode.ErrSessionNotFound
	}
	return nil
}

func (s *postgresSessionStore) DeleteAllByUserID(ctx context.Context, userID int64) error {
	query := ` DELETE FROM sessions WHERE user_id = $1`
	_, err := s.database.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("ユーザーの全セッション削除に失敗しました: %w", err)
	}

	return nil
}
