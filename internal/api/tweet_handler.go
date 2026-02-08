package api

import (
    "aita/internal/contextkeys"
    "aita/internal/models"
    "context"
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
		err := models.ErrSessionNotFound
        c.JSON(models.GetStatusCode(err), models.Fail(err))
        return
    }
    
    userID, ok := uidRow.(int64)
    if !ok {
        err := models.ErrSessionNotFound
        c.JSON(models.GetStatusCode(err), models.Fail(err))
        return
    }

    var req models.CreateTweetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        appErr := models.FilterBindError(err)
        c.JSON(models.GetStatusCode(appErr), models.Fail(appErr))
        return
    }

    tweet, err := h.tweetService.PostTweet(c.Request.Context(), userID, &req)
    if err != nil {
        c.JSON(models.GetStatusCode(err), models.Fail(err))
        return
    }

    c.JSON(http.StatusCreated, models.Success(models.NewTweetResponse(tweet)))
}