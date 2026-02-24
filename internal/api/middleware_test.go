package api

import (
	"aita/internal/contextkeys"
	"aita/internal/dto"
	"aita/internal/errcode"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthMiddleware(t *testing.T) {
    gin.SetMode(gin.TestMode)
    resp123 := &dto.SessionResponse{UserID: 123, Token: "valid_token1"}
    resp456 := &dto.SessionResponse{UserID: 456, Token: "valid_token6"}
    Auth123 := &dto.AuthContext{UserID: 123, Token: "valid_token1"}
    Auth456 := &dto.AuthContext{UserID: 456, Token: "valid_token6"}
    tests := []struct {
        name            string
        authHeader      string
        setupMock       func(m *mockSessionService)
        expectedStatus  int
        expectedAuth    *dto.AuthContext
    }{
        {
            name:       "【成功】有効なBearerトークン",
            authHeader: "Bearer valid_token",
            setupMock: func(m *mockSessionService) {
                m.On("Validate", mock.Anything, "valid_token").Return(resp123, nil)
            },
            expectedStatus:  http.StatusOK,
            expectedAuth: Auth123,
        },
        {
            name:       "【成功】大文字のBEARERでも認識される",
            authHeader: "BEARER upper_token",
            setupMock: func(m *mockSessionService) {
                m.On("Validate", mock.Anything, "upper_token").Return(resp456, nil)
            },
            expectedStatus:  http.StatusOK,
            expectedAuth: Auth456,
        },
        {
            name:            "【失败】ヘッダーが空 (Serviceは呼ばれない)",
            authHeader:      "",
            setupMock:       func(m *mockSessionService) {},
            expectedStatus:  http.StatusUnauthorized,
            expectedAuth: nil,
        },
        {
            name:            "【失败】Bearerプレフィックスがない (Serviceは呼ばれない)",
            authHeader:      "just_token_without_bearer",
            setupMock:       func(m *mockSessionService) {},
            expectedStatus:  http.StatusUnauthorized,
            expectedAuth: nil,
        },
        {
            name:            "【失败】Bearerのみでトークンが空 (Serviceは呼ばれない)",
            authHeader:      "Bearer ",
            setupMock:       func(m *mockSessionService) {},
            expectedStatus:  http.StatusUnauthorized,
            expectedAuth: nil,
        },
        {
            name:            "【失败】スキームがBasic (Serviceは呼ばれない)",
            authHeader:      "Basic dXNlcjpwYXNz",
            setupMock:       func(m *mockSessionService) {},
            expectedStatus:  http.StatusUnauthorized,
            expectedAuth: nil,
        },
        {
            name:       "【失败】トークンが期限切れ (extractは成功するがServiceで失敗)",
            authHeader: "Bearer expired_token",
            setupMock: func(m *mockSessionService) {
                m.On("Validate", mock.Anything, "expired_token").Return(nil, errcode.ErrSessionExpired)
            },
            expectedStatus:  http.StatusUnauthorized,
            expectedAuth: nil,
        },
        {
            name:       "【失败】ユーザー不在 (404 Not Foundを返す映射を確認)",
            authHeader: "Bearer ghost_user_token",
            setupMock: func(m *mockSessionService) {
                m.On("Validate", mock.Anything, "ghost_user_token").Return(nil, errcode.ErrUserNotFound)
            },
            expectedStatus:  http.StatusNotFound,
            expectedAuth: nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ms := new(mockSessionService)
            tt.setupMock(ms)

            w := httptest.NewRecorder()
            r := gin.New()

            r.Use(AuthMiddleware(ms))

            r.GET("/test", func(c *gin.Context) {
                val, exists := c.Get(contextkeys.AuthPayloadKey)
                if exists {
                    auth, ok := val.(*dto.AuthContext)
                    if ok {
                        c.JSON(http.StatusOK, auth)
                        return
                    }
                }
                c.Status(http.StatusOK)
            })

            req := httptest.NewRequest(http.MethodGet, "/test", nil)
            if tt.authHeader != "" {
                req.Header.Set("Authorization", tt.authHeader)
            }
            r.ServeHTTP(w, req)

            assert.Equal(t, tt.expectedStatus, w.Code)

            if tt.expectedStatus == http.StatusOK && tt.expectedAuth != nil {
                var auth dto.AuthContext
                json.Unmarshal(w.Body.Bytes(), &auth)
                
                assert.Equal(t, tt.expectedAuth.UserID, auth.UserID, "UserIDが正しく変換されていること")
                assert.Equal(t, tt.expectedAuth.Token, auth.Token, "Tokenが正しく変換されていること")
            }

            ms.AssertExpectations(t)
        })
    }
}