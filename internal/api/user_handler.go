package api

import (
	"aita/internal/contextkeys"
	"aita/internal/models"
	"context"

	"net/http"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	Register(ctx context.Context, req *models.SignupRequest) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.User, error) 
	ToMyPage(ctx context.Context, id int64) (*models.User, error)             
}

type SessionManageer interface {
    Issue(ctx context.Context, userID int64) (*models.IssueResponse, error)
    Revoke(ctx context.Context, token string) error 
}

type UserHandler struct {
	userService    UserService
	sessionService SessionManageer
}

func NewUserHandler(usvc UserService, sm SessionManageer) *UserHandler {
	return &UserHandler{
		userService:    usvc,
		sessionService: sm,
	}
}

func (h *UserHandler) respondWithToken(c *gin.Context, user *models.User, statusCode int) {
	issueResponse, err := h.sessionService.Issue(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(models.GetStatusCode(err), models.Fail(err))
		return
	}

	loginData := models.LoginResponse{
        SessionToken: issueResponse.Token,
        User:         models.NewUserResponse(user),
    }
    c.JSON(statusCode, models.Success(loginData))
}

func (h *UserHandler) SignUp(c *gin.Context) {
     var req models.SignupRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        appErr := models.FilterBindError(err)
        c.JSON(models.GetStatusCode(appErr), models.Fail(appErr))
        return
    }

    user, err := h.userService.Register(c.Request.Context(), &req)
    if err != nil {
        c.JSON(models.GetStatusCode(err), models.Fail(err))
        return
    }
    
    h.respondWithToken(c, user, http.StatusCreated)
}

func (h *UserHandler) Login(c *gin.Context) {
    var req models.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        appErr := models.FilterBindError(err)
        c.JSON(models.GetStatusCode(appErr), models.Fail(appErr))
        return  
    }

    user, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        c.JSON(models.GetStatusCode(err), models.Fail(err))
        return
    }

    h.respondWithToken(c, user, http.StatusOK)
}

func (h *UserHandler) GetMe(c *gin.Context) {
    val, exists := c.Get(contextkeys.AuthPayloadKey)
    userID, ok := val.(int64)
    if !exists || !ok {
        err := models.ErrSessionNotFound
        c.JSON(models.GetStatusCode(err), models.Fail(err))
        return
    }

    user, err := h.userService.ToMyPage(c.Request.Context(), userID)
    if err != nil {
        c.JSON(models.GetStatusCode(err), models.Fail(err))
        return
    }

    c.JSON(http.StatusOK, models.Success(models.NewUserResponse(user)))
}

func (h *UserHandler) Logout(c *gin.Context) {
    authHeader := c.GetHeader("Authorization")

    if err := h.sessionService.Revoke(c.Request.Context(), authHeader); err != nil {
        c.JSON(models.GetStatusCode(err), models.Fail(err))
    }

    c.JSON(http.StatusOK, models.SuccessMsg("ログアウトしました"))
}
