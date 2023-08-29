package dbcontext

import (
	"context"
	"database/sql"
)

type DB struct {
	db *sql.DB
}

type TransactionFunc func(ctx context.Context, f func(ctx context.Context) error) error

func New(db *sql.DB) *DB {
	return &DB{db}
}

func (db *DB) DB() *sql.DB {
	return db.db
}
