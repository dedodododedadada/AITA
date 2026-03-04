package main

import (
	"aita/internal/api"
	"aita/internal/cache"
	"aita/internal/configuration"
	"aita/internal/db"
	"aita/internal/pkg/crypto"
	"aita/internal/repository"
	"aita/internal/service"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/panjf2000/ants/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)


func main() {
	config := configuration.LoadConfig()
	database, err:= sqlx.Connect("postgres", config.DBConnStr)
	if err!= nil {
		log.Fatal("データベースに接続できません",err)
	}
	defer database.Close()
	log.Printf("✅ データベースへの接続に成功しました！")

	rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
        Password: config.RedisPassword,
		DB: 0, 
    })

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := rdb.Ping(ctx).Err(); err != nil {
        log.Fatalf("Redisに接続できません: %v", err)
    }
    log.Println("✅ Redisの接続に成功しました！")
	
	backfillPool, err := ants.NewPool(config.BackfillPoolSize)
	if err != nil {
		slog.Error("ルーチンプールの起動に失敗しました", "err", err)
		os.Exit(1)
	}
	defer backfillPool.Release()

	hasher := crypto.NewBcryptHasher(bcrypt.DefaultCost)
	tokenmanager := crypto.NewTokenManager()
	userStore := db.NewPostgresUserStore(database)
	sessionStore := db.NewRedisSessionStore(rdb)
	tweetStore := db.NewPostgresTweetStore(database)
	followStore := db.NewPostgresFollowStore(database)
	userCache := cache.NewRedisUserCache(rdb)
	followCache := cache.NewRedisFollowCache(rdb)
	userRepository := repository.NewUserRepository(userStore, userCache, backfillPool)
	serviceRepository := repository.NewSessionRepository(sessionStore)
	followRepository := repository.NewFollowRepository(followStore, followCache, backfillPool)
	userService := service.NewUserService(userRepository, hasher)
	sessionService := service.NewSessionService(serviceRepository, userService, tokenmanager)
	tweetService := service.NewTweetService(tweetStore)
	followService := service.NewFollowService(followRepository, userService)

	userHandler := api.NewUserHandler(userService, sessionService)
	tweetHandler := api.NewTweetHandler(tweetService)
	followHandler := api.NewFollowHandler(followService)

	router := api.SetupRouter(userHandler, tweetHandler, followHandler, sessionService)


	log.Printf("サーバーが起動し、ポート%sで待機中です",config.ServerAddress)
	if err :=router.Run(config.ServerAddress); err != nil{
		log.Fatal("サーバーの起動に失敗しました",err)
	}
}