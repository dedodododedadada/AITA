package api

import (
	"aita/internal/contextkeys"
	"aita/internal/models"
	"context"

	"github.com/gin-gonic/gin"
)

type AuthSessionService interface {
	Validate(ctx context.Context, token string) (*models.Session, error)
}

func AuthMiddleware(svc AuthSessionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		session, err := svc.Validate(c.Request.Context(), authHeader)
		if err != nil {
			c.AbortWithStatusJSON(models.GetStatusCode(err), models.Fail(err))
			return
		}
		c.Set(contextkeys.AuthPayloadKey, session.UserID)
		c.Next()
	}
}
