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
			  RETURNING id, username, email, password_hash, created_at, follower_count, following_count`

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
		&newUser.FollowerCount,
		&newUser.FollowingCount,
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
	query := `SELECT id, username, email, password_hash, created_at, follower_count, following_count FROM users WHERE email = $1`
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
	query := `SELECT id, username, email, password_hash, created_at, follower_count, following_count FROM users WHERE id = $1`
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

func (s *postgresUserStore) IncreaseFollowerCount(ctx context.Context, tx *sqlx.Tx, userID, delta int64) error {
	query := `UPDATE users SET follower_count = GREATEST(0, follower_count + $1) WHERE id = $2`

	var res sql.Result
	var err error 
	if tx != nil {
		res, err = tx.ExecContext(ctx, query, delta, userID)
	} else {
		res, err = s.database.ExecContext(ctx, query, delta, userID)
	}

	if err != nil {
		return fmt.Errorf("follower countsの更新に失敗しました: %w", err)
	}

	rows, _ := res.RowsAffected()
    if rows == 0 {
        return errcode.ErrUserNotFound
    }
    return nil
}

func (s *postgresUserStore) IncreaseFollowingCount(ctx context.Context, tx *sqlx.Tx, userID, delta int64) error {
	query :=`Update users SET following_count= GREATEST(0, following_count + $1)  WHERE id = $2`

	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.ExecContext(ctx, query, delta, userID)
	} else {
		res, err = s.database.ExecContext(ctx, query, delta, userID)
	}

	if err != nil {
		return fmt.Errorf("following countsの更新に失敗しました: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errcode.ErrUserNotFound
	}

	return nil
}