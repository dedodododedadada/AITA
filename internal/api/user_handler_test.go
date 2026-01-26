package api

import (
	"aita/internal/contextkeys"
	"aita/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserHandlerSignUp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{
		name			string
		inputBody  		models.SignupRequest
		setupMock   	func(mu *MockUserStore, ms *MockSessionStore)
		expectedStatus  int
	}{
		{
			name: "登録成功",
			inputBody: models.SignupRequest{
				Username: "mock_user",
				Email: "mock@example.com",
				Password: "password101",
			},
			setupMock: func(mu *MockUserStore, ms *MockSessionStore) {
				mu.On("Create", mock.Anything, mock.MatchedBy(func(req *models.SignupRequest) bool {
					return req.Username == "mock_user"
				})).Return(&models.User{
					ID: 1, 
					Username: "mock_user",
				}, nil)
				ms.On("Create", mock.Anything, int64(1), 24*time.Hour).Return("mock_token", &models.Session{}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "パラメーターエラー、ユーザー名が短すぎる",
			inputBody: models.SignupRequest{
				Username: "abc",
				Email: "mock2@example.com",
				Password: "password102",
			},
			setupMock: func(mu *MockUserStore, ms *MockSessionStore) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "パラメーターエラー、パスワードが短すぎる",
			inputBody: models.SignupRequest{
				Username: "userMock",
				Email: "mock3@example.com",
				Password: "123456",
			},
			setupMock: func(mmu *MockUserStore, ms *MockSessionStore) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "パラメーターエラー、メール形式が正しくない",
			inputBody: models.SignupRequest{
				Username: "userMock",
				Email: "mock3examplecom",
				Password: "12345678",
			},
			setupMock: func(mu *MockUserStore, ms *MockSessionStore) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "データベースエラー、ユーザーが既に存在",
			inputBody: models.SignupRequest{
				Username: "mock_user_2",
				Email: "mock@example.com",
				Password: "password101",
			},
			setupMock: func(mu *MockUserStore, ms *MockSessionStore) {
				mu.On("Create", mock.Anything, mock.MatchedBy(func(req *models.SignupRequest) bool {
					return req.Email == "mock@example.com"
				})).Return(nil, errors.New("ユーザー名かメールアドレスは登録済みだ"))
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := new(MockUserStore)
			ms := new(MockSessionStore)
			tt.setupMock(mu, ms)
			handler := NewUserHandler(mu, ms)
			r := gin.New()
			r.POST("/signup", handler.SignUp)
			body, _ := json.Marshal(tt.inputBody)
			req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t,tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
    			var resp map[string]interface{}
    			err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err, "エラーが発生しないこと")
    			assert.NotEmpty(t, resp["token"]) 
				userData, ok := resp["user"].(map[string]interface{})
				if assert.True(t, ok, "レスポンスにはユーザーオブジェクトが存在すること") {
					assert.Equal(t, "mock_user", resp["user"].(map[string]interface{})["username"])
				}
    			assert.Equal(t, "mock_user", userData["username"])
			}
			mu.AssertExpectations(t)
			ms.AssertExpectations(t)
		})
	}
} 

func TestUserHandlerLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	tests := []struct {
		name			string
		inputBody       models.LoginRequest
		setupMock       func(mu *MockUserStore, ms *MockSessionStore)
		expectedStatus  int  
	}{
		{
			name: "ログイン成功",
			inputBody: models.LoginRequest{
				Email: "user@example.com",
				Password: "secret123",
			},
			setupMock: func(mu *MockUserStore, ms *MockSessionStore) {
				mu.On("GetByEmail", mock.Anything, "user@example.com").Return(&models.User{
					ID: 101,
					Email: "user@example.com",
					PasswordHash: string(hashedPassword),
				}, nil)
				ms.On("Create", mock.Anything, int64(101), 24*time.Hour).Return("mock_token", &models.Session{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ログイン失敗、ユーザーが見つからない",
			inputBody: models.LoginRequest{
				Email: "notfound@example.com",
				Password: "anystring",
			},
			setupMock: func(mu *MockUserStore, ms *MockSessionStore)  {
				mu.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, errors.New("ユーザーが存在しない"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "ログイン失敗、パスワード間違い",
			inputBody: models.LoginRequest{
				Email: "user@example.com",
				Password: "wrong_password",
			},
			setupMock: func(mu *MockUserStore, ms *MockSessionStore) {
				mu.On("GetByEmail", mock.Anything, "user@example.com").Return(&models.User{
					PasswordHash : string(hashedPassword),
				}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "ログイン失敗、セッション生成エラー",
			inputBody: models.LoginRequest{
				Email: "user@example.com",
				Password: "secret123",
			},
			setupMock: func(mu *MockUserStore, ms *MockSessionStore) {
				mu.On("GetByEmail", mock.Anything, "user@example.com").Return(&models.User{
					ID: 101,
					PasswordHash: string(hashedPassword),
				}, nil)
				ms.On("Create", mock.Anything, int64(101), 24*time.Hour).Return("", nil, errors.New("セッションの作成に失敗"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := new(MockUserStore)
			ms := new(MockSessionStore)
			tt.setupMock(mu, ms)
			handler := NewUserHandler(mu, ms)
			r := gin.New()
			r.POST("/login", handler.Login)
			body, _ := json.Marshal(tt.inputBody)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code) 
			if tt.expectedStatus == http.StatusOK {
    			var resp map[string]interface{}
    			json.Unmarshal(w.Body.Bytes(), &resp)
   				assert.NotEmpty(t, resp["token"]) 
			}
			mu.AssertExpectations(t)
			ms.AssertExpectations(t)
		})
	}
}

func TestUserHandlerGetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct{
		name			string
		setupContext    func(c *gin.Context)
		setupMock       func(mu *MockUserStore)
		expectedStatus  int
	}{
		{
			name:  "ユーザー情報を正常に取得できること",
			setupContext: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, int64(101))
			},
			setupMock: func(mu *MockUserStore) {
				mu.On("GetByID", mock.Anything, int64(101)).Return(&models.User{
					ID:	101,
					Username: "mock_user",
					Email: "mock@example.com",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ContextにユーザーIDが存在しない場合",
			setupContext: func(c *gin.Context) {},
			setupMock:    func(mu *MockUserStore) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "ユーザーが存在しない場合",
			setupContext: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, int64(999))
			},
			setupMock: func(mu *MockUserStore) {
				mu.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.New("なし"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := new(MockUserStore)
			ms := new(MockSessionStore)
			tt.setupMock(mu)
			handler := NewUserHandler(mu, ms)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/me", nil)
			tt.setupContext(c)
			handler.GetMe(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
                var resp map[string]interface{}
                json.Unmarshal(w.Body.Bytes(), &resp)
                assert.Equal(t, "mock_user", resp["username"])
                assert.Nil(t, resp["password_hash"])
            }
			mu.AssertExpectations(t)
		})
	}
}


