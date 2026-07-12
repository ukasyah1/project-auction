package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func OpenPostgres(jdbcURL, username, password string) (*gorm.DB, error) {
	if strings.TrimSpace(jdbcURL) == "" || username == "" || password == "" {
		return nil, fmt.Errorf("DATABASE_URL, DATABASE_USERNAME, dan DATABASE_PASSWORD wajib diisi")
	}
	dsn, err := BuildPostgresDSN(jdbcURL, username, password)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get PostgreSQL connection pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping PostgreSQL database: %w", err)
	}
	return db, nil
}

func BuildPostgresDSN(jdbcURL, username, password string) (string, error) {
	address := strings.TrimPrefix(strings.TrimSpace(jdbcURL), "jdbc:")
	parsed, err := url.Parse(address)
	if err != nil || parsed.Scheme != "postgresql" {
		return "", fmt.Errorf("DATABASE_URL harus berbentuk jdbc:postgresql://host:port/database")
	}
	host, port, databaseName := parsed.Hostname(), parsed.Port(), strings.Trim(parsed.Path, "/")
	if host == "" || port == "" || databaseName == "" {
		return "", fmt.Errorf("DATABASE_URL harus berbentuk jdbc:postgresql://host:port/database")
	}
	if _, _, err := net.SplitHostPort(parsed.Host); err != nil {
		return "", fmt.Errorf("invalid PostgreSQL host and port: %w", err)
	}
	query := parsed.Query()
	if query.Get("sslmode") == "" {
		query.Set("sslmode", "disable")
	}
	return (&url.URL{Scheme: "postgresql", User: url.UserPassword(username, password), Host: parsed.Host, Path: "/" + databaseName, RawQuery: query.Encode()}).String(), nil
}
