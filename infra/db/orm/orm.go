package orm

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	gormDB   *gorm.DB
	gormOnce sync.Once
	gormErr  error
)

// initializeGormDB sets up a singleton GORM connection.
func initializeGormDB(ctx context.Context, host, user, password, dbName, port, sslMode string) error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Kolkata",
		host, user, password, dbName, port, sslMode,
	)

	var gormLogger logger.Interface
	if gin.Mode() == gin.DebugMode {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		gormErr = fmt.Errorf("failed to open GORM database: %w", err)
		return gormErr
	}

	sqlDB, err := db.DB()
	if err != nil {
		gormErr = fmt.Errorf("failed to get sql.DB: %w", err)
		return gormErr
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctxPing); err != nil {
		gormErr = fmt.Errorf("failed to ping GORM database: %w", err)
		return gormErr
	}

	if gin.Mode() == gin.DebugMode {
		entries, err := os.ReadDir("./db/models")
		if err != nil || len(entries) == 0 {
			log.Println("⚠️ Models directory is missing or empty, skipping AutoMigrate")
		} else {
			log.Println("ℹ️ AutoMigrate is enabled. Call AutoMigrateAndDistribute() with your models.")
		}
	}

	gormDB = db
	return nil
}

// InitGormDB initializes the GORM connection (singleton).
func InitGormDB(ctx context.Context, host, user, password, dbName, port, sslMode string) error {
	var initErr error
	gormOnce.Do(func() {
		initErr = initializeGormDB(ctx, host, user, password, dbName, port, sslMode)
	})
	return initErr
}

// GetGormDB returns the active DB instance.
func GetGormDB() (*gorm.DB, error) {
	if gormDB == nil {
		return nil, fmt.Errorf("GORM database not initialized: %w", gormErr)
	}
	return gormDB, gormErr
}

// CloseGormDB gracefully closes the DB connection.
func CloseGormDB() error {
	if gormDB != nil {
		if sqlDB, err := gormDB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				return fmt.Errorf("failed to close GORM DB: %w", err)
			}
			log.Println("✅ GORM DB connection closed")
		} else {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
	}
	return gormErr
}

// AutoMigrateAndDistribute runs migrations, creates distributed tables, and ensures indexes.
func AutoMigrateAndDistribute(models ...interface{}) error {
	if gormDB == nil {
		return fmt.Errorf("GORM database not initialized")
	}

	if err := gormDB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("GORM AutoMigrate failed: %w", err)
	}
	log.Println("✅ AutoMigrate executed successfully")

	return nil
}

// ensureDistributed checks if table is distributed, creates if not.
func ensureDistributed(sqlDB *sql.DB, table, column string) error {
	var exists bool
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM pg_dist_partition
			WHERE logicalrelid = $1::regclass
		);`
	if err := sqlDB.QueryRow(checkQuery, table).Scan(&exists); err != nil {
		return fmt.Errorf("failed checking distribution: %w", err)
	}

	if !exists {
		query := fmt.Sprintf("SELECT create_distributed_table('%s', '%s');", table, column)
		if _, err := sqlDB.Exec(query); err != nil {
			return fmt.Errorf("failed distributing %s: %w", table, err)
		}
		log.Printf("🚀 Distributed table: %s ON %s\n", table, column)
	} else {
		log.Printf("ℹ️ Table %s already distributed\n", table)
	}
	return nil
}

// ensureIndexes parses GORM tags and creates missing indexes.
func ensureIndexes(sqlDB *sql.DB, model interface{}, table string) error {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("gorm")

		if strings.Contains(tag, "index") || strings.Contains(tag, "uniqueIndex") {
			unique := strings.Contains(tag, "uniqueIndex")
			colName := extractColumnName(field)

			idxName := fmt.Sprintf("idx_%s_%s", table, colName)

			if strings.Contains(tag, "index:") {
				idxName = parseCustomIndexName(tag, "index:")
			}
			if strings.Contains(tag, "uniqueIndex:") {
				idxName = parseCustomIndexName(tag, "uniqueIndex:")
			}

			var exists bool
			checkQuery := `
				SELECT EXISTS (
					SELECT 1 FROM pg_indexes
					WHERE tablename = $1 AND indexname = $2
				);`
			if err := sqlDB.QueryRow(checkQuery, table, idxName).Scan(&exists); err != nil {
				return fmt.Errorf("failed checking index %s: %w", idxName, err)
			}

			if !exists {
				indexType := "INDEX"
				if unique {
					indexType = "UNIQUE INDEX"
				}
				query := fmt.Sprintf("CREATE %s %s ON %s (%s);", indexType, idxName, table, colName)
				if _, err := sqlDB.Exec(query); err != nil {
					return fmt.Errorf("failed creating index %s: %w", idxName, err)
				}
				log.Printf("🆕 Created %s on %s(%s)\n", idxName, table, colName)
			} else {
				log.Printf("ℹ️ Index %s already exists on %s\n", idxName, table)
			}
		}
	}
	return nil
}

// getDistributionColumn finds a distkey tag or falls back to common fields.
func getDistributionColumn(model interface{}) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("gorm")
		if strings.Contains(tag, "distkey:") {
			return strings.TrimPrefix(tag, "distkey:")
		}
	}

	candidates := []string{"user_id", "org_id", "tenant_id", "id"}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		col := extractColumnName(field)
		if contains(candidates, col) {
			return col
		}
	}
	return ""
}

func extractColumnName(field reflect.StructField) string {
	tag := field.Tag.Get("gorm")
	if strings.Contains(tag, "column:") {
		parts := strings.Split(tag, "column:")
		if len(parts) > 1 {
			return strings.Fields(parts[1])[0]
		}
	}
	return strings.ToLower(field.Name)
}

func parseCustomIndexName(tag, prefix string) string {
	parts := strings.Split(tag, prefix)
	if len(parts) > 1 {
		return strings.Fields(parts[1])[0]
	}
	return ""
}

func contains(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}
