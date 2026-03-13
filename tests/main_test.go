package tests

import (
	"aita/internal/cache"
	"aita/internal/db"
	"aita/internal/pkg/crypto"
	"aita/internal/pkg/messagequeue"
	"aita/internal/repository"
	"aita/internal/service"
	"aita/internal/testconfig"
	"log"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

var (
	testUserStore    	repository.UserStore  
	testUserCache    	repository.UserCache 
	testFollowStore  	repository.FollowStore
	testFollowCache  	repository.FollowCache
	testSessionStore 	repository.SessionStore
	testTweetStore   	repository.TweetStore
	testTweetCache   	repository.TweetCache
	testTimeLineCache   repository.TimeLineCache
	testTokemanager  	service.TokenManager
	testHasher       	service.PasswordHasher
	testContext      	*testConfig.TestContext
	testMQ           	*messagequeue.RedisMQ
)

func TestMain(m *testing.M) {
	var teardown func()
	testContext, teardown = testConfig.RunTestMain(m)
	log.Println("Database migration successful!")

	testHasher = crypto.NewBcryptHasher(bcrypt.DefaultCost)
	testTokemanager = crypto.NewTokenManager() 
	testUserCache = cache.NewRedisUserCache(testContext.TestRDB)
	testFollowCache = cache.NewRedisFollowCache(testContext.TestRDB)
	testTweetCache = cache.NewRedisTweetCache(testContext.TestRDB)
	testTimeLineCache = cache.NewRedisTimelineCache(testContext.TestRDB)
	testUserStore = db.NewPostgresUserStore(testContext.TestDB)
	testSessionStore = db.NewRedisSessionStore(testContext.TestRDB)
	testTweetStore = db.NewPostgresTweetStore(testContext.TestDB)
	testFollowStore = db.NewPostgresFollowStore(testContext.TestDB)
	
	testStream := "test:aita:tweet:stream"
	testGroup  := "test:fanout:group"
	testConsumer := "test:consumer-1"
	testMQ = messagequeue.NewRedisMQ(testContext.TestRDB, testStream, testGroup, testConsumer)
	
    testContext.CleanupTestDB()

	exitCode := m.Run()

	teardown()
	os.Exit(exitCode)
}

