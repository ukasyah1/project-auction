package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func OpenPostgres(databaseURL, databasePort, databaseName, username, password string) (*gorm.DB, error) {
	if strings.TrimSpace(databaseURL) == "" || strings.TrimSpace(username) == "" || password == "" {
		return nil, fmt.Errorf("DATABASE_URL, DATABASE_USERNAME, dan DATABASE_PASSWORD wajib diisi")
	}

	dsn, err := BuildPostgresDSNFromConfig(databaseURL, databasePort, databaseName, username, password)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
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

// BuildPostgresDSN retains support for callers that provide a complete URL.
func BuildPostgresDSN(databaseURL, username, password string) (string, error) {
	return BuildPostgresDSNFromConfig(databaseURL, "", "", username, password)
}

// BuildPostgresDSNFromConfig accepts either:
//   - jdbc:postgresql://host:port/database
//   - postgresql://host:port/database
//   - DATABASE_URL=host with DATABASE_PORT and DATABASE_NAME
func BuildPostgresDSNFromConfig(databaseURL, databasePort, databaseName, username, password string) (string, error) {
	address := strings.TrimSpace(databaseURL)
	if address == "" {
		return "", fmt.Errorf("DATABASE_URL wajib diisi")
	}

	if !strings.Contains(address, "://") {
		return buildPostgresDSN(
			address,
			strings.TrimSpace(databasePort),
			strings.Trim(strings.TrimSpace(databaseName), "/"),
			username,
			password,
			nil,
		)
	}

	address = strings.TrimPrefix(address, "jdbc:")
	if strings.HasPrefix(address, "postgres://") {
		address = "postgresql://" + strings.TrimPrefix(address, "postgres://")
	}

	parsed, err := url.Parse(address)
	if err != nil || parsed.Scheme != "postgresql" {
		return "", fmt.Errorf("DATABASE_URL tidak valid")
	}
	if parsed.User != nil {
		return "", fmt.Errorf("username dan password harus diisi melalui DATABASE_USERNAME dan DATABASE_PASSWORD")
	}

	return buildPostgresDSN(
		parsed.Hostname(),
		parsed.Port(),
		strings.Trim(parsed.Path, "/"),
		username,
		password,
		parsed.Query(),
	)
}

func buildPostgresDSN(host, port, databaseName, username, password string, query url.Values) (string, error) {
	if host == "" || port == "" || databaseName == "" {
		return "", fmt.Errorf("DATABASE_URL wajib memuat host; DATABASE_PORT dan DATABASE_NAME juga wajib untuk format host-only")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return "line 106", fmt.Errorf("DATABASE_PORT harus berupa angka antara 1 dan 65535")
	}
	if strings.TrimSpace(username) == "" || password == "" {
		return "line 109", fmt.Errorf("DATABASE_USERNAME dan DATABASE_PASSWORD wajib diisi")
	}

	if query == nil {
		query = make(url.Values)
	}
	if query.Get("sslmode") == "" {
		query.Set("sslmode", "disable")
	}

	dsn := &url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(username, password),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + databaseName,
		RawQuery: query.Encode(),
	}
	return dsn.String(), nil
}
