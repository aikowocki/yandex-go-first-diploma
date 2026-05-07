package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress        string `env:"RUN_ADDRESS"`
	DatabaseDSN          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JWTSecret            string `env:"JWT_SECRET"`
	CacheAddress         string `env:"CACHE_ADDRESS"`
	PostgresDriver       string `env:"POSTGRES_DRIVER"`
}

func New() (*Config, error) {
	cfg := &Config{}
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8081", "server address")
	flag.StringVar(&cfg.PostgresDriver, "db-driver", "pgx", "db driver(optional)")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database connection DSN")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "localhost:8080", "accrual system address")
	flag.StringVar(&cfg.JWTSecret, "s", "secret", "jwt secret")
	flag.StringVar(&cfg.CacheAddress, "c", "", "cache address(optional)")

	flag.Parse()
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
