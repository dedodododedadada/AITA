package api

import (
	"aita/internal/contextkeys"
	"aita/internal/models"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{
		name			string
		setupAuth       func(req *http.Request)
		setupMock       func(m *MockSessionStore)
		expectedStatus  int
	} {
		{
			name: "認証成功",
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer valid-token")
			},
			setupMock: func(m *MockSessionStore) {
				m.On("GetByToken", mock.Anything, "valid-token").Return(&models.Session{
					UserID: 123,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil)
			},
			expectedStatus: http.StatusOK,

		},
		{
			name: "ヘッダーなし",
			setupAuth: func(req *http.Request) {},
			setupMock: func(m *MockSessionStore) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "認証形式が正しくない (Bearerが不足)",
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "just-a-token")
			},
			setupMock: func(m *MockSessionStore) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "トークンが無効または存在しない",
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalid-token")
			},
			setupMock: func(m *MockSessionStore) {
				m.On("GetByToken", mock.Anything, "invalid-token").Return(nil, errors.New("会話無効"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "セッション期限切れ",
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer out-dated-token")
			},
			setupMock: func(m *MockSessionStore) {
				m.On("GetByToken", mock.Anything, "out-dated-token").Return(&models.Session{
					UserID: 234,
					ExpiresAt: time.Now().Add(-1 *time.Hour), 
				},nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func (t *testing.T) {
			mockStore := new(MockSessionStore)
			tt.setupMock(mockStore)
			r := gin.New()
			r.Use(AuthMiddleware(mockStore))
			r.GET("/protected", func(c *gin.Context) {
				if val, exist := c.Get(contextkeys.AuthPayloadKey); exist {
					uid := val.(int64)
					c.JSON(http.StatusOK, gin.H{"user_id" : uid})
				}
			})
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			tt.setupAuth(req)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockStore.AssertExpectations(t)
		})
	}
}
