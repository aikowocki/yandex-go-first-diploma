package handler

import (
	"context"
	"net/http"

	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/response"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	pinger Pinger
}

func NewHealthHandler(pinger Pinger) *HealthHandler {
	return &HealthHandler{pinger: pinger}
}

func (h *HealthHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.pinger.Ping(r.Context()); err != nil {
		response.WriteError(w, http.StatusServiceUnavailable, "db unavailable")
		return
	}
	w.WriteHeader(http.StatusOK)
}
