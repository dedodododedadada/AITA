package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

var projectRoot string

func init() {
	_, b, _, _ := runtime.Caller(0)
	projectRoot = filepath.Join(filepath.Dir(b), "..", "..")
}

func GetPath(relPath string) string {
	return filepath.Join(projectRoot, relPath)
}

type Config struct{
	DBConnStr        string
	ServerAddress    string
}

func LoadConfig() *Config{
	envPath := GetPath(".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Fatal("エラー :.envが見つかりません")
	}
	dbURL := os.Getenv("DB_URL")
	if os.Getenv("APP_ENV") == "test" {
        dbURL = os.Getenv("DB_TEST_URL")
		log.Println("DB_TEST_URLに切り替えます")
    }

	if dbURL == "" {
		log.Fatal("エラー：環境変数 DB_URL が設定されていません")
	}
	return &Config{
		DBConnStr:       dbURL,
		ServerAddress:   ":8080",
	}
}

