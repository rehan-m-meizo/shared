package pg

import (
	"context"
	"shared/sagakit/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) *DB { return &DB{Pool: pool} }

func (d *DB) BeginTx(ctx context.Context) (db.Tx, error) {
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

type Tx struct {
	tx pgx.Tx
}

func (t *Tx) Exec(ctx context.Context, query string, args ...any) error {
	_, err := t.tx.Exec(ctx, query, args...)
	return err
}

func (t *Tx) Query(ctx context.Context, query string, args ...any) (db.Rows, error) {
	rows, err := t.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

func (t *Tx) Commit() error   { return t.tx.Commit(context.Background()) }
func (t *Tx) Rollback() error { return t.tx.Rollback(context.Background()) }

type Rows struct {
	rows pgx.Rows
}

func (r *Rows) Next() bool             { return r.rows.Next() }
func (r *Rows) Scan(dest ...any) error { return r.rows.Scan(dest...) }
func (r *Rows) Close() error           { r.rows.Close(); return nil }
