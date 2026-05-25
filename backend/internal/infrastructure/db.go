package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"noant/config"

	_ "github.com/go-sql-driver/mysql"
)

func NewTiDBConnection(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=skip-verify&charset=utf8mb4&parseTime=true&loc=Local",
		cfg.TiDBUser,
		cfg.TiDBPassword,
		cfg.TiDBHost,
		cfg.TiDBPort,
		cfg.TiDBDatabase,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBPoolSize)
	db.SetMaxIdleConns(cfg.DBPoolSize / 2)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Test connection with retry
	var pingErr error
	for i := 0; i < 3; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if pingErr != nil {
		return nil, fmt.Errorf("failed to ping database after 3 attempts: %w", pingErr)
	}

	return db, nil
}


// PingContext wraps sql.DB.PingContext for health checks
func PingDB(db *sql.DB, ctx context.Context) error {
	return db.PingContext(ctx)
}