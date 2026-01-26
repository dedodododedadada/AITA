package tests

import (
	"aita/internal/api"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserLifeCycleIntegration(t *testing.T) {
	testContext.CleanupTestDB()
	userHandler := api.NewUserHandler(testUserStore, testSessionStore)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// to be continued
}	