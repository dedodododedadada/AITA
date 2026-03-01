package tests

import (
	"aita/internal/api"
	"aita/internal/dto"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserLifeCycleIntegration(t *testing.T) {
	testContext.CleanupTestDB()
	userRepository := repository.NewUserRepository(testUserStore, testUserCache)
	userService := service.NewUserService(userRepository, testHasher)
	sesseionRepository := repository.NewSessionRepository(testSessionStore)
	sessionService := service.NewSessionService(sesseionRepository, userService, testTokemanager)
	tweetService := service.NewTweetService(testTweetStore)
	userHandler := api.NewUserHandler(userService, sessionService)
	tweetHandler := api.NewTweetHandler(tweetService)

	gin.SetMode(gin.TestMode)
	r := api.SetupRouter(userHandler, tweetHandler, sessionService)
	signupPayload := dto.SignupRequest{
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

	loginPayload := dto.LoginRequest{
		Email:    "dev@aita.com",
		Password: "password123",
	}
	jsonLogin, _ := json.Marshal(loginPayload)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBuffer(jsonLogin))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var loginResp struct {
		Data struct {
			Token string `json:"session_token"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &loginResp)
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
