package tests

import (
	"aita/internal/api"
	"aita/internal/pkg/app"
	"aita/internal/repository"
	"aita/internal/service"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/panjf2000/ants/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserLifeCycleIntegration(t *testing.T) {
	testContext.CleanupTestDB()
	testPool, err := ants.NewPool(10)
    require.NoError(t, err, "テスト用コルーチンプールの初期化に失敗しました")
    
    defer testPool.Release()
	userRepository := repository.NewUserRepository(testUserStore, testUserCache, testPool)
	sesseionRepository := repository.NewSessionRepository(testSessionStore)
	followRepository := repository.NewFollowRepository(testFollowStore, testFollowCache, testPool)
	tweetRepository := repository.NewTweetRepository(testTweetStore, testTweetCache, testPool)
	userService := service.NewUserService(userRepository, testHasher)
	sessionService := service.NewSessionService(sesseionRepository, userService, testTokemanager)
	followService := service.NewFollowService(followRepository, userService)
	tweetService := service.NewTweetService(tweetRepository)
	userHandler := api.NewUserHandler(userService, sessionService)
	tweetHandler := api.NewTweetHandler(tweetService)
	followHandler := api.NewFollowHandler(followService)

	gin.SetMode(gin.TestMode)
	r := api.SetupRouter(userHandler, tweetHandler, followHandler,sessionService)
	signupPayload := app.SignupRequest{
		Username: "frontend_dev",
		Email:    "dev@aita.com",
		Password: "password123",
	}
	jsonSignup, _ := json.Marshal(signupPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/signup", bytes.NewBuffer(jsonSignup))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	log.Printf("%s", w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "frontend_dev")
	assert.Contains(t, w.Body.String(), "dev@aita.com")

	loginPayload := app.LoginRequest{
		Email:    "dev@aita.com",
		Password: "password123",
	}
	jsonLogin, _ := json.Marshal(loginPayload)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBuffer(jsonLogin))
	r.ServeHTTP(w, req)
	log.Printf("%s", w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)

	var loginResp struct {
		Data struct {
			Token string `json:"session_token"`
		} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	token := loginResp.Data.Token
	require.NotEmpty(t, token)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	fmt.Println("Body:", w.Body.String())
	assert.Contains(t, w.Body.String(), "frontend_dev")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/me", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
