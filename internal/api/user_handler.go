package api

import(
	"aita/internal/db"
	"aita/internal/models"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	store db.UserStore
}

func NewUserHandler(store db.UserStore) *UserHandler {
	return &UserHandler{store: store}
}

func (h *UserHandler) SignUp(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なリクエストデータです" +err.Error()})
		return
	}
	user, err := h.store.Create(c, &req) 
	if err != nil {
		if err.Error() == "ユーザー名かメールアドレスは登録済みです" {
			c.JSON(http.StatusConflict, gin.H{"error":err.Error()}) 
		} else {
		    c.JSON(http.StatusInternalServerError, gin.H{"error":"ユーザーの作成に失敗しました"}) 
		}
		return
	}
	c.JSON(http.StatusCreated, models.NewUserResponse(user))
}

func(h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"無効なリクエストデータです"+ err.Error()})
		return	
	}
	user, err := h.store.GetByEmail(c, req.Email)
	if err != nil {
		if err.Error() == "ユーザーが存在しません"{
			c.JSON(http.StatusUnauthorized, gin.H{"error":"メールアドレスまたはパスワードが正しくありません"})//401
		} else{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"サーバー内部エラーが発生しました"})//500
		}
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash),[]byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error":"メールアドレスまたはパスワードが正しくありません"})
		return
	}
	sessionToken :="a_very_secret_and_random_session_token"// will be updated in session
	loginResponse := models.LoginResponse{
		SessionToken: sessionToken,
		User:         models.NewUserResponse(user),
	}
	c.JSON(http.StatusOK,loginResponse)
}
