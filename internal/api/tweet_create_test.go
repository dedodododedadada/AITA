package api

import (
	"aita/internal/contextkeys"
	"aita/internal/models"
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
            requestBody: models.CreateTweetRequest{Content: "AITAの初投稿!"},
            setupAuth: func(c *gin.Context) {
                c.Set(contextkeys.AuthPayloadKey, int64(10))
            },
            setupMock: func(mt *mockTweetService) {
                expectedTweet := &models.Tweet{
                    ID: 100, 
                    Content: "AITAの初投稿!", 
                    UserID: 10,
                    CreatedAt: time.Now(),
                }
                mt.On("PostTweet", mock.Anything, int64(10), mock.Anything).Return(expectedTweet, nil)
            },
            expectedStatus: http.StatusCreated,
            checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
                fmt.Println("Raw Response:", w.Body.String())
                var resp models.Response
                json.Unmarshal(w.Body.Bytes(), &resp)
                data := resp.Data.(map[string]any)
                assert.Equal(t, float64(100), data["id"]) 
                assert.Equal(t, "AITAの初投稿!", data["content"])
            },
        },
        {
            name:        "JSON構文エラー",
            requestBody: `{"content": "incomplete...`,
            setupAuth: func(c *gin.Context) {
                c.Set(contextkeys.AuthPayloadKey, int64(10))
            },
            setupMock:      func(mt *mockTweetService) {},
            expectedStatus: http.StatusBadRequest,
            checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
                var resp models.Response
                json.Unmarshal(w.Body.Bytes(), &resp)
                assert.Equal(t, "INVALID_JSON_FORMAT", resp.Code)
            },
        },
        {
            name:        "バリデーションエラー：内容が空",
            requestBody: models.CreateTweetRequest{Content: ""},
            setupAuth: func(c *gin.Context) {
                c.Set(contextkeys.AuthPayloadKey, int64(10))
            },
            setupMock: func(mt *mockTweetService) {
                mt.On("PostTweet", mock.Anything, int64(10), mock.Anything).Return(nil, models.ErrRequiredFieldMissing)
            },
            expectedStatus: http.StatusBadRequest,
            checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
                var resp models.Response
                json.Unmarshal(w.Body.Bytes(), &resp)
                assert.Equal(t, "REQUIRED_FIELD_MISSING", resp.Code)
            },
        },
        {
            name:        "未認証エラー：ContextにIDがない",
            requestBody: models.CreateTweetRequest{Content: "Hello"},
            setupAuth:   func(c *gin.Context) {}, 
            setupMock:   func(mt *mockTweetService) {},
            expectedStatus: http.StatusUnauthorized,
            checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
                var resp models.Response
                json.Unmarshal(w.Body.Bytes(), &resp)
                assert.Equal(t, "SESSION_NOT_FOUND", resp.Code)
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