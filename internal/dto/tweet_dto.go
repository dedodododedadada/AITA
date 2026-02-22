package dto

import (
	"aita/internal/errcode"
	"aita/internal/models"
	"strings"
	"time"
	"unicode/utf8"
)



type CreateTweetRequest struct {
	Content		 string        `json:"content" binding:"max=1000"`
	ImageURL    *string        `json:"image_url" binding:"omitempty,url"`
}

type UpdateTweetRequest struct {
    Content      string        `json:"content" binding:"required,max=1000"`
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

func (r *CreateTweetRequest) Validate() error {
    r.Content = strings.TrimSpace(r.Content)
    if r.Content == "" {
        return errcode.ErrRequiredFieldMissing
    }
	if r.ImageURL != nil {
        trimmed := strings.TrimSpace(*r.ImageURL)
        if trimmed == "" {
            r.ImageURL = nil
        } else {
            *r.ImageURL = trimmed
        }
    }
   	if utf8.RuneCountInString(r.Content) > 1000 {
		return errcode.ErrInvalidContentFormat
	}
    return nil
}

func (r *UpdateTweetRequest) Validate() error {
    r.Content = strings.TrimSpace(r.Content)
    if r.Content == "" {
        return errcode.ErrRequiredFieldMissing
    }
    if utf8.RuneCountInString(r.Content) > 1000 {
        return errcode.ErrInvalidContentFormat
    }
    return nil
}

func NewTweetResponse(tweet *models.Tweet) TweetResponse {
	return TweetResponse{
		ID: 		tweet.ID,
		UserID:     tweet.UserID,
		Content:    tweet.Content,
		ImageURL:   tweet.ImageURL,
		CreatedAt:  tweet.CreatedAt,
		UpdatedAt:  tweet.UpdatedAt,
		IsEdited:   tweet.IsEdited,
	}
}