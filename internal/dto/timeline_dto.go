package dto

import (
	"aita/internal/pkg/utils"
	"fmt"
	"time"
)

const (
    ActionCreate = "create"
    ActionDelete = "delete"
    ActionUpdate = "update" 
)
type FanoutTask struct {
	TweetID   int64        `json:"tweet_id"`
	AuthorID  int64		   `json:"author_id"`
	MsgID     string       `json:"msg_id"`
	CreatedAt time.Time	   `json:"created_at"`
	Action    string       `json:"action"`
}

func NewFanoutTask(tweetID, authorID int64, createdAt time.Time, action string) *FanoutTask{
	return &FanoutTask{
		TweetID: tweetID,
		AuthorID: authorID,
		CreatedAt: createdAt,
		Action: action,
	}
}

func (t *FanoutTask) ToMap() map[string]any {
    return map[string]any{
        "tweet_id":  fmt.Sprintf("%d", t.TweetID),
        "author_id": fmt.Sprintf("%d", t.AuthorID),
        "at":        fmt.Sprintf("%d", t.CreatedAt.Unix()),
        "action":    t.Action,
    }
}

func (t *FanoutTask) FromMap(msgID string, values map[string]any) {
    t.MsgID = msgID
    t.TweetID = utils.ParseInt64(values["tweet_id"].(string))
    t.AuthorID = utils.ParseInt64(values["author_id"].(string))
    t.CreatedAt = time.Unix(utils.ParseInt64(values["at"].(string)), 0)
    t.Action = values["action"].(string)
}