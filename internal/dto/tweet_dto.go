package dto

import (
	"aita/internal/models"
	"aita/internal/pkg/app"
	"time"
)

const(
	editwindow = 10 * time.Minute
)





type TweetRecord struct {
	ID            int64        
	UserID        int64		   
	Content       string      
	ImageURL     *string      
	CreatedAt     time.Time   
	UpdatedAt     time.Time   
	IsEdited      bool         
}



func (tr TweetRecord) ToTweetResponse() *app.TweetResponse {
	return &app.TweetResponse{
		ID: 		tr.ID,
		UserID:     tr.UserID,
		Content:    tr.Content,
		ImageURL:   tr.ImageURL,
		CreatedAt:  tr.CreatedAt,
		UpdatedAt:  tr.UpdatedAt,
		IsEdited:   tr.IsEdited,
	}
}

func(tr *TweetRecord) CanBeUpdated() bool {
	if tr == nil {
		return false
	}

	duration := time.Now().UTC().Sub(tr.CreatedAt.UTC())
	
	return duration <= editwindow
}

func (tr *TweetRecord) ToModel() *models.Tweet {
	if tr == nil {
		return &models.Tweet{}
	}

	return &models.Tweet{
		ID: tr.ID,
		UserID: tr.UserID,
		Content: tr.Content,
		ImageURL: tr.ImageURL,
		CreatedAt: tr.CreatedAt,
		UpdatedAt: tr.UpdatedAt,
		IsEdited: tr.IsEdited,
	}
}

func NewTweetRecord(tweet *models.Tweet) *TweetRecord {
	if tweet == nil {
		return &TweetRecord{}
	}

	return &TweetRecord{
		ID: tweet.ID,
		UserID: tweet.UserID,
		Content: tweet.Content,
		ImageURL: tweet.ImageURL,
		CreatedAt: tweet.CreatedAt,
		UpdatedAt: tweet.UpdatedAt,
		IsEdited: tweet.IsEdited,
	}
}