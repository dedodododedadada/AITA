package tests

import (
	"aita/internal/db"
	"aita/internal/pkg/testutils"
	"log"
	"os"
	"testing"
)

var (
	testUserStore		db.UserStore
	testSessionStore    db.SessionStore
	testTweetStore      db.TweetStore
	testContext        *testutils.TestContext
)

func TestMain(m *testing.M) {
	tc, teardown := testutils.RunTestMain(m)
	testContext = tc
	log.Println("Migration successful!")
	testUserStore = db.NewPostgresUserStore(testContext.TestDB)
	testSessionStore = db.NewPostgresSessionStore(testContext.TestDB)
	testTweetStore = db.NewPostgresTweetStore(testContext.TestDB)
	testContext.CleanupTestDB()
	exitCode := m.Run()
	teardown()
	os.Exit(exitCode)
}


