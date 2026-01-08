package api

import(
	"aita/internal/db"
	"aita/internal/models"
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
)

const AuthPayloadKey = "authorization_user_id"

func AuthMiddlare(store db.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です。ログインしてください"})
			return
		}
		fields := string.Fields(authHeader)
		if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証形式が正しくありません"})
			return
		} 
		token := fields[1]
		session, err := store.GetByToken(c.Request.Context(),token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": error.Error()})
			return
		}
		c.Set(AuthPayloadKey, session.UserID)
		c.Next()
	}
}
