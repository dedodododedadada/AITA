package testutils

import (
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"                                        
    _ "github.com/golang-migrate/migrate/v4/database/postgres"   
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

type TestContext struct {
	TestDB *sqlx.DB
}

func  (ctx *TestContext) CleanupTestDB() {
	_, err := ctx.TestDB.Exec(`TRUNCATE TABLE sessions, tweets, users RESTART IDENTITY CASCADE;`)
	if err != nil {
		log.Fatalf("テストデータベースに接続できません: %v", err)
	}
}
func RunTestMain(m *testing.M, envPath string) (*TestContext, func()) {
	os.Setenv("APP_ENV", "test")
	_ = godotenv.Load(envPath)
	testDBConnStr := os.Getenv("DB_TEST_URL")
	db, err := sqlx.Connect("postgres", testDBConnStr)
	if(err != nil) {
		log.Fatalf("テストデータベースに接続できません (%.50s...): %v", testDBConnStr, err)
	}
	mig, err := migrate.New(
        "file://../../migrations", 
        testDBConnStr,
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
	return &TestContext{TestDB: db}, teardown
}

