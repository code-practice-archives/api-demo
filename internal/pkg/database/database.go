package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"strings"

	"github.com/code-practice-archives/api-demo/migrations"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 按配置打开 MySQL，启动时自动建库并迁移表结构。
func Open(cfg Config) (*gorm.DB, error) {
	dsn := cfg.DSN()
	if err := ensureMySQLDatabase(dsn); err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	fsys, err := fs.Sub(migrations.FS, "mysql")
	if err != nil {
		return fmt.Errorf("migrations fs: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectMySQL, sqlDB, fsys)
	if err != nil {
		return fmt.Errorf("goose provider: %w", err)
	}

	if _, err := provider.Up(context.Background()); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}

	log.Println("database tables migrated")
	return nil
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
