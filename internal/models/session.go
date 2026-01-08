package models
import(
	"time"
)

type Session struct{
	ID            int64        `db:"id"`
	UserID        int64        `db:"user_id"`
	TokenHash     string       `db:"token_hash"`
	ExpiresAt     time.Time    `db:"expires_at"`
}

func(s *Session) isExpired() bool {
	if s == nil {
		return true
	}
	return time.Now().After(s.ExpiresAt)
}