package api

import (
	"aita/internal/contextkeys"
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
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		setupMock      func(m *mockSessionService)
		expectedStatus int
		expectedUserID int64
	}{
		{
			name:       "認証成功: 有効なBearerトークン",
			authHeader: "Bearer valid_token",
			setupMock: func(m *mockSessionService) {
				session := &models.Session{UserID: 123}
				m.On("Validate", mock.Anything, "Bearer valid_token").Return(session, nil)
			},
			expectedStatus: http.StatusOK,
			expectedUserID: 123,
		},
		{
			name:       "認証成功:大文字でもServiceが許容する場合",
			authHeader: "BEARER raw_token_string",
			setupMock: func(m *mockSessionService) {
				session := &models.Session{UserID: 456}
				m.On("Validate", mock.Anything, "raw_token_string").Return(session, nil)
			},
			expectedStatus: http.StatusOK,
			expectedUserID: 456,
		},
		{
			name:       "未認証：ヘッダーが空",
			authHeader: "",
			setupMock: func(m *mockSessionService) {
				m.On("Validate", mock.Anything, "").Return(nil, models.ErrSessionNotFound)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "認証失敗：トークンが期限切れ",
			authHeader: "Bearer expired_token",
			setupMock: func(m *mockSessionService) {
				m.On("Validate", mock.Anything, "Bearer expired_token").Return(nil, models.ErrSessionExpired)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "ユーザー不在：トークンは正しいがユーザーが削除された",
			authHeader: "Bearer valid_but_no_user",
			setupMock: func(m *mockSessionService) {
				m.On("Validate", mock.Anything, "Bearer valid_but_no_user").Return(nil, models.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound, 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mSession := new(mockSessionService)
			tt.setupMock(mSession)

			w := httptest.NewRecorder()
			r := gin.New()

			r.Use(AuthMiddleware(mSession))

			r.GET("/test", func(c *gin.Context) {
				userID, exists := c.Get(contextkeys.AuthPayloadKey)
				if exists {
					c.JSON(http.StatusOK, gin.H{"user_id": userID})
				} else {
					c.Status(http.StatusOK)
				}
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]int64
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.expectedUserID, resp["user_id"])
			}

			mSession.AssertExpectations(t)
		})
	}
}