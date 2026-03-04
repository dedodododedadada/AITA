package app

import (
	"aita/internal/errcode"
	"aita/internal/pkg/utils"
	"strings"
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

type FollowRequest struct {
	TargetID int64 `json:"target_id" binding:"required,gt=0"`
}

type UnFollowRequest struct {
	TargetID int64 `json:"target_id" binding:"required,gt=0"`
}


func (r *SignupRequest) Validate() error {
	r.Username = strings.TrimSpace(r.Username)
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)
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