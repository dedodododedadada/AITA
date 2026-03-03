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

type UserInfo struct {
	ID            	int64        	`db:"id" json:"id"`
	Username      	string       	`db:"username" json:"username"`
}



func (u *User) ToCacheInfo() *UserInfo {
	return &UserInfo{
		ID:           u.ID,
		Username:     u.Username,
	}
}






