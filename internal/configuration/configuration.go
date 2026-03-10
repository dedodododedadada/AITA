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
	TweetStream      	string 
    FanoutGroup      	string 
    ConsumerName    	string 

	DBMaxOpenConns    	int 
    DBMaxIdleConns    	int 
    DBConnMaxLifetime 	int 
    RedisPoolSize     	int 
	RedisMinIdleConns   int

    WorkerPoolSize   	int 
    BackfillPoolSize 	int 

    //BackfillDBLimit 	int 
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




	cfg := &Config{
		DBConnStr:     		dbURL,
		ServerAddress: 		":8080",
		AppEnv:        		appEnv,
		RedisHost:     		os.Getenv("REDIS_HOST"),
		RedisPort:     		os.Getenv("REDIS_PORT"),
		RedisPassword: 		os.Getenv("REDIS_PASSWORD"),
		TweetStream:      	os.Getenv("TWEET_STREAM"),
        FanoutGroup:      	os.Getenv("FANOUT_GROUP"),
        ConsumerName:     	os.Getenv("CONSUMER_NAME"),
		DBMaxOpenConns:    	getEnvInt("DB_MAX_OPEN", 300),
        DBMaxIdleConns:    	getEnvInt("DB_MAX_IDLE", 50),
        DBConnMaxLifetime: 	getEnvInt("DB_MAX_LIFETIME", 30),
		RedisPoolSize:    	getEnvInt("REDIS_POOL_SIZE", 500),
		RedisMinIdleConns:  getEnvInt("REDIS_MIN_IDLE",20),
		BackfillPoolSize: 	getEnvInt("BACKFILL_POOL_SIZE", 500),
		WorkerPoolSize: 	getEnvInt("WORKER_POOL_SIZE", 2000),
		//BackfillDBLimit:	getEnvInt("BACKFILL_DB_LIMIT",70),	
	}

	if cfg.TweetStream == "" { 
		cfg.TweetStream = "aita:tweet:stream" 
	}
    if cfg.FanoutGroup == "" { 
		cfg.FanoutGroup = "aita:fanout:group" 
	}
    if cfg.ConsumerName == "" {
		hostname, _ := os.Hostname()
        cfg.ConsumerName = "api-node-" + hostname
    }



	requiredFields := map[string]string{
        "DB_URL":         cfg.DBConnStr,
        "REDIS_HOST":     cfg.RedisHost,
        "REDIS_PORT":     cfg.RedisPort,
        "REDIS_PASSWORD": cfg.RedisPassword,
        "TWEET_STREAM":   cfg.TweetStream, 
    }

	for key, val := range requiredFields {
        if val == "" {
            slog.Error("エラー: パラメーターが見つかりません", "key", key)
            os.Exit(1) 
        }
    }

	return cfg
}

func getEnvInt(key string, defaultVal int) int {
    if s := os.Getenv(key); s != "" {
        if v, err := strconv.Atoi(s); err == nil {
            return v
        }
    }
    return defaultVal
}