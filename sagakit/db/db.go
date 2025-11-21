package db

import "context"

type Tx interface {
    Exec(ctx context.Context, query string, args ...any) error
    Query(ctx context.Context, query string, args ...any) (Rows, error)
    Commit() error
    Rollback() error
}

type Rows interface {
    Next() bool
    Scan(dest ...any) error
    Close() error
}

type DB interface {
    BeginTx(ctx context.Context) (Tx, error)
}
