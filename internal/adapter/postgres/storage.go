package postgres

import (
	"context"

	"github.com/aikowocki/yandex-go-first-diploma/internal/port"
)

type Storage struct {
	db          *DB
	txManager   *TxManager
	userRepo    *UserRepo
	orderRepo   *OrderRepo
	balanceRepo *BalanceRepo
}

func NewStorage(ctx context.Context, dsn string) (*Storage, error) {
	db, err := NewPool(ctx, dsn)
	if err != nil {
		return nil, err
	}

	txm := NewTxManager(db)
	return &Storage{
		db:          db,
		txManager:   txm,
		userRepo:    NewUserRepo(txm),
		orderRepo:   NewOrderRepo(txm),
		balanceRepo: NewBalanceRepo(txm),
	}, nil
}

func (s *Storage) DB() port.DB {
	return s.db
}

func (s *Storage) TxManager() port.TxManager {
	return s.txManager
}

func (s *Storage) UserRepo() port.UserRepository {
	return s.userRepo
}

func (s *Storage) OrderRepo() port.OrderRepository {
	return s.orderRepo
}

func (s *Storage) BalanceRepo() port.BalanceRepository {
	return s.balanceRepo
}
