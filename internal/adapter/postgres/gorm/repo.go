package postgres_gorm

import (
	"context"

	"gorm.io/gorm"
)

type baseRepo struct {
	txManager *TxManager
}

func (r *baseRepo) db(ctx context.Context) *gorm.DB {
	return r.txManager.GetDB(ctx)
}
