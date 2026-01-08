package db
import(
	"context"
	"encoding/base64"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"aita/internal/models"
	"errors"
	"time"
	"github.com/jmoiron/sqlx"
)

type Sessionstore interface {
	Create(ctx context.Context, userID int64, duration time.Duration) (string, *Session, error)
	GetByToken(ctx context.Context, token string)(*Session, error)
}

type PostgresSessionStore struct {
	db *sqlx.DB
}

func NewPostgresSessionStore(db *sqlx.DB) *PostgresSessionStore {
	return &PostgresSessionStore{db:db}
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

func (s *PostgresSessionStore) Create(ctx context.Context, userID int64, duration time.Duration) (string, *Session, error){
	rawToken, err := generateRandomToken(32)
	if err != nil {
		return "", nil, err
	}
	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(duration)
	var session Session
	query := `INSERT INTO SESSIONS (user_id, token_hash, expires_at) 
			  VALUES ($1, $2, $3)
			  RETURNING id, user_id, token_hash, expires_at`
	err = s.db.GetContext(ctx, &session, query, userID, tokenHash, expiresAt)
	if err != nil {
		return"", nil, err
	}
	return rawToken, &session, nil
}

func (s *PostgresSessionStore) GetByToken(ctx context.Context, token string) (*Session, error) {
	tokenHash := hashToken(token)
	var session Session
	query := `SELECT id, user_id,token_hash, expires_at
			  FROM sessions 
			  WHERE token_hash = $1 AND expires_at > NOW()`
	err := s.db.GetContext(ctx, &session, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("セッションの期限が切れているか、無効です")
		}
		return nil, err
	}
	return &session, nil
}




