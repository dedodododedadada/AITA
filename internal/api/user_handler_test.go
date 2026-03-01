package api

import (
	"aita/internal/contextkeys"
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/pkg/app"
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
)

func TestSignUp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		requestBody    any
		setupMock      func(mu *mockUserService, ms *mockSessionService)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "ユーザー登録成功",
			requestBody: dto.SignupRequest{
				Username: "mock_user",
				Email:    "taro@example.com",
				Password: "password123",
			},
			setupMock: func(mu *mockUserService, ms *mockSessionService) {
				record := &dto.UserRecord{
					ID:           1,
					Username:     "mock_User",
					Email:        "taro@example.com",
					PasswordHash: "password123hash",
					CreatedAt:    time.Now().UTC(),
				}
				sessionResp := &dto.SessionResponse{
					UserID: record.ID,
					Token: "valid_token_string",
				}
				mu.On("Register", mock.Anything, mock.MatchedBy(func(username string) bool {
					return username == "mock_user"
				}), mock.MatchedBy(func(email string) bool {
					return email == "taro@example.com"
				}), mock.MatchedBy(func(password string) bool {
					return password == "password123"
				})).Return(record, nil)
				ms.On("Issue", mock.Anything, record.ID).Return(sessionResp, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				data := resp.Data.(map[string]any)
				assert.Equal(t, "valid_token_string", data["session_token"])
			},
		},
		{
			name:           "リクエスト形式不正: 無効なJSONを送信した場合",
			requestBody:    `{"username": "incomplete_json`,
			setupMock:      func(mu *mockUserService, ms *mockSessionService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)

				assert.Equal(t, "INVALID_JSON_FORMAT", resp.Code)
				assert.Equal(t, "JSONの構文が正しくありません", resp.Error)
			},
		},
		{
			name: "バリデーションエラー：メールアドレス重複",
			requestBody: dto.SignupRequest{
				Username: "mock_user",
				Email:    "exists@example.com",
				Password: "password123",
			},
			setupMock: func(mu *mockUserService, ms *mockSessionService) {
				mu.On("Register", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errcode.ErrEmailConflict)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)

				assert.Equal(t, "EMAIL_CONFLICT", resp.Code)
				assert.Contains(t, resp.Error, "既に使用されています")
			},
		},
		{
			name: "セッション発行失敗",
			requestBody: dto.SignupRequest{
				Username: "error_user",
				Email:    "issue_fail@test.com",
				Password: "password123",
			},
			setupMock: func(mu *mockUserService, ms *mockSessionService) {
				record := &dto.UserRecord{
					ID:           50,
					Username:     "error_user",
					Email:        "issue_fail@test.com",
					PasswordHash: "password123hashed",
					CreatedAt:    time.Now().UTC(),
				}
				mu.On("Register", mock.Anything, mock.MatchedBy(func(username string) bool {
					return username == "error_user"
				}), mock.MatchedBy(func(email string) bool {
					return email == "issue_fail@test.com"
				}), mock.MatchedBy(func(password string) bool {
					return password == "password123"
				})).Return(record, nil)
				ms.On("Issue", mock.Anything, int64(50)).Return(nil, errors.New("redis connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, "INTERNAL_SERVER_ERROR", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, ms := new(mockUserService), new(mockSessionService)
			h := NewUserHandler(mu, ms)
			tt.setupMock(mu, ms)

			var buf bytes.Buffer
			if s, ok := tt.requestBody.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tt.requestBody)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/signup", &buf)

			h.SignUp(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		requestBody    any
		setupMock      func(mu *mockUserService, ms *mockSessionService)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "ログイン成功",
			requestBody: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(mu *mockUserService, ms *mockSessionService) {
				user := &dto.UserRecord{ID: 1, Email: "test@example.com"}
				sessionResp := &dto.SessionResponse{
					UserID: user.ID,
					Token: "valid_token_string",
				}
				mu.On("Login", mock.Anything, "test@example.com", "password123").Return(user, nil)
				ms.On("Issue", mock.Anything, user.ID).Return(sessionResp, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				data := resp.Data.(map[string]any)
				assert.Equal(t, "valid_token_string", data["session_token"])
			},
		},
		{
			name:           "JSON構文エラー",
			requestBody:    `{"email": "bad-json"...`,
			setupMock:      func(mu *mockUserService, ms *mockSessionService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, "INVALID_JSON_FORMAT", resp.Code)
			},
		},
		{
			name: "メールアドレスまたはパスワードが間違っている場合",
			requestBody: dto.LoginRequest{
				Email:    "wrong@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(mu *mockUserService, ms *mockSessionService) {
				mu.On("Login", mock.Anything, mock.Anything, mock.Anything).Return(nil, errcode.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, "INVALID_CREDENTIALS", resp.Code)
				assert.Contains(t, resp.Error, "正しくありません")
			},
		},
		{
			name: "トークン発行失敗",
			requestBody: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(mu *mockUserService, ms *mockSessionService) {
				user := &dto.UserRecord{ID: 1, Email: "test@example.com"}
				mu.On("Login", mock.Anything, "test@example.com", "password123").Return(user, nil)
				ms.On("Issue", mock.Anything, user.ID).Return(nil, errors.New("internal server error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, "INTERNAL_SERVER_ERROR", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu, ms := new(mockUserService), new(mockSessionService)
			h := NewUserHandler(mu, ms)
			tt.setupMock(mu, ms)

			var buf bytes.Buffer
			if s, ok := tt.requestBody.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tt.requestBody)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/login", &buf)

			h.Login(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestGetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		setupMock      func(mu *mockUserService)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "マイページ取得成功",
			setupContext: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext{UserID: 101,Token: "Valid_token"})
			},
			setupMock: func(mu *mockUserService) {
				user := &dto.UserRecord{ID: 101, Username: "test_user", Email: "test@example.com", FollowerCount: 10, FollowingCount: 20}
				mu.On("ToMyPage", mock.Anything, int64(101)).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				data := resp.Data.(map[string]any)
				assert.Equal(t, "test_user", data["username"])
				assert.EqualValues(t, 101, data["id"])
				assert.EqualValues(t, 10, data["follower_count"])
				assert.EqualValues(t, 20, data["following_count"])
			},
		},
		{
			name:           "未認証エラー",
			setupContext:   func(c *gin.Context) {},
			setupMock:      func(mu *mockUserService) {},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, "SESSION_NOT_FOUND", resp.Code)
			},
		},
		{
			name: "ユーザー不在",
			setupContext: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext{UserID: 404, Token: "unexisted_token"})
			},
			setupMock: func(mu *mockUserService) {
				mu.On("ToMyPage", mock.Anything, int64(404)).Return(nil, errcode.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, "USER_NOT_FOUND", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := new(mockUserService)
			h := NewUserHandler(mu, nil)
			tt.setupMock(mu)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/me", nil)
			tt.setupContext(c)

			h.GetMe(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			mu.AssertExpectations(t)
		})
	}
}

