package api

import (
	"aita/internal/contextkeys"
	"aita/internal/models"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TweetService interface {
	PostTweet(ctx context.Context, userID int64, req *models.CreateTweetRequest) (*models.Tweet, error)
}


type TweetHandler struct {
	tweetService TweetService
}

func NewTweetHandler(svc TweetService) *TweetHandler {
	return &TweetHandler{tweetService: svc}
}

func (h *TweetHandler) Create(c *gin.Context) {
	uidRow, exists := c.Get(contextkeys.AuthPayloadKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未認証です"})
		return
	}
	userID, ok := uidRow.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザーIDの型が正しくありません"})
		return
	}

	var req models.CreateTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエスト形式が正しくありません"})
		return
	}

	tweet, err := h.tweetService.PostTweet(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, models.ErrRequiredFieldMissing) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ツイート内容を入力してください"})
			return
		}
		if errors.Is(err, models.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ユーザーが存在しません"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバー内部エラーが発生しました"})
		return
	}
	c.JSON(http.StatusCreated, models.NewTweetResponse(tweet))
}