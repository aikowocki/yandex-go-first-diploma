package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aikowocki/yandex-go-first-diploma/internal/adapter/handler/middleware"
	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/response"
	"github.com/aikowocki/yandex-go-first-diploma/internal/usecase"
	"go.uber.org/zap"
)

type OrderHandler struct {
	uc *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

type orderResponse struct {
	Number    string   `json:"number"`
	Status    string   `json:"status"`
	Accrual   *float64 `json:"accrual,omitempty"`
	UpdatedAt string   `json:"uploaded_at"`
}

func toOrderListResponse(orders []entity.Order) []orderResponse {
	result := make([]orderResponse, len(orders))

	for i, o := range orders {
		result[i] = orderResponse{
			Number:    o.Number,
			Status:    string(o.Status),
			UpdatedAt: o.CreatedAt.Format(time.RFC3339),
		}

		if o.Accrual != nil {
			accrual := float64(*o.Accrual) / 100
			result[i].Accrual = &accrual
		}
	}
	return result
}

func (h *OrderHandler) SubmitOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 256))

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	number := strings.TrimSpace(string(body))
	if number == "" {
		response.WriteError(w, http.StatusBadRequest, entity.ErrOrderNumberNotValid.Error())
		return
	}

	userID := middleware.MustGetUserID(r.Context())

	if err := h.uc.SubmitOrder(r.Context(), userID, number); err != nil {
		switch true {
		case errors.Is(err, entity.ErrOrderAlreadySubmitted):
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, entity.ErrOrderNotBelongToUser):
			response.WriteError(w, http.StatusConflict, entity.ErrOrderNotBelongToUser.Error())
		case errors.Is(err, entity.ErrOrderNumberNotValid):
			response.WriteError(w, http.StatusUnprocessableEntity, entity.ErrOrderNumberNotValid.Error())
		default:
			zap.S().Errorw("failed to submit order",
				"error", err,
				"userID", userID,
				"number", number,
			)
			response.WriteError(w, http.StatusInternalServerError, "Internal error")
		}
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID := middleware.MustGetUserID(r.Context())

	orders, err := h.uc.GetUserOrders(r.Context(), userID)
	if err != nil {
		zap.S().Errorw("failed to get user orders",
			"error", err,
			"userID", userID,
		)
		response.WriteError(w, http.StatusInternalServerError, "Internal error")
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(toOrderListResponse(orders))
	if err != nil {
		zap.S().Errorw("failed to encode resposne", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "Internal error")
	}
}
