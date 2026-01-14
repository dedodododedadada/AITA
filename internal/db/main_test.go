package db
import(
	"log"
	"os"
	"testing"
	"github.com/jmoiron/sqlx"
	_"github.com/lib/pq"
	"github.com/joho/godotenv"
	"github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	testUserStore UserStore
	testSessionStore SessionStore
	testDB *sqlx.DB
)


func TestMain(m *testing.M) {
	_ = godotenv.Load("../../.env") 
	testDBConnStr := os.Getenv("DB_TEST_URL")
	var err error
	testDB, err = sqlx.Connect("postgres", testDBConnStr)
	if err != nil {
		log.Fatalf("テストデータベースに接続できません (%.50s...): %v", testDBConnStr, err)
	}
	mig, err := migrate.New(
        "file://../../migrations", 
        os.Getenv("DB_TEST_URL"),
    )
    if err != nil {
        log.Fatalf("Could not create migrate instance: %v", err)
    }

    if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatalf("Could not run migrate up: %v", err)
    }
    log.Println("Migration successful!")
	testUserStore = NewPostgresUserStore(testDB)
	testSessionStore = NewPostgresSessionStore(testDB)
	cleanUpTestDB()
	exitCode := m.Run()
	testDB.Close()
	os.Exit(exitCode)
}

func cleanUpTestDB() {
	_, err := testDB.Exec(`TRUNCATE TABLE sessions, tweets, users RESTART IDENTITY CASCADE;`)
	if err != nil {
		log.Fatalf("テストデータベースに接続できません: %v", err)
	}
}
