package notification_test

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/notification"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/suite"
)

const (
	testDBURL = "postgresql://admin:admin@localhost:5432/ledger_test?sslmode=disable"
	dbName    = "ledger_test"
)

type RepositoryTestSuite struct {
	suite.Suite
	pool        *pgxpool.Pool
	store       db.Store
	repo        notification.Repository
	mockTracer  tracing.TracingService
	mockMetrics metrics.MetricsProvider
}

func (suite *RepositoryTestSuite) SetupSuite() {
	// Drop the database first to ensure a clean state
	exec.Command("dropdb", "--if-exists", dbName).Run()

	// Create the test database
	cmd := exec.Command("createdb", "--username=admin", "--owner=admin", dbName)
	if err := cmd.Run(); err != nil {
		log.Fatalf("could not create test database: %v", err)
	}

	// Get the absolute path to the migration directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working directory: %v", err)
	}
	migrationPath := filepath.Join(wd, "../../../db/migration")

	// Run migrations
	m, err := migrate.New("file://"+migrationPath, testDBURL)
	if err != nil {
		log.Fatalf("could not create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("could not run up migrations: %v", err)
	}

	// Connect to the database
	pool, err := pgxpool.New(context.Background(), testDBURL)
	if err != nil {
		log.Fatalf("could not connect to test database: %v", err)
	}
	suite.pool = pool
	suite.store = db.NewStore(suite.pool)

	// Initialize repositories and mocks
	suite.mockTracer = tracing.NewMockTracingService(nil)
	suite.mockMetrics, _ = metrics.NewMetricsService(metrics.MetricsConfig{Enabled: false})
	suite.repo = notification.NewRepository(suite.store)
}

func (suite *RepositoryTestSuite) TearDownSuite() {
	suite.pool.Close()

	// Drop the test database
	cmd := exec.Command("dropdb", dbName)
	if err := cmd.Run(); err != nil {
		log.Fatalf("could not drop test database: %v", err)
	}
}

func TestRepositoryTestSuite(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping repository tests in CI environment")
	}
	suite.Run(t, new(RepositoryTestSuite))
}

func (suite *RepositoryTestSuite) createTestUser(ctx context.Context, tenantID uuid.UUID) *identity.User {
	req := &identity.CreateUserRequest{
		EntityID: tenantID,
		Username: "testuser" + uuid.New().String(),
		Email:    "test" + uuid.New().String() + "@example.com",
		Password: "password",
		UserType: "INTERNAL",
	}

	var user *identity.User
	err := suite.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore db.Store) error {
		identityRepo := identity.NewRepository(txStore, suite.mockTracer, suite.mockMetrics)
		var innerErr error
		user, innerErr = identityRepo.CreateUser(ctx, req, "hashedpassword")
		return innerErr
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(user)
	return user
}

func (suite *RepositoryTestSuite) TestGetAndUpdateUserNotificationPreferences() {
	ctx := context.Background()
	tenantID := uuid.New()

	// 1. Create a new user for testing, ensuring it's done within the tenant context
	user := suite.createTestUser(ctx, tenantID)

	// 2. Get preferences for the new user (should be default)
	var prefs *notification.NotificationPreferences
	err := suite.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore db.Store) error {
		repo := notification.NewRepository(txStore)
		var innerErr error
		prefs, innerErr = repo.GetUserNotificationPreferences(ctx, user.ID)
		return innerErr
	})
	suite.NoError(err)
	suite.NotNil(prefs)
	suite.Equal(user.ID, prefs.UserID)
	suite.True(prefs.EmailNotifications)
	suite.False(prefs.SlackNotifications)

	// 3. Update the preferences
	prefs.EmailNotifications = false
	prefs.SlackNotifications = true
	err = suite.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore db.Store) error {
		repo := notification.NewRepository(txStore)
		return repo.UpdateUserNotificationPreferences(ctx, user.ID, prefs)
	})
	suite.NoError(err)

	// 4. Get the preferences again and verify the update
	var updatedPrefs *notification.NotificationPreferences
	err = suite.store.WithTenant(ctx, tenantID, func(ctx context.Context, txStore db.Store) error {
		repo := notification.NewRepository(txStore)
		var innerErr error
		updatedPrefs, innerErr = repo.GetUserNotificationPreferences(ctx, user.ID)
		return innerErr
	})
	suite.NoError(err)
	suite.NotNil(updatedPrefs)
	suite.Equal(user.ID, updatedPrefs.UserID)
	suite.False(updatedPrefs.EmailNotifications)
	suite.True(updatedPrefs.SlackNotifications)
}
