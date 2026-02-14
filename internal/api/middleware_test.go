package api

import (
	"aita/internal/contextkeys"
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
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

    session123 := &models.Session{ID: 10, UserID: 123}
    session456 := &models.Session{ID: 20, UserID: 456}

    tests := []struct {
        name            string
        authHeader      string
        setupMock       func(m *mockSessionService)
        expectedStatus  int
        expectedSession *models.Session 
    }{
        {
            name:       "【成功】有効なBearerトークン",
            authHeader: "Bearer valid_token",
            setupMock: func(m *mockSessionService) {
                m.On("Validate", mock.Anything, "valid_token").Return(session123, nil)
            },
            expectedStatus:  http.StatusOK,
            expectedSession: session123,
        },
        {
            name:       "【成功】大文字のBEARERでも認識される",
            authHeader: "BEARER upper_token",
            setupMock: func(m *mockSessionService) {
                m.On("Validate", mock.Anything, "upper_token").Return(session456, nil)
            },
            expectedStatus:  http.StatusOK,
            expectedSession: session456,
        },
        {
            name:            "【失败】ヘッダーが空 (Serviceは呼ばれない)",
            authHeader:      "",
            setupMock:       func(m *mockSessionService) {},
            expectedStatus:  http.StatusUnauthorized,
            expectedSession: nil,
        },
        {
            name:            "【失败】Bearerプレフィックスがない (Serviceは呼ばれない)",
            authHeader:      "just_token_without_bearer",
            setupMock:       func(m *mockSessionService) {},
            expectedStatus:  http.StatusUnauthorized,
            expectedSession: nil,
        },
        {
            name:            "【失败】Bearerのみでトークンが空 (Serviceは呼ばれない)",
            authHeader:      "Bearer ",
            setupMock:       func(m *mockSessionService) {},
            expectedStatus:  http.StatusUnauthorized,
            expectedSession: nil,
        },
        {
            name:            "【失败】スキームがBasic (Serviceは呼ばれない)",
            authHeader:      "Basic dXNlcjpwYXNz",
            setupMock:       func(m *mockSessionService) {},
            expectedStatus:  http.StatusUnauthorized,
            expectedSession: nil,
        },
        {
            name:       "【失败】トークンが期限切れ (extractは成功するがServiceで失敗)",
            authHeader: "Bearer expired_token",
            setupMock: func(m *mockSessionService) {
                m.On("Validate", mock.Anything, "expired_token").Return(nil, errcode.ErrSessionExpired)
            },
            expectedStatus:  http.StatusUnauthorized,
            expectedSession: nil,
        },
        {
            name:       "【失败】ユーザー不在 (404 Not Foundを返す映射を確認)",
            authHeader: "Bearer ghost_user_token",
            setupMock: func(m *mockSessionService) {
                m.On("Validate", mock.Anything, "ghost_user_token").Return(nil, errcode.ErrUserNotFound)
            },
            expectedStatus:  http.StatusNotFound,
            expectedSession: nil,
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

            if tt.expectedStatus == http.StatusOK && tt.expectedSession != nil {
                var resp dto.AuthContext
                json.Unmarshal(w.Body.Bytes(), &resp)
                
                assert.Equal(t, tt.expectedSession.UserID, resp.UserID, "UserIDが正しく変換されていること")
                assert.Equal(t, tt.expectedSession.ID, resp.SessionID, "SessionIDが正しく変換されていること")
            }

            ms.AssertExpectations(t)
        })
    }
}