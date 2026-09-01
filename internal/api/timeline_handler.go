package api

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/pkg/app"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TimeLineService interface {
	GetHomeTimeLine(ctx context.Context, userID int64, page, size int) ([]*dto.TweetRecord, error)
}

type TimeLineHandler struct {
	timeLineService TimeLineService
}

func NewTimeLineHandler(s TimeLineService) *TimeLineHandler {
	return &TimeLineHandler{timeLineService: s}
}

func (h *TimeLineHandler) GetHomeTimeLine(c *gin.Context) {
	auth, err := GetAuthContext(c)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "0"))
	if err != nil || page < 0 {
		page = 0
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "20"))
	if err != nil || size <= 0 || size > 100 {
		size = 20
	}

	tweets, err := h.timeLineService.GetHomeTimeLine(c.Request.Context(), auth.UserID, page, size)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	responses := make([]*app.TweetResponse, 0, len(tweets))
	for _, tweet := range tweets {
		if tweet != nil {
			responses = append(responses, tweet.ToTweetResponse())
		}
	}

	c.JSON(http.StatusOK, app.Success(responses))
}
