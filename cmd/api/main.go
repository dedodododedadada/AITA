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
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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
	
	hasher := crypto.NewBcryptHasher(bcrypt.DefaultCost)
	tokenmanager := crypto.NewTokenManager()
	userStore := db.NewPostgresUserStore(database)
	sessionStore := db.NewRedisSessionStore(rdb)
	tweetStore := db.NewPostgresTweetStore(database)
	userCache := cache.NewRedisUserCache(rdb)
	userRepository := repository.NewUserRepository(userStore, userCache)
	serviceRepository := repository.NewSessionRepository(sessionStore)
	userService := service.NewUserService(userRepository, hasher)
	sessionService := service.NewSessionService(serviceRepository, userService, tokenmanager)
	tweetService := service.NewTweetService(tweetStore)
	userHandler := api.NewUserHandler(userService, sessionService)
	tweetHandler := api.NewTweetHandler(tweetService)

	router := api.SetupRouter(userHandler, tweetHandler, sessionService)


	log.Printf("サーバーが起動し、ポート%sで待機中です",config.ServerAddress)
	if err :=router.Run(config.ServerAddress); err != nil{
		log.Fatal("サーバーの起動に失敗しました",err)
	}
}