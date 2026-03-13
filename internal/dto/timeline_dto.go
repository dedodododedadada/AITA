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

func (t *FanoutTask) FromMap(msgID string, values map[string]any) error {
	if msgID == "" {
		return fmt.Errorf("FromMap: メッセージIDが空です")
	}
	tweetIDStr, ok1 := values["tweet_id"].(string)
    authorIDStr, ok2 := values["author_id"].(string)
    atStr, ok3 := values["at"].(string)
    action, ok4 := values["action"].(string)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return fmt.Errorf(
			"FromMap: 必須フィールドの欠落または型エラー (tweet_id:%v, author_id:%v, at:%v, action:%v)",
			ok1, ok2, ok3, ok4)
	}

    t.TweetID = utils.ParseInt64(tweetIDStr)
    t.AuthorID = utils.ParseInt64(authorIDStr)
    t.CreatedAt = time.Unix(utils.ParseInt64(atStr), 0)
    t.Action = action

	if t.TweetID <= 0 || t.AuthorID <= 0 {
        return fmt.Errorf("FromMap: IDを0にすることはできません")
    }
	return nil
}