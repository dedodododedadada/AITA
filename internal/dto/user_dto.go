package dto

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"aita/internal/pkg/utils"
	"strings"
	"time"
)

type SignupRequest struct {
	Username string `json:"username" binding:"required,min=4,max=50" `
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	SessionToken string        `json:"session_token"`
	User         *UserResponse `json:"user"`
}

func (r *SignupRequest) Validate() error {
	r.Username =strings.TrimSpace(r.Username)
	r.Email = strings.TrimSpace(r.Email)
	r.Password  = strings.TrimSpace(r.Password )
	if r.Username == "" || r.Email == "" || r.Password == "" {
		return errcode.ErrRequiredFieldMissing
	}
	if len(r.Username) < 4 || len(r.Username) > 50 {
		return errcode.ErrInvalidUsernameFormat
	}
	if len(r.Password) < 8 || len(r.Password) > 72 {
		return errcode.ErrInvalidPasswordFormat
	}
	if !utils.IsValidEmail(r.Email) || len(r.Email) > 255 {
		return errcode.ErrInvalidEmailFormat
	}	

	return nil
}


func (r *LoginRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)
	if r.Email == "" || r.Password == "" {
		return errcode.ErrRequiredFieldMissing
	}
	if len(r.Password) < 8 || len(r.Password) > 72 {
		return errcode.ErrInvalidPasswordFormat
	}
	if !utils.IsValidEmail(r.Email) || len(r.Email) > 255 {
		return errcode.ErrInvalidEmailFormat
	}
	return nil
}

func NewUserResponse(user *models.User) *UserResponse {
	if user == nil {
        return nil
    }

	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}