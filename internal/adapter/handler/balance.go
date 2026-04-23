package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/aikowocki/yandex-go-first-diploma/internal/adapter/handler/middleware"
	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/response"
	"github.com/aikowocki/yandex-go-first-diploma/internal/usecase"
	"go.uber.org/zap"
)

type BalanceHandler struct {
	uc *usecase.BalanceUseCase
}

func NewBalanceHandler(uc *usecase.BalanceUseCase) *BalanceHandler {
	return &BalanceHandler{uc: uc}
}

type balanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type withdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := middleware.MustGetUserID(r.Context())

	balance, err := h.uc.GetBalance(r.Context(), userID)

	if err != nil {
		zap.S().Errorw("failed to get balance", "error", err, "userID", userID)
		response.WriteError(w, http.StatusInternalServerError, "Internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(
		balanceResponse{
			Current:   float64(balance.Current) / 100,
			Withdrawn: float64(balance.Withdrawn) / 100,
		}); err != nil {
		zap.S().Errorw("failed to encode response", "error", err)
	}
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req withdrawRequest

	if err := json.NewDecoder(io.LimitReader(r.Body, 1024)).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad request")
		return
	}

	if req.Sum <= 0 {
		response.WriteError(w, http.StatusBadRequest, "invalid sum")
		return
	}

	userID := middleware.MustGetUserID(r.Context())
	amount := int64(req.Sum * 100)

	err := h.uc.Withdraw(r.Context(), userID, req.Order, amount)

	if err != nil {
		switch true {
		case errors.Is(err, entity.ErrOrderNumberNotValid):
			response.WriteError(w, http.StatusUnprocessableEntity, "invalid order number")
			return
		case errors.Is(err, entity.ErrBalanceInsufficientFunds):
			response.WriteError(w, http.StatusPaymentRequired, "insufficient funds")
			return
		default:
			zap.S().Errorw("failed to wighdraw", "error", err, "userID", userID)
			response.WriteError(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := middleware.MustGetUserID(r.Context())

	withdrawals, err := h.uc.GetWithdrawals(r.Context(), userID)

	if err != nil {
		zap.S().Errorw("failed to get withdrawals", "error", err, "userID", userID)
		response.WriteError(w, http.StatusInternalServerError, "Internal error")
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(toWithdrawalsResponse(withdrawals)); err != nil {
		zap.S().Errorw("failed to encode response", "error", err)
	}
}

func toWithdrawalsResponse(withdrawals []entity.Transaction) []withdrawalResponse {
	result := make([]withdrawalResponse, len(withdrawals))

	for i, w := range withdrawals {
		result[i] = withdrawalResponse{
			Order:       w.OrderNumber,
			Sum:         float64(w.Amount) / 100,
			ProcessedAt: w.CreatedAt.Format(time.RFC3339),
		}
	}
	return result
}
