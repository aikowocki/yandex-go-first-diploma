package postgres_gorm

import (
	"context"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormDB struct {
	*gorm.DB
}

func NewDB(dsn string) (*GormDB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	return &GormDB{db}, nil
}

func (db *GormDB) Ping(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (db *GormDB) Close() {
	sqlDB, _ := db.DB.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}
