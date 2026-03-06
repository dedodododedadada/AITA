package db

import (
	"aita/internal/pkg/testutils"
	"log"
	"os"
	"testing"
)


var (
    testUserStore    *postgresUserStore
    testSessionStore *redisSessionStore
    testTweetStore   *postgresTweetStore
	testFollowStore  *postgresFollowStore
    testContext      *testutils.TestContext 
)

func TestMain(m *testing.M) {
	var teardown func()
	testContext, teardown = testutils.RunTestMain(m)
    log.Println("Migration successful!")
	testUserStore = NewPostgresUserStore(testContext.TestDB)
	testSessionStore = NewRedisSessionStore(testContext.TestRDB)
	testTweetStore = NewPostgresTweetStore(testContext.TestDB)
	testFollowStore = NewPostgresFollowStore(testContext.TestDB)
	testContext.CleanupTestDB()
	exitCode := m.Run()
	teardown()
	os.Exit(exitCode)
}



