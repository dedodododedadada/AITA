package db
import(
	"context"
	"database/sql"
	"errors"
	"aita/internal/models"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"github.com/lib/pq"
)
type UserStore interface {
	Create(ctx context.Context, req *models.SignupRequest)(*models.User, error)
	GetByEmail(ctx context.Context, email string)(*models.User, error)
}

type PostgresUserStore struct {
	db *sqlx.DB
}


func NewPostgresUserStore(db *sqlx.DB) *PostgresUserStore {
	return &PostgresUserStore{db: db}
}

func(s *PostgresUserStore) Create(ctx context.Context, req *models.SignupRequest) (*models.User,error){
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password),bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	var user models.User
	query := `INSERT INTO users(username, email, password_hash) VALUES ($1, $2, $3) RETURNING id, username, email, password_hash, created_at`
	err = s.db.GetContext(ctx, &user, query, req.Username, req.Email, string(hashedPassword))
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505"{
			return nil, errors.New("ユーザー名かメールアドレスは登録済みです")
		}
		return nil, err
	}
	return &user,nil
}

func(s *PostgresUserStore) GetByEmail(ctx context.Context, email string) (*models.User,error) {
	var user models.User
	query := `SELECT id, username, email, password_hash,created_at FROM users WHERE email = $1`
	err := s.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("ユーザーが存在しません")
		}
		return nil, err
	}
	return &user, nil
}