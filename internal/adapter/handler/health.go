package handler

import (
	"net/http"

	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/response"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

func (h *HealthHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.pool.Ping(r.Context()); err != nil {
		response.WriteError(w, http.StatusServiceUnavailable, "db unavailable")
		return
	}
	w.WriteHeader(http.StatusOK)
}
