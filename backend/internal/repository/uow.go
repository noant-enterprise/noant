package repository

import (
    "context"
    "database/sql"
    "fmt"
)

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
