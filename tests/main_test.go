package tests

import (
	"aita/internal/db"
	"aita/internal/pkg/crypto"
	"aita/internal/pkg/testutils"
	"aita/internal/service"
	"log"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

var (
	testUserStore    service.UserStore       
	testSessionStore service.SessionStore
	testTweetStore   service.TweetStore
	testTokemanager  service.TokenManager
	testHasher       service.PasswordHasher
	testContext      *testutils.TestContext
)

func TestMain(m *testing.M) {
	tc, teardown := testutils.RunTestMain(m)
	testContext = tc
	log.Println("Database migration successful!")

	testHasher = crypto.NewBcryptHasher(bcrypt.DefaultCost)
	testTokemanager = crypto.NewTokenManager()

	testUserStore = db.NewPostgresUserStore(testContext.TestDB)
	testSessionStore = db.NewPostgresSessionStore(testContext.TestDB)
	testTweetStore = db.NewPostgresTweetStore(testContext.TestDB)
	
	
    testContext.CleanupTestDB()

	exitCode := m.Run()

	teardown()
	os.Exit(exitCode)
}

