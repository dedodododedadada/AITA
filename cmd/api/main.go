package main
import(
	"aita/internal/api"
	"aita/internal/db"
	"os"
	"log"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)
// 	Configuration（+file
type Config struct{
	DBConnStr        string
	ServerAddress    string
}

func loadConfig() *Config{
	if err := godotenv.Load(); err != nil {
		log.Println(".envファイルが見つかりません。システム環境変数を使用します")
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("エラー：環境変数 DB_URL が設定されていません")
	}
	return &Config{
		DBConnStr:       dbURL,
		ServerAddress:   ":8080",
	}
}

func setupRouter(userHandler *api.UserHandler) *gin.Engine{
	router := gin.Default()
	v1 := router.Group("/api/v1")
	{
		v1.POST("/signup", userHandler.SignUp)
	}
	return router
}

func main() {
	config := loadConfig()

	database, err:= sqlx.Connect("postgres",config.DBConnStr)
	if err!= nil {
		log.Fatal("データベースに接続できません",err)
	}
	defer database.Close()
	log.Printf("データベースへの接続に成功しました")

	userStore := db.NewPostgresUserStore(database)
	userHandler := api.NewUserHandler(userStore)

	router := setupRouter(userHandler)
	log.Printf("サーバーが起動し、ポート%sで待機中です",config.ServerAddress)
	if err :=router.Run(config.ServerAddress); err != nil{
		log.Fatal("サーバーの起動に失敗しました",err)
	}
}