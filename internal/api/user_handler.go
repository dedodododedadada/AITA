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

type UserService interface {
	Register(ctx context.Context, username string, email string, password string) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.User, error)
	ToMyPage(ctx context.Context, id int64) (*models.User, error)
}

type SessionManager interface {
	Issue(ctx context.Context, userID int64) (string, error)
	Revoke(ctx context.Context, sessionID int64) error
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

func (h *UserHandler) respondWithToken(c *gin.Context, user *models.User, statusCode int) {
	token, err := h.sessionService.Issue(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	loginData := dto.LoginResponse{
		SessionToken: token,
		User:         dto.NewUserResponse(user),
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

	c.JSON(http.StatusOK, app.Success(dto.NewUserResponse(user)))
}

func (h *UserHandler) Logout(c *gin.Context) {
	auth, err := GetAuthContext(c)
	if err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}
	if err := h.sessionService.Revoke(c.Request.Context(), auth.SessionID); err != nil {
		c.JSON(errcode.GetStatusCode(err), app.Fail(err))
		return
	}

	c.JSON(http.StatusOK, app.SuccessMsg("ログアウトしました"))
}
