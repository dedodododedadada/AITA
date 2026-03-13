package main

import (
	"aita/internal/api"
	"aita/internal/cache"
	"aita/internal/configuration"
	"aita/internal/db"
	"aita/internal/pkg/crypto"
	"aita/internal/pkg/messagequeue"
	"aita/internal/producer"
	"aita/internal/repository"
	"aita/internal/service"
	"aita/internal/worker"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/panjf2000/ants/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)


func main() {
	config := configuration.LoadConfig()
	if config.AppEnv == "production" {
        gin.SetMode(gin.ReleaseMode)
    } else {
        gin.SetMode(gin.DebugMode)
    }
	
	database, err:= sqlx.Connect("postgres", config.DBConnStr)
	if err!= nil {
		log.Fatal("データベースに接続できません",err)
	}
	database.SetMaxOpenConns(config.DBMaxOpenConns)
    database.SetMaxIdleConns(config.DBMaxIdleConns)
    database.SetConnMaxLifetime(time.Duration(config.DBConnMaxLifetime) * time.Minute)
	defer database.Close()
	log.Printf("✅ データベースへの接続に成功しました！")

	rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
        Password: config.RedisPassword,
		DB: 0, 
		PoolSize: config.RedisPoolSize,
		MinIdleConns: config.RedisMinIdleConns,
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
	workerPool, err := ants.NewPool(config.WorkerPoolSize)
 		if err != nil {
		slog.Error("ルーチンプールの起動に失敗しました", "err", err)
		os.Exit(1)
	}
	defer workerPool.Release()

	tweetMQ :=  messagequeue.NewRedisMQ(rdb, config.TweetStream, config.FanoutGroup, "api-server-1")
	if err := tweetMQ.InitMQ(context.Background()); err != nil {
        log.Fatalf("MQ の初期化に失敗しました: %v", err)
    }
    log.Println("✅ Redis Stream (MQ) の初期化に成功しました！")

	hasher := crypto.NewBcryptHasher(bcrypt.DefaultCost)

	tokenmanager := crypto.NewTokenManager()

	fanoutProducer := producer.NewFanoutProducer(tweetMQ, workerPool)

	userStore := db.NewPostgresUserStore(database)
	sessionStore := db.NewRedisSessionStore(rdb)
	tweetStore := db.NewPostgresTweetStore(database)
	followStore := db.NewPostgresFollowStore(database)

	userCache := cache.NewRedisUserCache(rdb)
	followCache := cache.NewRedisFollowCache(rdb)
	tweetCache := cache.NewRedisTweetCache(rdb)
	timelineCache := cache.NewRedisTimelineCache(rdb)

	userRepository := repository.NewUserRepository(userStore, userCache, backfillPool)
	serviceRepository := repository.NewSessionRepository(sessionStore)
	followRepository := repository.NewFollowRepository(followStore, followCache, backfillPool)
	tweetRepository := repository.NewTweetRepository(tweetStore, tweetCache, backfillPool)
	timeLineRepository := repository.NewTimeLineRepository(timelineCache, backfillPool)

	userService := service.NewUserService(userRepository, hasher)
	sessionService := service.NewSessionService(serviceRepository, userService, tokenmanager)
	tweetService := service.NewTweetService(tweetRepository, fanoutProducer)
	followService := service.NewFollowService(followRepository, userService)
	timeLineService := service.NewTimeLineService(timeLineRepository, tweetService, backfillPool)
	fanoutWorker := worker.NewFanoutWorker(tweetMQ, followService, timeLineService, workerPool)

	userHandler := api.NewUserHandler(userService, sessionService)
	tweetHandler := api.NewTweetHandler(tweetService)
	followHandler := api.NewFollowHandler(followService)

	router := api.SetupRouter(userHandler, tweetHandler, followHandler, sessionService)

	workerCtx, workerCancel := context.WithCancel(context.Background())
    defer workerCancel()

	go func() {
		slog.Info("Main: FanoutWorker をバックグラウンドで開始します")
        fanoutWorker.Start(workerCtx)
	}()

	srv := &http.Server{
		Addr:    config.ServerAddress,
		Handler: router,
	}

	go func() {
		log.Printf("🚀 サーバーが启动し、ポート %s で待機中です", config.ServerAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Main: サーバーの起動に失敗しました: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	slog.Warn("Main: 終了シグナルを受信しました。シャットダウンを開始します...")


	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Main: API サーバーの強制終了", "error", err)
	}

	workerCancel()

	slog.Info("Main: 進行中のバックグラウンドタスクが完了するのを待機しています (5秒)...")
	time.Sleep(5 * time.Second)

	slog.Info("Main: リソースを解放しています...")
	
	WorkerPool.Release()
    backfillPool.Release()
	
	database.Close()
	rdb.Close()


	slog.Info("Main: すべてのプロセスが正常に終了しました。お疲れ様でした。")
}

