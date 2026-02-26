package api

import (
	"aita/internal/contextkeys"
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/pkg/app"
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthSessionService interface {
	Validate(ctx context.Context, token string) (*dto.SessionResponse, error)
	ShouldRefresh(expiresAt, createdAt time.Time) (bool, error)
	RefreshAsync(token string) 
}

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

		if should, _ := svc.ShouldRefresh(response.ExpiresAt, response.CreatedAt); should {
            svc.RefreshAsync(token)
        } 

		c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext {
			UserID: response.UserID,
			Token: response.Token,
		})

		c.Next()
	}
}
