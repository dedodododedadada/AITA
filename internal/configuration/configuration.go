package configuration

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

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
	DBConnStr        	string
	ServerAddress    	string
	AppEnv           	string
	RedisHost        	string
	RedisPort        	string
	RedisPassword    	string
	BackfillPoolSize 	int
}

func LoadConfig() *Config{
	envPath := GetPath(".env")
	if err := godotenv.Load(envPath); err != nil {
		slog.Error("エラー :.envが見つかりません")
		os.Exit(1)
	}

	appEnv := os.Getenv("APP_ENV")
	if 	appEnv == "" {
		appEnv = "development"
	}

	var dbURL string
	if appEnv== "test" {
        dbURL = os.Getenv("DB_TEST_URL")
		slog.Info("DB_TEST_URLに切り替えます")
    } else {
		dbURL = os.Getenv("DB_URL")
	}

	poolSizeStr := os.Getenv("BACKFILL_POOL_SIZE")
	poolSize, err := strconv.Atoi(poolSizeStr)

	if err != nil || poolSize <= 0 {
		poolSize = 100
	}

	cfg := &Config{
		DBConnStr:     dbURL,
		ServerAddress: ":8080",
		AppEnv:        appEnv,
		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		BackfillPoolSize: poolSize,
	}

	requiredFields := map[string]string{
        "DB_URL":         cfg.DBConnStr,
        "REDIS_HOST":     cfg.RedisHost,
        "REDIS_PORT":     cfg.RedisPort,
        "REDIS_PASSWORD": cfg.RedisPassword,
    }

	for key, val := range requiredFields {
        if val == "" {
            slog.Error("エラー: パラメーターが見つかりません", "key", key)
            os.Exit(1) 
        }
    }

	return cfg
}


