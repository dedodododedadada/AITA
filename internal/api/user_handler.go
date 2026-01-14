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
	userstore db.UserStore
	sessionstore db.SessionStore
}

func NewUserHandler(us db.UserStore, ss db.SessionStore) *UserHandler {
	return &UserHandler{
		userstore: us,
		sessionstore: ss,
	}
}

func (h *UserHandler) respondWithToken(c *gin.Context, user *models.User, statusCode int){
	duration := 24 * time.Hour
	rawToken, _, err := h.sessionstore.Create(c.Request.Context(), user.ID, duration)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なリクエストデータです" +err.Error()})
		return
	}
	user, err := h.userstore.Create(c.Request.Context(), &req) 
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なリクエストデータです"+ err.Error()})
		return	
	}
	user, err := h.userstore.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if err.Error() == "ユーザーが存在しません"{
			c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})//401
		} else{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバー内部エラーが発生しました"})//500
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
