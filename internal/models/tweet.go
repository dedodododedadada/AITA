package models

import(
	"time"
)

type Tweet struct{
	ID            int64        `db:"id"`
	UserID        int64		   `db:"user_id"`
	Content       string       `db:"content"`
	ImageURL     *string       `db:"image_url"`
	CreatedAt     time.Time    `db:"created_at"` 
	UpdatedAt     time.Time    `db:"updated_at"`
	IsEdited      bool         `db:"is_edited"`
}





func(t *Tweet) CanBeUpdated() bool {
	if t == nil {
		return false
	}

	duration := time.Now().UTC().Sub(t.CreatedAt.UTC())
	
	return duration <= editWindow
}





