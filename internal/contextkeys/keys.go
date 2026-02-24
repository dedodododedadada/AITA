package contextkeys

import (
	"aita/internal/dto"

	"github.com/gin-gonic/gin"
)

type contextKeys string

const (
	AuthPayloadKey = "authorization_user_id"
)

func GetAuthPayload(c *gin.Context) (*dto.AuthContext, bool) {
	val, exists := c.Get(AuthPayloadKey)
	if !exists {
		return nil, false
	}
	payload, ok := val.(*dto.AuthContext)
	return payload, ok
}