package main

import (
	"aita/internal/api"
	"aita/internal/config"
	"aita/internal/db"
	"aita/internal/pkg/crypto"
	"aita/internal/service"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)


func main() {
	config := config.LoadConfig()
	database, err:= sqlx.Connect("postgres",config.DBConnStr)
	if err!= nil {
		log.Fatal("データベースに接続できません",err)
	}
	defer database.Close()
	log.Printf("データベースへの接続に成功しました")
	
	hasher := crypto.NewBcryptHasher(bcrypt.DefaultCost)
	tokenmanager := crypto.NewTokenManager()
	userStore := db.NewPostgresUserStore(database)
	sessionStore := db.NewPostgresSessionStore(database)
	tweetStore := db.NewPostgresTweetStore(database)
	userService := service.NewUserService(userStore, hasher)
	sessionService := service.NewSessionService(sessionStore, userService, tokenmanager)
	tweetService := service.NewTweetService(tweetStore)
	userHandler := api.NewUserHandler(userService, sessionService)
	tweetHandler := api.NewTweetHandler(tweetService)

	router := api.SetupRouter(userHandler, tweetHandler, sessionService)

	//srv := &http.Server{}
	log.Printf("サーバーが起動し、ポート%sで待機中です",config.ServerAddress)
	if err :=router.Run(config.ServerAddress); err != nil{
		log.Fatal("サーバーの起動に失敗しました",err)
	}
}