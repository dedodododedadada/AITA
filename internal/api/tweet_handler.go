package api

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"aita/internal/pkg/app"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TweetService interface {
	PostTweet(ctx context.Context, userID int64, content string, imageURL *string) (*models.Tweet, error)
	FetchTweet(ctx context.Context, tweetID int64) (*models.Tweet, error)
	EditTweet(ctx context.Context, newContent string, tweetID int64, userID int64) (*models.Tweet, bool, error)
	RemoveTweet(ctx context.Context, tweetID int64, userID int64) error
}

type TweetHandler struct {
	tweetService TweetService
}

func NewTweetHandler(svc TweetService) *TweetHandler {
	return &TweetHandler{tweetService: svc}
}

func (h *TweetHandler) Create(c *gin.Context) {
	auth, err := GetAuthContext(c)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	var req dto.CreateTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errcode.FilterBindError(err)
		c.JSON(errcode.GetStatusCode(appErr), app.Fail(appErr))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	tweet, err := h.tweetService.PostTweet(
		c.Request.Context(),
		auth.UserID,
		req.Content,
		req.ImageURL,
	)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	c.JSON(http.StatusCreated, app.Success(dto.NewTweetResponse(tweet)))
}

func (h *TweetHandler) Get(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	tweet, err := h.tweetService.FetchTweet(c.Request.Context(), id)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}
	c.JSON(http.StatusOK, app.Success(dto.NewTweetResponse(tweet)))

}
func (h *TweetHandler) Update(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}
	auth, err := GetAuthContext(c)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}
	var req dto.UpdateTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errcode.FilterBindError(err)
		c.JSON(errcode.GetStatusCode(appErr), app.Fail(appErr))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}
	tweet, isChanged, err := h.tweetService.EditTweet(c.Request.Context(), req.Content, id, auth.UserID)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	if !isChanged {
		c.JSON(http.StatusOK, app.SuccessMsg("内容に変更はありません"))
		return
	}

	c.JSON(http.StatusOK, app.Success(dto.NewTweetResponse(tweet)))

}
func (h *TweetHandler) Delete(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}
	auth, err := GetAuthContext(c)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}
	err = h.tweetService.RemoveTweet(c.Request.Context(), id, auth.UserID)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	c.JSON(http.StatusOK, app.SuccessMsg("ツイートの削除成功"))
}
