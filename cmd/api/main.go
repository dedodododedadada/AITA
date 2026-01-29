package main

import (
	"aita/internal/api"
	"aita/internal/config"
	"aita/internal/db"
	"aita/internal/service"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)


func main() {
	config := config.LoadConfig()
	database, err:= sqlx.Connect("postgres",config.DBConnStr)
	if err!= nil {
		log.Fatal("データベースに接続できません",err)
	}
	defer database.Close()
	log.Printf("データベースへの接続に成功しました")

	userStore := db.NewPostgresUserStore(database)
	sessionStore := db.NewPostgresSessionStore(database)
	tweetStore := db.NewPostgresTweetStore(database)
	tweetService := service.NewTweetService(tweetStore)
	userHandler := api.NewUserHandler(userStore, sessionStore)
	tweetHandler := api.NewTweetHandler(tweetService)

	router := api.SetupRouter(userHandler, tweetHandler, sessionStore)

	//srv := &http.Server{}
	log.Printf("サーバーが起動し、ポート%sで待機中です",config.ServerAddress)
	if err :=router.Run(config.ServerAddress); err != nil{
		log.Fatal("サーバーの起動に失敗しました",err)
	}
}