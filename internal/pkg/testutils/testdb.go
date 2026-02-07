package testutils

import (
	"aita/internal/config"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type TestContext struct {
	TestDB *sqlx.DB
	DSN string
}

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
	_, err := ctx.TestDB.Exec(`TRUNCATE TABLE sessions, tweets, users RESTART IDENTITY CASCADE;`)
	if err != nil {
		log.Fatalf("テストデータベースに接続できません: %v", err)
	}
}
func RunTestMain(m *testing.M) (*TestContext, func()) {
	os.Setenv("APP_ENV", "test")
	cfg := config.LoadConfig()
	db, err := sqlx.Connect("postgres", cfg.DBConnStr)
	if err != nil {
		log.Fatalf("テストデータベースに接続できません (%.50s...): %v", cfg.DBConnStr, err)
	}
	migrationDir := config.GetPath("migrations")
	migrationURL := "file://" + filepath.ToSlash(migrationDir)
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
	teardown := func() {
		srcErr, dbErr := mig.Close()
		if srcErr != nil || dbErr != nil {
			log.Printf("マイグレーションインスタンスの停止に失敗しました: %v, %v", srcErr, dbErr)
		}
		db.Close()
	}
	return &TestContext{
		TestDB: db,
		DSN:    cfg.DBConnStr,
	}, teardown
}

