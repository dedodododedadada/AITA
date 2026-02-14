package dto

import "time"

type Response struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Code    string `json:"code,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	SessionToken string       `json:"session_token"`
	User         *UserResponse `json:"user"`
}

type TweetResponse struct{
	ID            int64        `json:"id"`
	UserID        int64		   `json:"user_id"`
	Content       string       `json:"content"`
	ImageURL     *string       `json:"image_url"`
	CreatedAt     time.Time    `json:"created_at"` 
	UpdatedAt     time.Time    `json:"updated_at"`
	IsEdited      bool         `json:"is_edited"`
}


