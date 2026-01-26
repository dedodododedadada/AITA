package db

import(
	"aita/internal/models"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/jmoiron/sqlx"
)

type SessionStore interface {
	Create(ctx context.Context, userID int64, duration time.Duration) (string, *models.Session, error)
	GetByToken(ctx context.Context, token string)(*models.Session, error)
}

type PostgresSessionStore struct {
	database *sqlx.DB
}

func NewPostgresSessionStore(DB *sqlx.DB) *PostgresSessionStore {
	return &PostgresSessionStore{database:DB}
}

func generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _,err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b),nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func (s *PostgresSessionStore) Create(ctx context.Context, userID int64, duration time.Duration) (string, *models.Session, error){
	rawToken, err := generateRandomToken(32)
	if err != nil {
		return "", nil, err
	}
	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(duration)
	var session models.Session
	query := `INSERT INTO sessions(user_id, token_hash, expires_at) 
			  VALUES ($1, $2, $3)
			  RETURNING id, user_id, token_hash, expires_at`
	err = s.database.GetContext(ctx, &session, query, userID, tokenHash, expiresAt)
	if err != nil {
		return"", nil, err
	}
	return rawToken, &session, nil
}

func (s *PostgresSessionStore) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	tokenHash := hashToken(token)
	var session models.Session
	query := `SELECT id, user_id,token_hash, expires_at
			  FROM sessions 
			  WHERE token_hash = $1`
	err := s.database.GetContext(ctx, &session, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	//if session.IsExpired() {
		//return nil, errors.New("セッションの期限が切れているので、無効です")
	//}
	return &session, nil
}




