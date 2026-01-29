package db

import (
	"aita/internal/pkg/testutils"
	"log"
	"os"
	"testing"
)


var (
	testUserStore		UserStore
	testSessionStore    SessionStore
	testTweetStore      TweetStore
	testContext         *testutils.TestContext 
)

func TestMain(m *testing.M) {
	tc, teardown := testutils.RunTestMain(m)
	testContext = tc
    log.Println("Migration successful!")
	testUserStore = NewPostgresUserStore(testContext.TestDB)
	testSessionStore = NewPostgresSessionStore(testContext.TestDB)
	testTweetStore = NewPostgresTweetStore(testContext.TestDB)
	testContext.CleanupTestDB()
	exitCode := m.Run()
	teardown()
	os.Exit(exitCode)
}


