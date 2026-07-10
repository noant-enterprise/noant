package repository

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"
    "time"
)

var (
    ErrDeadlock = errors.New("deadlock detected")
)

const maxDeadlockRetries = 3

func isDeadlock(err error) bool {
    if err == nil {
        return false
    }
    // MySQL/TiDB deadlock error code 1213
    return strings.Contains(err.Error(), "deadlock") ||
        strings.Contains(err.Error(), "Error 1213") ||
        strings.Contains(err.Error(), "Deadlock found")
}

func retryOnDeadlock(fn func() error) error {
    var lastErr error
    for i := 0; i < maxDeadlockRetries; i++ {
        lastErr = fn()
        if lastErr == nil {
            return nil
        }
        if !isDeadlock(lastErr) {
            return lastErr
        }
        backoff := time.Duration(100*(i+1)) * time.Millisecond
        time.Sleep(backoff)
    }
    return fmt.Errorf("deadlock retry exhausted: %w", lastErr)
}

type UnitOfWork struct {
    db *sql.DB
    tx *sql.Tx
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
    return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Begin(ctx context.Context) error {
    tx, err := u.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    u.tx = tx
    return nil
}

func (u *UnitOfWork) Commit() error {
    if u.tx == nil {
        return fmt.Errorf("no active transaction")
    }
    return u.tx.Commit()
}

func (u *UnitOfWork) Rollback() {
    if u.tx != nil {
        _ = u.tx.Rollback()
    }
}

func (u *UnitOfWork) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    if u.tx != nil {
        return u.tx.ExecContext(ctx, query, args...)
    }
    return u.db.ExecContext(ctx, query, args...)
}

func (u *UnitOfWork) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
    if u.tx != nil {
        return u.tx.QueryRowContext(ctx, query, args...)
    }
    return u.db.QueryRowContext(ctx, query, args...)
}

func (u *UnitOfWork) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    if u.tx != nil {
        return u.tx.QueryContext(ctx, query, args...)
    }
    return u.db.QueryContext(ctx, query, args...)
}
