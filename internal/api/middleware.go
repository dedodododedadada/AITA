package api

import (
	"aita/internal/contextkeys"
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/pkg/app"
	"context"

	"github.com/gin-gonic/gin"
)

type AuthSessionService interface {
	Validate(ctx context.Context, token string) (*dto.SessionResponse, error)
}
// Note : add asynchronous sliding expiration
func AuthMiddleware(svc AuthSessionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token, err := extractBearerToken(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(errcode.GetStatusCode(err), app.Fail(err))
			return
		}
		response, err := svc.Validate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(errcode.GetStatusCode(err), app.Fail(err))
			return
		}

		c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext {
			UserID: response.UserID,
			Token: response.Token,
		})
		c.Next()
	}
}
