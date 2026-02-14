package api

import (
	"aita/internal/contextkeys"
	"aita/internal/dto"
	"aita/internal/errcode"
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
		token, err := extractBearerToken(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(dto.GetStatusCode(err), dto.Fail(err))
			return
		}
		session, err := svc.Validate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(dto.GetStatusCode(err), dto.Fail(err))
			return
		}

		if session == nil {
			err := errcode.ErrSessionNotFound
			c.AbortWithStatusJSON(dto.GetStatusCode(err), dto.Fail(err))
			return
		}

		c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext{
			UserID: session.UserID,
			SessionID: session.ID,
		})
		c.Next()
	}
}
