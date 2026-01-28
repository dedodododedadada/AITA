package api

import (
	"aita/internal/contextkeys"
	"aita/internal/db"

	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(store db.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です。ログインしてください"})
			return
		}
		fields := strings.Fields(authHeader)
		if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証形式が正しくありません"})
			return
		} 
		token := fields[1]
		session, err := store.GetByToken(c.Request.Context(),token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "セッションが無効または期限切れです"})
			return
		}
		if session.IsExpired() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error":"セッションが無効または期限切れです"})
			return
		}
		c.Set(contextkeys.AuthPayloadKey, session.UserID)
		c.Next()
	}
}
