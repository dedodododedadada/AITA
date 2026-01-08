package db
import(
	"log"
	"os"
	"testing"
	"github.com/jmoiron/sqlx"
)

var testStore UserStore

const testDbConnStr = "postgresql://aita_admin:password@localhost:5439/aita_db?sslmode=disable"

func TestMain(m *testing.M) {
	testDB, err := sqlx.Connect("postgres", testDbConnStr)
	if err != nil {
		log.Fatalf("テストデータベースに接続できません (%.50s...): %v", testDbConnStr, err)
	}

	testStore = NewPostgresUserStore(testDB)
	exitCode := m.Run()
	os.Exit(exitCode)
}

func cleanUpTestDB() {
	db := testStore.(*PostgresUserStore).db
	_, err := db.Exec(`TRUNCATE TABLE sessions, tweets, users RESTART IDENTITY CASCADE;`)
	if err != nil {
		log.Fatalf("テストデータベースに接続できません: %v", err)
	}
}
