package models

import(
	"time"
)


type Session struct{
	ID            int64        `db:"id" json:"id"`
	UserID        int64        `db:"user_id" json:"uid"`
	TokenHash     string       `db:"token_hash" json:"hash"`
	ExpiresAt     time.Time    `db:"expires_at" json:"exp"`
	CreatedAt     time.Time    `db:"created_at" json:"iat"`
}

