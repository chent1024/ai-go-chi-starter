package postgres

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestConfigurePool(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	configurePool(db, Options{
		MaxOpenConns:    12,
		MaxIdleConns:    4,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 15 * time.Minute,
	})

	if stats := db.Stats(); stats.MaxOpenConnections != 12 {
		t.Fatalf("MaxOpenConnections = %d", stats.MaxOpenConnections)
	}
}
