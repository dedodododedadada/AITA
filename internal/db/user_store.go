package db

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type postgresUserStore struct {
	database *sqlx.DB
}

func NewPostgresUserStore(db *sqlx.DB) *postgresUserStore {
	return &postgresUserStore{database: db}
}

func (s *postgresUserStore) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `INSERT INTO users(username, email, password_hash) 
			  VALUES ($1, $2, $3) 
			  RETURNING id, username, email, password_hash, created_at`

	var newUser models.User
	err := s.database.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
	).Scan(
		&newUser.ID,
		&newUser.Username,
		&newUser.Email,
		&newUser.PasswordHash,
		&newUser.CreatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case errCodeUniqueViolation:
				switch pqErr.Constraint {
				case constraintUsernameK:
					return nil, errcode.ErrUsernameConflict
				case constraintUseremailK:
					return nil, errcode.ErrEmailConflict
				}
			case errCodeStringDataRightTruncation:
				return nil, errcode.ErrValueTooLong
			}
		}
		return nil, fmt.Errorf("ユーザーの生成に失敗しました: %w", err)
	}

	newUser.CreatedAt = newUser.CreatedAt.UTC()
	return &newUser, nil
}

func (s *postgresUserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var newUser models.User
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email = $1`
	err := s.database.GetContext(ctx, &newUser, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, fmt.Errorf("emailによるユーザー取得に失敗しました: %w", err)
	}

	newUser.CreatedAt = newUser.CreatedAt.UTC()
	return &newUser, nil
}

func (s *postgresUserStore) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var newUser models.User
	query := `SELECT id, username, email, password_hash,created_at FROM users WHERE id = $1`
	err := s.database.GetContext(ctx, &newUser, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, fmt.Errorf("userIDによるユーザー取得に失敗しました: %w", err)
	}

	newUser.CreatedAt = newUser.CreatedAt.UTC()
	return &newUser, nil
}
