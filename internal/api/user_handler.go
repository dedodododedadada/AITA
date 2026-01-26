package api

import (
	"aita/internal/contextkeys"
	"aita/internal/db"
	"aita/internal/models"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userStore db.UserStore
	sessionStore db.SessionStore
}

func NewUserHandler(US db.UserStore, SS db.SessionStore) *UserHandler {
	return &UserHandler{
		userStore: US,
		sessionStore: SS,
	}
}

func (h *UserHandler) respondWithToken(c *gin.Context, user *models.User, statusCode int) {
	duration := 24 * time.Hour
	rawToken, _, err := h.sessionStore.Create(c.Request.Context(), user.ID, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "セッションの作成に失敗しました"})
		return
	}
	c.JSON(statusCode, models.LoginResponse{
		SessionToken: rawToken,
		User:         models.NewUserResponse(user),
	})
}

func (h *UserHandler) SignUp(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエスト形式が正しくありません"})
		return
	}
	user, err := h.userStore.Create(c.Request.Context(), &req) 
	if err != nil {
		if err.Error() == "ユーザー名かメールアドレスは登録済みです" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) 
		} else {
		    c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザーの作成に失敗しました"}) 
		}
		return
	}
	h.respondWithToken(c, user, http.StatusCreated)
}

func(h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエスト形式が正しくありません"})
		return	
	}
	user, err := h.userStore.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if err.Error() == "ユーザーが存在しません" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
		} else{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバー内部エラーが発生しました"})
		}
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash),[]byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
		return
	}
	h.respondWithToken(c, user, http.StatusOK)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	val, exists := c.Get(contextkeys.AuthPayloadKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証に失敗しました"})
		return
	}
	userID , ok := val.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "無効なユーザーIDタイプです"})
		return
	}
	user, err := h.userStore.GetByID(c.Request.Context(), userID)
	if err != nil{
		if errors.Is(err, db.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ユーザーが存在しません"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバー内部エラーが発生しました"})
		return 
	}
	c.JSON(http.StatusOK,  models.NewUserResponse(user))
}
