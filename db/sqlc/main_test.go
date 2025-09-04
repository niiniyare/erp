package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var (
	testQueries *Queries
	testDB      *pgxpool.Pool
	testStore   Store
)

func TestMain(m *testing.M) {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		databaseUrl = "postgresql://admin:admin@localhost:5432/ledger?sslmode=disable"
		log.Println("DATABASE_URL not set, using default:", databaseUrl)
	}

	var err error
	testDB, err = pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)
	testStore = NewStore(testDB)

	exitCode := m.Run()

	testDB.Close()
	os.Exit(exitCode)
}

func createModuleForTest(t *testing.T, ctx context.Context) *Module {
	arg := CreateModuleParams{
		Name:        "test_module",
		DisplayName: stringPtr("Test Module"),
		Category:    stringPtr("CORE"),
	}
	module, err := testQueries.CreateModule(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, module)
	return module
}

func createUserForTest(t *testing.T, ctx context.Context) *User {
	arg := CreateUserParams{
		Username: "testuser",
		Email:    "testuser@example.com",
	}
	user, err := testQueries.CreateUser(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	return user
}

func createResourceForTest(t *testing.T, ctx context.Context, moduleID uuid.UUID) *Resource {
	arg := CreateResourceParams{
		ModuleID:     moduleID,
		Name:         "test_resource",
		ResourceType: "API",
	}
	resource, err := testQueries.CreateResource(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, resource)
	return resource
}

func createActionForTest(t *testing.T, ctx context.Context) *Action {
	arg := CreateActionParams{
		Name:       "test_action",
		ActionType: "READ",
	}
	action, err := testQueries.CreateAction(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, action)
	return action
}
