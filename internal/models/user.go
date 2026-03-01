package models

import (
	"time"
)

type User struct{
	ID            	int64        	`db:"id"`
	Username      	string       	`db:"username"`
	Email         	string       	`db:"email"`
	PasswordHash  	string       	`db:"password_hash"`
	CreatedAt     	time.Time    	`db:"created_at"`
	FollowerCount 	int64        	`db:"follower_count"`
	FollowingCount  int64           `db:"following_count"`
} 

type UserCacheInfo struct {
	ID            	int64        	`json:"id"`
	Username      	string       	`json:"username"`
	Email         	string       	`json:"email"`
	PasswordHash  	string       	`json:"password_hash"`
	CreatedAt     	time.Time    	`json:"created_at"`
}

func (u *User) ToCacheInfo() *UserCacheInfo {
	return &UserCacheInfo{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
}

func (c *UserCacheInfo) ToUser(follower, following int64) *User {
	return &User{
		ID:             c.ID,
		Username:       c.Username,
		Email:          c.Email,
		PasswordHash:   c.PasswordHash,
		CreatedAt:      c.CreatedAt,
		FollowerCount:  follower,
		FollowingCount: following,
	}
}




