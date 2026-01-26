package config
import(
	"os"
	"log"
	
	"github.com/joho/godotenv"
)
type Config struct{
	DBConnStr        string
	ServerAddress    string
}

func LoadConfig() *Config{
	if err := godotenv.Load(); err != nil {
		log.Println(".envファイルが見つかりません。システム環境変数を使用します")
	}
	dbURL := os.Getenv("DB_URL")
	if os.Getenv("APP_ENV") == "test" {
        dbURL = os.Getenv("DB_TEST_URL")
    }
	if dbURL == "" {
		log.Fatal("エラー：環境変数 DB_URL が設定されていません")
	}
	return &Config{
		DBConnStr:       dbURL,
		ServerAddress:   ":8080",
	}
}

