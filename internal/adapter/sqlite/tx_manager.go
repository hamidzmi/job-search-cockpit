package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

type txKey struct{}

func (s *Store) exec(ctx context.Context) dbExecutor {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return s.db
}

func (s *Store) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return fn(ctx)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
