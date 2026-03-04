package api

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/pkg/app"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FollowService interface {
	Follow(ctx context.Context, userID, targetID int64) (*dto.FollowRecord, error)
	UnFollow(ctx context.Context, userID, targetID int64) error
	GetFollowers(ctx context.Context, userID int64) ([]*dto.UserSlimRecord, error)
	GetFollowings(ctx context.Context, userID int64) ([]*dto.UserSlimRecord, error)
	GetRelation(ctx context.Context, userID, targetID int64) (*dto.RelationRecord, error)
}

type FollowHandler struct {
	followService FollowService
}

func NewFollowHandler(srv FollowService) *FollowHandler {
	return &FollowHandler{followService: srv}
}


func (h *FollowHandler) Follow(c *gin.Context) {
	auth, err := GetAuthContext(c)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}
	
	var req app.FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	follow, err := h.followService.Follow(c.Request.Context(), auth.UserID, req.TargetID)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	c.JSON(http.StatusCreated, app.Success(follow.ToFollowResponse()))
}


func (h *FollowHandler) UnFollow(c *gin.Context) {
	auth, err := GetAuthContext(c)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}
	
	var req app.UnFollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	err = h.followService.UnFollow(c.Request.Context(), auth.UserID, req.TargetID)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	c.JSON(http.StatusOK, app.SuccessMsg("フォロウィングの削除成功"))
}

func (h *FollowHandler) GetFollowers(c *gin.Context) {
	userID, err := GetIDParam(c, "id")
    if err != nil {
        c.JSON(errcode.GetStatusCode(err), app.Fail(err))
        return
    }

    infos, err := h.followService.GetFollowers(c.Request.Context(), userID)
    if err != nil {
        c.JSON(errcode.GetStatusCode(err), app.Fail(err))
        return
    }

	c.JSON(http.StatusOK, app.Success(infos))
}

func (h *FollowHandler) GetFollowings(c *gin.Context) {
	userID, err := GetIDParam(c, "id")
    if err != nil {
        c.JSON(errcode.GetStatusCode(err), app.Fail(err))
        return
    }

    infos, err := h.followService.GetFollowings(c.Request.Context(), userID)
    if err != nil {
        c.JSON(errcode.GetStatusCode(err), app.Fail(err))
        return
    }

	c.JSON(http.StatusOK, app.Success(infos))
}

func (h *FollowHandler) GetRelation(c *gin.Context) {
	auth, err := GetAuthContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, app.Fail(err))
        return
    }

	targetID, err := GetIDParam(c, "id")
    if err != nil {
        c.JSON(http.StatusBadRequest, app.Fail(err))
        return
    }

	relation, err := h.followService.GetRelation(c.Request.Context(), auth.UserID, targetID)
    if err != nil {
        c.JSON(errcode.GetStatusCode(err), app.Fail(err))
        return
    }

    c.JSON(http.StatusOK, app.Success(relation.ToRelationResponse(auth.UserID, targetID)))
}
