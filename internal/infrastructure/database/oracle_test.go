package database

import (
	"strings"
	"testing"
)

func TestPostgresDSNFromJDBCURL(t *testing.T) {
	dsn, err := BuildPostgresDSN(
		"jdbc:postgresql://localhost:5432/weblelang",
		"test-user",
		"test-password",
	)
	if err != nil {
		t.Fatalf("build PostgreSQL DSN: %v", err)
	}

	for _, expected := range []string{"localhost", "5432", "weblelang", "test-user", "sslmode=disable"} {
		if !strings.Contains(dsn, expected) {
			t.Fatalf("expected DSN to contain %q", expected)
		}
	}
}

func TestPostgresDSNRejectsInvalidURL(t *testing.T) {
	if _, err := BuildPostgresDSN("localhost", "user", "password"); err == nil {
		t.Fatal("expected invalid URL error")
	}
}
