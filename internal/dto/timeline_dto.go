package dto

import "time"

type FanoutTask struct {
	MsgID     string       `json:"msg_id"`
	TweetID   int64        `json:"tweet_id"`
	AuthorID  int64		   `json:"author_id"`
	CreatedAt time.Time	   `json:"created_at"`
}

func NewFanoutTask(msgID string, tweetID, authorID int64, createdAt time.Time) *FanoutTask{
	return &FanoutTask{
		MsgID: msgID,
		TweetID: tweetID,
		AuthorID: authorID,
		CreatedAt: createdAt,
	}
}