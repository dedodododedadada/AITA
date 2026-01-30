package api

import (
	"aita/internal/contextkeys"
	"aita/internal/models"
	"aita/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TweetHandler struct {
	tweetService *service.TweetService
}

func NewTweetHandler(svc *service.TweetService) *TweetHandler {
	return &TweetHandler{tweetService: svc}
}

func (h *TweetHandler) Create(c *gin.Context) {
	uidRow, exists := c.Get(contextkeys.AuthPayloadKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未認証です"})
		return
	}
	userID := uidRow.(int64)

	var req *models.CreateTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエスト形式が正しくありません"})
		return
	}

	tweet, err := h.tweetService.PostTweet(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, models.ErrContentEmpty) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ツイート内容を入力してください"})
			return
		}
		if errors.Is(err, models.ErrUserNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ユーザーが存在しません"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバー内部エラーが発生しました"})
		return
	}
	c.JSON(http.StatusCreated, models.NewTweetResponse(tweet))
}