func TestUserHandlerLogout(t *testing.T) {

	gin.SetMode(gin.TestMode)


	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		setupMock      func(ms *mockSessionService)
		expectedStatus int
		expectMsg      string
	}{
		{
			name: "ログアウト成功",
			setupContext: func(c *gin.Context) {
				auth := &dto.AuthContext{
					UserID: 101,
					Token: "valid_token",
				}
				c.Set(contextkeys.AuthPayloadKey, auth)
			},
			setupMock: func(ms *mockSessionService) {
				ms.On("Revoke", mock.Anything, mock.MatchedBy(func(userID int64) bool {
					return userID == 101
				}), mock.MatchedBy(func(token string) bool {
					return token == "valid_token"
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectMsg:      "ログアウトしました",
		},
		{
			name:           "セッションが見たからない",
			setupContext:   func(c *gin.Context) {},
			setupMock:      func(ms *mockSessionService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "サーバー内部エラー",
			setupContext: func(c *gin.Context) {
				auth := &dto.AuthContext{
					UserID: 101,
					Token: "valid_token",
				}
				c.Set(contextkeys.AuthPayloadKey, auth)
			},
			setupMock: func(ms *mockSessionService) {
				ms.On("Revoke", mock.Anything, mock.MatchedBy(func(userID int64) bool {
					return userID == 101
				}), mock.MatchedBy(func(token string) bool {
					return token == "valid_token"
				})).Return(errcode.ErrSessionNotFound)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ms := new(mockSessionService)
			h := NewUserHandler(nil, ms)

			tt.setupMock(ms)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request, _ = http.NewRequest(http.MethodPost, "/logout", nil)

			tt.setupContext(c)

			h.Logout(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectMsg != "" {
				assert.Contains(t, w.Body.String(), tt.expectMsg)
			}

			ms.AssertExpectations(t)
		})
	}
}
