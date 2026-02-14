package api

import (
	"aita/internal/contextkeys"
	"aita/internal/dto"
	"aita/internal/errcode"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", errcode.ErrSessionNotFound
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		token := strings.TrimSpace(parts[1])
		if token == "" {
			return "", errcode.ErrSessionNotFound
		}
		return token, nil
	}
	return "", errcode.ErrSessionNotFound
}

func GetAuthContext(c *gin.Context) (*dto.AuthContext, error) {
    val, ok := c.Get(contextkeys.AuthPayloadKey)
    if !ok {
        return nil, errcode.ErrSessionNotFound
    }

    auth, ok := val.(*dto.AuthContext)
    if !ok || auth == nil {
        return nil, errcode.ErrSessionNotFound
    }

    return auth, nil
}

func GetIDParam(c *gin.Context, name string)(int64, error) {
	idStr := c.Param(name)
	id, err :=strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errcode.ErrInvalidIDFormat
	}
	return id, nil
}

