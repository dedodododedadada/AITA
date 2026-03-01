package api

import (
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/pkg/app"
	"context"

	"net/http"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	Register(ctx context.Context, username, email,password string) (*dto.UserRecord, error)
	Login(ctx context.Context, email, password string) (*dto.UserRecord, error) 
	ToMyPage(ctx context.Context, userID int64) (*dto.UserRecord, error) 
}

type SessionManager interface {
	Issue(ctx context.Context, userID int64) (*dto.SessionResponse, error)
	Revoke(ctx context.Context, userID int64, token string) error
}

type UserHandler struct {
	userService    UserService
	sessionService SessionManager
}

func NewUserHandler(usvc UserService, sm SessionManager) *UserHandler {
	return &UserHandler{
		userService:    usvc,
		sessionService: sm,
	}
}

func (h *UserHandler) respondWithToken(c *gin.Context, user *dto.UserRecord, statusCode int) {
	response, err := h.sessionService.Issue(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	loginData := dto.LoginResponse{
		SessionToken: response.Token,
		User:         user.ToUserResponse(),
	}
	c.JSON(statusCode, app.Success(loginData))
}

func (h *UserHandler) SignUp(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errcode.FilterBindError(err)
		c.JSON(errcode.GetStatusCode(appErr), app.Fail(appErr))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	user, err := h.userService.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	h.respondWithToken(c, user, http.StatusCreated)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errcode.FilterBindError(err)
		c.JSON(errcode.GetStatusCode(appErr), app.Fail(appErr))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	user, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	h.respondWithToken(c, user, http.StatusOK)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	auth, err := GetAuthContext(c)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	user, err := h.userService.ToMyPage(c.Request.Context(), auth.UserID)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	c.JSON(http.StatusOK, app.Success(user.ToUserProfile()))
}

func (h *UserHandler) Logout(c *gin.Context) {
	auth, err := GetAuthContext(c)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	if err := h.sessionService.Revoke(c.Request.Context(), auth.UserID, auth.Token); err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	c.JSON(http.StatusOK, app.SuccessMsg("ログアウトしました"))
}
