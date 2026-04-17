package postgres

import (
	"context"
	"errors"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	baseRepo
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{baseRepo: baseRepo{pool: pool}}
}

func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	q := "INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, created_at, updated_at"
	err := r.pool.QueryRow(ctx, q, user.Login, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
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
	q := "SELECT id, password_hash, created_at, updated_at FROM users WHERE login = $1"

	err := r.pool.QueryRow(ctx, q, login).
		Scan(&user.ID, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}
