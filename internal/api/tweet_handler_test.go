package api

import (
	"aita/internal/contextkeys"
	"aita/internal/dto"
	"aita/internal/errcode"
	"aita/internal/models"
	"aita/internal/pkg/app"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTweetCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		requestBody    any
		setupAuth      func(c *gin.Context)
		setupMock      func(mt *mockTweetService)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:        "ツイート投稿成功",
			requestBody: dto.CreateTweetRequest{Content: "AITAの初投稿!"},
			setupAuth: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext{UserID: 10})
			},
			setupMock: func(mt *mockTweetService) {
				mt.On(
					"PostTweet",
					mock.Anything,
					int64(10),
					"AITAの初投稿!",
					mock.Anything,
				).Return(&models.Tweet{
					ID:        100,
					Content:   "AITAの初投稿!",
					UserID:    10,
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				fmt.Println("Raw Response:", w.Body.String())
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				data := resp.Data.(map[string]any)
				assert.EqualValues(t, 100, (data["id"]))
				assert.Equal(t, "AITAの初投稿!", data["content"])
			},
		},
		{
			name:           "未認証エラー:ContextにIDがない",
			requestBody:    dto.CreateTweetRequest{Content: "Hello"},
			setupAuth:      func(c *gin.Context) {},
			setupMock:      func(mt *mockTweetService) {},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, "SESSION_NOT_FOUND", resp.Code)
			},
		},
		{
			name:        "JSON構文エラー",
			requestBody: `{"content": "incomplete...`,
			setupAuth: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext{UserID: 10})
			},
			setupMock:      func(mt *mockTweetService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, "INVALID_JSON_FORMAT", resp.Code)
			},
		},
		{
			name:        "バリデーションエラー：内容が空",
			requestBody: dto.CreateTweetRequest{Content: ""},
			setupAuth: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext{UserID: 10})
			},
			setupMock: func(mt *mockTweetService) {
				mt.On("PostTweet", mock.Anything, int64(10), mock.Anything).Return(nil, errcode.ErrRequiredFieldMissing)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, "REQUIRED_FIELD_MISSING", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := new(mockTweetService)
			h := NewTweetHandler(mt)
			tt.setupMock(mt)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var buf bytes.Buffer
			if s, ok := tt.requestBody.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tt.requestBody)
			}
			c.Request = httptest.NewRequest(http.MethodPost, "/tweets", &buf)

			tt.setupAuth(c)
			h.Create(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestTweetGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		tweetID        string
		setupMock      func(mt *mockTweetService)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:    "ツイート取得成功",
			tweetID: "100",
			setupMock: func(mt *mockTweetService) {
				mt.On("FetchTweet", mock.Anything, int64(100)).Return(&models.Tweet{
					ID: 100, Content: "テスト取得", UserID: 10,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				fmt.Printf("\n--- [DEBUG] Raw JSON Response ---\n%s\n----------------------------------\n", w.Body.String())
				var resp app.Response
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err, "JSONのパースに失敗しました")
				if w.Code != http.StatusOK {
					return
				}
				data, ok := resp.Data.(map[string]any)
				if !ok {
					t.Fatalf("Data should be a map, but got %T: %+v", resp.Data, resp.Data)
				}
				assert.EqualValues(t, 100, data["id"])
			},
		},
		{
			name:    "エラー: 存在しないID",
			tweetID: "999",
			setupMock: func(mt *mockTweetService) {
				mt.On("FetchTweet", mock.Anything, int64(999)).Return(nil, errcode.ErrTweetNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, "TWEET_NOT_FOUND", resp.Code)
			},
		},
		{
			name:           "エラー: 無効なID形式",
			tweetID:        "abc",
			setupMock:      func(mt *mockTweetService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, "INVALID_ID_FORMAT", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := new(mockTweetService)
			h := NewTweetHandler(mt)
			tt.setupMock(mt)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Params = []gin.Param{{Key: "id", Value: tt.tweetID}}

			h.Get(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestTweetUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		tweetID        string
		requestBody    any
		setupAuth      func(c *gin.Context)
		setupMock      func(mt *mockTweetService)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:        "ツイート更新成功",
			tweetID:     "100",
			requestBody: dto.UpdateTweetRequest{Content: "更新後の内容"},
			setupAuth: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext{UserID: 10})
			},
			setupMock: func(mt *mockTweetService) {
				mt.On("EditTweet", mock.Anything, "更新後の内容", int64(100), int64(10)).
					Return(&models.Tweet{ID: 100, Content: "更新後の内容", IsEdited: true}, true, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				data := resp.Data.(map[string]any)
				assert.True(t, data["is_edited"].(bool))
			},
		},
		{
			name:        "更新成功：内容に変更なし",
			tweetID:     "100",
			requestBody: dto.UpdateTweetRequest{Content: "同じ内容"},
			setupAuth: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext{UserID: 10})
			},
			setupMock: func(mt *mockTweetService) {
				mt.On("EditTweet", mock.Anything, "同じ内容", int64(100), int64(10)).
					Return(&models.Tweet{ID: 100, Content: "同じ内容"}, false, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Contains(t, resp.Message, "変更はありません")
			},
		},
		{
			name:        "エラー：編集期限切れ",
			tweetID:     "100",
			requestBody: dto.UpdateTweetRequest{Content: "手遅れな更新"},
			setupAuth: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext{UserID: 10})
			},
			setupMock: func(mt *mockTweetService) {
				mt.On("EditTweet", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, false, errcode.ErrEditTimeExpired)
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := new(mockTweetService)
			h := NewTweetHandler(mt)
			tt.setupMock(mt)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", &buf)
			c.Params = []gin.Param{{Key: "id", Value: tt.tweetID}}

			tt.setupAuth(c)
			h.Update(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestTweetDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		tweetID        string
		setupAuth      func(c *gin.Context)
		setupMock      func(mt *mockTweetService)
		expectedStatus int
	}{
		{
			name:    "削除成功",
			tweetID: "100",
			setupAuth: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext{UserID: 10})
			},
			setupMock: func(mt *mockTweetService) {
				mt.On("RemoveTweet", mock.Anything, int64(100), int64(10)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "エラー：他人のツイートを削除しようとした",
			tweetID: "100",
			setupAuth: func(c *gin.Context) {
				c.Set(contextkeys.AuthPayloadKey, &dto.AuthContext{UserID: 99})
			},
			setupMock: func(mt *mockTweetService) {
				mt.On("RemoveTweet", mock.Anything, mock.Anything, mock.Anything).
					Return(errcode.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := new(mockTweetService)
			h := NewTweetHandler(mt)
			tt.setupMock(mt)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
			c.Params = []gin.Param{{Key: "id", Value: tt.tweetID}}

			tt.setupAuth(c)
			h.Delete(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
