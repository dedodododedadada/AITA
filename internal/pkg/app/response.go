package app

import (
	"aita/internal/errcode"
	"time"
)


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

type FollowResponse struct {
	FollowerID  int64    	`json:"follower_id"`
	FollowingID int64       `json:"following_id"`
	CreatedAt   time.Time   `json:"created_at"`
}

type RelationResponse struct {
    MeID       int64 `json:"me_id"`       
    TargetID   int64 `json:"target_id"`   
    Following  bool  `json:"following"`    
    FollowedBy bool  `json:"followed_by"`  
    IsMutual   bool  `json:"is_mutual"`   
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

type Response struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Code    string `json:"code,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}


func Fail(err error) Response {
	return Response{
		Error: err.Error(),
		Code:  errcode.GetBusinessCode(err),
	}
}

func Success(data any) Response {
	return Response{
		Data: data,
		Code: "SUCCESS",
	}
}

func SuccessMsg(msg string) Response {
	return Response{
		Message: msg,
		Code:    "SUCCESS",
	}
}

func SuccessWithMeta(data any, meta any) Response {
	return Response{
		Data: data,
		Meta: meta,
		Code: "SUCCESS",
	}
}
