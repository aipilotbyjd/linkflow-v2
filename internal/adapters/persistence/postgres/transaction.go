package postgres

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

func WithTransaction(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, txKey{}, tx)
		return fn(ctx)
	})
}

func GetTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}
