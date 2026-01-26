package main
import(
	"aita/internal/api"
	"aita/internal/db"
	"aita/internal/config"
	"log"

	"github.com/jmoiron/sqlx"
	_"github.com/lib/pq"
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
	userHandler := api.NewUserHandler(userStore, sessionStore)
	router := api.SetupRouter(userHandler, sessionStore)
	log.Printf("サーバーが起動し、ポート%sで待機中です",config.ServerAddress)
	if err :=router.Run(config.ServerAddress); err != nil{
		log.Fatal("サーバーの起動に失敗しました",err)
	}
}