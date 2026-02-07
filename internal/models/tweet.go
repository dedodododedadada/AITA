package models

import(
	"time"
)

type Tweet struct{
	ID            int64        `db:"id"`
	UserID        int64		   `db:"user_id"`
	Content       string       `db:"content"`
	ImageURL     *string       `db:"image_url"`
	CreatedAt     time.Time    `db:"created_at"` 
}

type CreateTweetRequest struct {
	Content		 string        `json:"content" binding:"max=1000"`
	ImageURL    *string        `json:"image_url" binding:"omitempty,url"`
}

type TweetResponse struct{
	ID            int64        `db:"id"`
	UserID        int64		   `db:"user_id"`
	Content       string       `db:"content"`
	ImageURL     *string       `db:"image_url"`
	CreatedAt     time.Time    `db:"created_at"` 
}

func NewTweetResponse(tweet *Tweet) TweetResponse {
	return TweetResponse{
		ID: 		tweet.ID,
		UserID:     tweet.UserID,
		Content:    tweet.Content,
		ImageURL:   tweet.ImageURL,
		CreatedAt:  tweet.CreatedAt,
	}
}




