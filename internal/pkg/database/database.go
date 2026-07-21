package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-practice-archives/api-demo/migrations"
	"github.com/glebarez/sqlite"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 按配置打开数据库，启动时自动建库（MySQL）并迁移表结构。
func Open(cfg Config) (*gorm.DB, error) {
	if cfg.Driver == DriverMySQL {
		if err := ensureMySQLDatabase(cfg.DSN); err != nil {
			return nil, err
		}
	}

	dialector, err := newDialector(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// SQLite 内存库每个连接是独立库，限制为单连接避免迁移丢失。
	if cfg.Driver == DriverSQLite && isMemoryDSN(cfg.DSN) {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("get sql db: %w", err)
		}
		sqlDB.SetMaxOpenConns(1)
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

// OpenSQLite 打开 SQLite，供测试使用。dsn 传 ":memory:" 即可。
func OpenSQLite(dsn string) (*gorm.DB, error) {
	return Open(Config{Driver: DriverSQLite, DSN: dsn})
}

func migrate(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	dialect, dir, err := gooseDialect(db.Dialector.Name())
	if err != nil {
		return err
	}

	fsys, err := fs.Sub(migrations.FS, dir)
	if err != nil {
		return fmt.Errorf("migrations fs %q: %w", dir, err)
	}

	provider, err := goose.NewProvider(dialect, sqlDB, fsys)
	if err != nil {
		return fmt.Errorf("goose provider: %w", err)
	}

	if _, err := provider.Up(context.Background()); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}

	log.Println("database tables migrated")
	return nil
}

func gooseDialect(name string) (goose.Dialect, string, error) {
	switch name {
	case "mysql":
		return goose.DialectMySQL, "mysql", nil
	case "sqlite":
		return goose.DialectSQLite3, "sqlite", nil
	default:
		return "", "", fmt.Errorf("unsupported goose dialect %q", name)
	}
}

// ensureMySQLDatabase 在连接业务库前创建数据库（若不存在）。
func ensureMySQLDatabase(dsn string) error {
	cfg, err := mysqldriver.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("parse mysql dsn: %w", err)
	}
	if cfg.DBName == "" {
		return nil
	}

	dbName := cfg.DBName
	cfg.DBName = ""

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return fmt.Errorf("open mysql server: %w", err)
	}
	defer db.Close()

	stmt := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		strings.ReplaceAll(dbName, "`", "``"),
	)
	if _, err := db.Exec(stmt); err != nil {
		return fmt.Errorf("create database %s: %w", dbName, err)
	}

	log.Printf("database %q ensured", dbName)
	return nil
}

func newDialector(driver, dsn string) (gorm.Dialector, error) {
	switch driver {
	case DriverMySQL:
		if dsn == "" {
			return nil, fmt.Errorf("mysql dsn is required")
		}
		return mysql.Open(dsn), nil
	case DriverSQLite:
		if dsn == "" {
			return nil, fmt.Errorf("sqlite dsn is required")
		}
		if err := prepareSQLiteDSN(dsn); err != nil {
			return nil, err
		}
		return sqlite.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported db driver %q (want mysql or sqlite)", driver)
	}
}

func prepareSQLiteDSN(dsn string) error {
	if isMemoryDSN(dsn) {
		return nil
	}
	dir := filepath.Dir(dsn)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create db dir: %w", err)
	}
	return nil
}

func isMemoryDSN(dsn string) bool {
	return dsn == ":memory:" || strings.Contains(dsn, "mode=memory")
}
