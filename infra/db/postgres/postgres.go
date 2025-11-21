package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"shared/utils"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	dbInstance *sqlx.DB
	once       sync.Once
	initErr    error
)

const migrationPath = "./db/migrations"

// ---------------------- Core DB Init ----------------------

func initializeDB(ctx context.Context, host, user, password, dbName, port, sslMode string) error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Kolkata",
		host, user, password, dbName, port, sslMode,
	)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctxPing); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if gin.Mode() == gin.DebugMode {
		if err := runMigrations(db); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	dbInstance = db
	return nil
}

func runMigrations(db *sqlx.DB) error {
	entries, err := os.ReadDir(migrationPath)
	if err != nil || len(entries) == 0 {
		return errors.New("migration directory is missing or empty, skipping migrations")
	}

	hasSQL := false
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			hasSQL = true
			break
		}
	}
	if !hasSQL {
		return errors.New("no .sql files found, skipping migrations")
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create driver instance: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// ---------------------- Tenant Table Helpers ----------------------

// TenantTable returns the full tenant-specific table name
func TenantTable(baseTableName, tenantID string) string {
	return fmt.Sprintf("%s_%s", baseTableName, tenantID)
}

// CreateTenantTable creates a table for a tenant if it does not exist
func CreateTenantTable(ctx context.Context, baseTableName, tenantID, tableSQL string) error {
	if dbInstance == nil {
		return errors.New("database not initialized")
	}

	tableName := TenantTable(baseTableName, tenantID)
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", utils.PqQuoteIdentifier(tableName), tableSQL)

	if _, err := dbInstance.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create tenant table '%s': %w", tableName, err)
	}

	return nil
}

// DropTenantTable drops a tenant table safely
func DropTenantTable(ctx context.Context, baseTableName, tenantID string) error {
	if dbInstance == nil {
		return errors.New("database not initialized")
	}

	tableName := TenantTable(baseTableName, tenantID)
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", utils.PqQuoteIdentifier(tableName))

	if _, err := dbInstance.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to drop tenant table '%s': %w", tableName, err)
	}

	return nil
}

// ---------------------- DB Lifecycle ----------------------

func InitDB(ctx context.Context, host, user, password, dbName, port, sslMode string) error {
	var err error
	once.Do(func() {
		err = initializeDB(ctx, host, user, password, dbName, port, sslMode)
		initErr = err
	})
	return err
}

func GetDB() (*sqlx.DB, error) {
	if dbInstance == nil {
		return nil, fmt.Errorf("database not initialized: %w", initErr)
	}
	return dbInstance, initErr
}

func CloseDB() error {
	if dbInstance != nil {
		if err := dbInstance.Close(); err != nil {
			return fmt.Errorf("failed to close DB: %w", err)
		}
		dbInstance = nil
	}
	return initErr
}
