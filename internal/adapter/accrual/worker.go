package accrual

import (
	"context"
	"errors"
	"time"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"go.uber.org/zap"
)

type OrderRepo interface {
	FindPending(ctx context.Context) ([]entity.Order, error)
	UpdateStatus(ct context.Context, number string, status entity.OrderStatus, accrual *int64) error
}

type BalanceRepo interface {
	AddAccrual(ctx context.Context, userID int64, orderNumber string, amount int64) error
}

type TxManager interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}

type Worker struct {
	client      *Client
	orderRepo   OrderRepo
	balanceRepo BalanceRepo
	txManager   TxManager
}

func NewWorker(client *Client, orderRepo OrderRepo, balanceRepo BalanceRepo, txManager TxManager) *Worker {
	return &Worker{
		client:      client,
		orderRepo:   orderRepo,
		balanceRepo: balanceRepo,
		txManager:   txManager,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processOrders(ctx)
		}
	}

}

func (w *Worker) processOrders(ctx context.Context) {
	orders, err := w.orderRepo.FindPending(ctx)

	if err != nil {
		zap.S().Errorw("failed to find pending orders", "error", err)
		return
	}

	for _, order := range orders {
		if ctx.Err() != nil {
			zap.S().Infow("worker stoppping, skipping remaining orders")
			return
		}

		resp, err := w.client.GetOrderAccrual(ctx, order.Number)
		if err != nil {
			var tooMany *ErrTooManyRequests
			if errors.As(err, &tooMany) {
				zap.S().Infow("rate limited, sleeping", "duration", tooMany.RetryAfter)
				time.Sleep(tooMany.RetryAfter)
				return
			}
			zap.S().Errorw("failed to get accrual", "error", err, "order", order.Number)
			continue
		}

		if resp == nil {
			continue
		}

		if status := mapAccrualStatus(resp.Status); status == entity.OrderStatusProcessed {
			accrualAmount := int64(resp.Accrual * 100)

			err = w.txManager.Do(ctx, func(ctx context.Context) error {
				if err := w.orderRepo.UpdateStatus(ctx, order.Number, status, &accrualAmount); err != nil {
					return err
				}
				return w.balanceRepo.AddAccrual(ctx, order.UserID, order.Number, accrualAmount)
			})
		} else {
			err = w.orderRepo.UpdateStatus(ctx, order.Number, status, nil)
		}

		if err != nil {
			zap.S().Errorw("failed to update order", "error", err, "order", order.Number)
		}
	}
}

func mapAccrualStatus(accrualStatus string) entity.OrderStatus {
	switch accrualStatus {
	case "REGISTERED":
		return entity.OrderStatusNew
	case "PROCESSING":
		return entity.OrderStatusProcessing
	case "INVALID":
		return entity.OrderStatusInvalid
	case "PROCESSED":
		return entity.OrderStatusProcessed
	default:
		return entity.OrderStatusNew
	}
}
