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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := map[string]struct{
		setupContext    func(c *gin.Context)
		requestBody     string 		     
		setupMock   	func(mts *mockTweetService)
		expectedStatus  int
		expectedBody    string
	}{
		"ツイート生成に成功した": {
			setupContext: func (c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, int64(101))
			},
			requestBody: `{"content": "Test Message"}`,
			setupMock: func(mts *mockTweetService) {
				mts.On("PostTweet", mock.Anything, int64(101), mock.Anything).
				Return(&models.Tweet{ID: 101, Content:"Test Message"}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: "",
		},
		"ユーザー未認証": {
			setupContext: func (c *gin.Context) {},
			requestBody: ``,
			setupMock: func(mts *mockTweetService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:  "未認証です",
		},
		"ユーザーIDの型が正しくない": {
			setupContext: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, "101")
			},
			requestBody: `{"content": "Test Message"}`,
			setupMock: func(mts *mockTweetService) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: "ユーザーIDの型が正しくありません",
		},
		"リクエスト形式が正しくない": {
			setupContext: func (c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, int64(101))
			},
			requestBody: `{"bad_json": `,
			setupMock: func(mts *mockTweetService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: "リクエスト形式が正しくありません",
		},
		"ツイートが空である": {
			setupContext: func (c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, int64(101))
			},
			requestBody: `{"content": ""}`,
			setupMock: func(mts *mockTweetService) {
				mts.On("PostTweet", mock.Anything, int64(101), mock.Anything).
				Return(nil, models.ErrRequiredFieldMissing)
			},
			expectedStatus:http.StatusBadRequest,
			expectedBody: "ツイート内容を入力してください",
		},
		"ユーザーが存在しない": {
			setupContext: func (c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, int64(99999))
			},
			requestBody: `{"content": "Test Message"}`,
			setupMock: func(mts *mockTweetService) {
				mts.On("PostTweet", mock.Anything, int64(99999), mock.Anything).
				Return(nil, models.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: "ユーザーが存在しません",
		},
		"サーバー内部エラー": {
			setupContext: func (c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, int64(101))
			},
			requestBody: `{"content": "Test Message"}`,
			setupMock: func(mts *mockTweetService) {
				mts.On("PostTweet", mock.Anything, int64(101), mock.Anything).
				Return(nil, errors.New("サーバー内部のエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: "サーバー内部エラーが発生しました",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T){
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			mockSvc := new(mockTweetService)
			handler := NewTweetHandler(mockSvc)
			
			ctx.Request, _ = http.NewRequest("POST", "./tweet", bytes.NewBufferString(tt.requestBody))
			ctx.Request.Header.Set("Content-Type", "application/json")

			tt.setupContext(ctx)
			tt.setupMock(mockSvc)

			handler.Create(ctx)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedBody != "" {
				var actual map[string]string
				json.Unmarshal(recorder.Body.Bytes(), &actual)
				assert.Equal(t, tt.expectedBody, actual["error"])
			} else {
				var resp models.TweetResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &resp)
				assert.NoError(t, err, "レスポンスは有効なJSON形式である必要がある")
				assert.Equal(t, int64(101), resp.ID)
				assert.Equal(t, "Test Message", resp.Content)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}