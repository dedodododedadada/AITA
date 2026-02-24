package configuration

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

var projectRoot string

// パスの自動検知
func init() {
	// 環境変数のチェック（本番環境のカスタマイズ用）
	projectRoot = os.Getenv("AITA_PROJECT_ROOT")
	if projectRoot != "" {
		return
	}

	//カレントディレクトリに .env があるか確認（バイナリと設定ファイルが同階層にある場合）
	if _, err := os.Stat(".env"); err == nil {
		cwd, _ := os.Getwd()
		projectRoot = cwd
		return
	}

	//開発環境用フォールバック（runtime を使用してソースコードのパスを取得）
	_, b, _, _ := runtime.Caller(0)
	projectRoot = filepath.Join(filepath.Dir(b), "..", "..")
}

func GetPath(relPath string) string {
	return filepath.Join(projectRoot, relPath)
}

type Config struct{
	DBConnStr        string
	ServerAddress    string
	AppEnv           string
	RedisHost        string
	RedisPort        string
	RedisPassword    string
}

func LoadConfig() *Config{
	envPath := GetPath(".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Fatal("エラー :.envが見つかりません")
	}

	appEnv := os.Getenv("APP_ENV")
	if 	appEnv == "" {
		appEnv = "development"
	}

	var dbURL string
	if appEnv== "test" {
        dbURL = os.Getenv("DB_TEST_URL")
		log.Println("DB_TEST_URLに切り替えます")
    } else {
		dbURL = os.Getenv("DB_URL")
	}

	if dbURL == "" {
		log.Fatal("エラー: 環境変数 DB_URL が設定されていません")
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		log.Fatal("エラー: redisHostが見つかりません")
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		log.Fatal("エラー: redisPortが見つかりません")
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		log.Fatal("エラー: redisPasswordが見つかりません")
	}

	return &Config{
		DBConnStr:     dbURL,
		ServerAddress: ":8080",
		AppEnv:        appEnv,
		RedisHost:     redisHost,
		RedisPort:     redisPort,
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
	}
}


