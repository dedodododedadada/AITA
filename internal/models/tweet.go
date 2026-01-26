package models

import(
	"time"
)

type Tweet struct{
	ID            int64        `db:"id"`
	UserID        int64		   `db:"user_id"`
	Content       string       `db:"content"`
	ImageURL      *string      `db:"image_url"`
	CreatedAt     time.Time    `db:"created_at"` 
}
// Todo: Implement post creation after db\session_store.go is finalized.