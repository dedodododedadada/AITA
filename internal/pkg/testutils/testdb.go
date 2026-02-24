package testutils

import (
	"aita/internal/configuration"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type TestContext struct {
	TestDB *sqlx.DB
	TestRDB *redis.Client
	DSN string
}

// データベースのオフライン状態を擬似的に再現する
func OpenDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, err
}


func  (ctx *TestContext) CleanupTestDB() {
	_, err := ctx.TestDB.Exec(`TRUNCATE TABLE follows, tweets, sessions, users RESTART IDENTITY CASCADE;`)
	if err != nil {
		log.Fatalf("テストデータベースに接続できません: %v", err)
	}
}
func RunTestMain(m *testing.M) (*TestContext, func()) {
	os.Setenv("APP_ENV", "test")

	cfg := configuration.LoadConfig()

	db, err := sqlx.Connect("postgres", cfg.DBConnStr)
	if err != nil {
		log.Fatalf("テストデータベースに接続できません (%.50s...): %v", cfg.DBConnStr, err)
	}

	migrationDir := configuration.GetPath("migrations")
	absPath, _ := filepath.Abs(migrationDir)
	cleanPath :=  filepath.ToSlash(absPath)

	var migrationURL string

	if runtime.GOOS == "windows" {
		migrationURL = "file://" + strings.TrimPrefix(cleanPath, "/")
	} else {
		migrationURL = "file://" + cleanPath
	}
	log.Printf("Migration URL: %s", migrationURL)

	mig, err := migrate.New(
        migrationURL, 
        cfg.DBConnStr,
    )
	if err != nil {
        log.Fatalf("マイグレーションインスタンスの生成に失敗しました: %v", err)
    }
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange  {
		log.Fatalf("マイグレーションの実行に失敗しました: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB: 1,
	})

	teardown := func() {
		srcErr, dbErr := mig.Close()
		if srcErr != nil || dbErr != nil {
			log.Printf("マイグレーションインスタンスの停止に失敗しました: %v, %v", srcErr, dbErr)
		}
		db.Close()
	}
	return &TestContext{
		TestDB: db,
		TestRDB: rdb,
		DSN:    cfg.DBConnStr,
	}, teardown
}

