package postgres_gorm

import (
	"context"
	"errors"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type UserRepo struct {
	baseRepo
}

func NewUserRepo(txManger *TxManager) *UserRepo {
	return &UserRepo{baseRepo: baseRepo{txManager: txManger}}
}

func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {

	if err := r.db(ctx).Create(user).Error; err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.ErrUserExists
		}

		return err
	}
	return nil
}

func (r *UserRepo) FindByLogin(ctx context.Context, login string) (*entity.User, error) {
	user := &entity.User{}

	if err := r.db(ctx).Where("login = ?", login).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}
