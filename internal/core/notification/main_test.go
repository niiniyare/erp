package notification

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/niiniyare/erp/db/sqlc"
)

var testPool *pgxpool.Pool
var testStore db.Store

func TestMain(m *testing.M) {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	testPool, err = pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testStore = db.NewStore(testPool)

	os.Exit(m.Run())
}
