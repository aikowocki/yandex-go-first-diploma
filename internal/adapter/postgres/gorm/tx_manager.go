package postgres_gorm

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

type TxManager struct {
	db *GormDB
}

func NewTxManager(db *GormDB) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctxWithTx := context.WithValue(ctx, txKey{}, tx)
		return fn(ctxWithTx)
	})
}

func (m *TxManager) GetDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return m.db.WithContext(ctx)
}
