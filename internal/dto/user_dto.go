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

type UserRecord struct {
	ID            	int64        	
	Username      	string       	
	Email         	string       	
	PasswordHash  	string       
	CreatedAt     	time.Time    	
	FollowerCount 	int64        	
	FollowingCount  int64         
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

type UserProfile struct {
	ID            	int64     `json:"id"`   	
	Username      	string    `json:"username"`   	       	
	FollowerCount 	int64     `json:"follower_count"`   	
	FollowingCount  int64     `json:"following_count"`  
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

func (u *UserRecord)ToUserResponse() *UserResponse {
	if u == nil {
        return nil
    }

	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

func (ur *UserRecord)ToUserModel() *models.User {
	if ur == nil {
		return nil 
	}

	return &models.User{
		ID: ur.ID,
		Username: ur.Username,
		Email: ur.Email,
		PasswordHash: ur.PasswordHash,
		CreatedAt: ur.CreatedAt,
		FollowerCount: ur.FollowerCount,
		FollowingCount: ur.FollowingCount,
	}
}

func (u *UserRecord)ToUserProfile() *UserProfile {
	if u == nil {
        return nil
    }

	return &UserProfile{
		ID:        		u.ID,
		Username:  		u.Username,
		FollowerCount: 	u.FollowerCount,
		FollowingCount: u.FollowingCount,
	}
}
func NewUserRecord(user *models.User) *UserRecord {
	if user == nil {
		return nil
	}

	return &UserRecord{
		ID: user.ID,
		Username: user.Username,
		Email: user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt: user.CreatedAt,
		FollowerCount: user.FollowerCount,
		FollowingCount: user.FollowingCount,
	}
}