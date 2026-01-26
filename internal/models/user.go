package models

import(
	"time"
)

type SignupRequest struct{
	Username      string        `json:"username" binding:"required,min=4,max=50" `
	Email         string        `json:"email" binding:"required,email"`
	Password      string        `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct{
	Email         string        `json:"email" binding:"required,email"`
	Password      string        `json:"password" binding:"required,min=8,max=72"`
}

type User struct{
	ID            int64        `db:"id"`
	Username      string       `db:"username"`
	Email         string       `db:"email"`
	PasswordHash  string       `db:"password_hash"`
	CreatedAt     time.Time    `db:"created_at"`
} 

type UserResponse struct{
	ID            int64        `json:"id"`
	Username      string       `json:"username"`
	Email         string       `json:"email"`
	CreatedAt     time.Time    `json:"created_at"`
}

type LoginResponse struct{
	SessionToken  string       `json:"session_token"`	
	User          UserResponse `json:"user"`
}

func NewUserResponse(user *User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
