package postgres

import "github.com/jackc/pgx/v5/pgxpool"

type baseRepo struct {
	pool *pgxpool.Pool
}
