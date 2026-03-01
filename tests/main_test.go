package tests

import (
	"aita/internal/cache"
	"aita/internal/db"
	"aita/internal/pkg/crypto"
	"aita/internal/pkg/testutils"
	"aita/internal/repository"
	"aita/internal/service"
	"log"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

var (
	testUserStore    repository.Userstore  
	testUserCache    repository.UserCache    
	testSessionStore repository.SessionStore
	testTweetStore   service.TweetStore
	testTokemanager  service.TokenManager
	testHasher       service.PasswordHasher
	testContext      *testutils.TestContext
)

func TestMain(m *testing.M) {
	var teardown func()
	testContext, teardown = testutils.RunTestMain(m)
	log.Println("Database migration successful!")

	testHasher = crypto.NewBcryptHasher(bcrypt.DefaultCost)
	testTokemanager = crypto.NewTokenManager()
	testUserCache = cache.NewRedisUserCache(testContext.TestRDB)
	testUserStore = db.NewPostgresUserStore(testContext.TestDB)
	testSessionStore = db.NewRedisSessionStore(testContext.TestRDB)
	testTweetStore = db.NewPostgresTweetStore(testContext.TestDB)
	
	
    testContext.CleanupTestDB()

	exitCode := m.Run()

	teardown()
	os.Exit(exitCode)
}

