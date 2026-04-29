package usecase

import (
	"context"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/aikowocki/yandex-go-first-diploma/internal/port"
)

type AccrualClient interface {
	GetOrderAccrual(ctx context.Context, number string) (*AccrualResult, error)
}

type AccrualResult struct {
	Order   string
	Status  string
	Accrual float64
}

type AccrualOrderRepo interface {
	FindPending(ctx context.Context) ([]entity.Order, error)
	UpdateStatus(ctx context.Context, number string, status entity.OrderStatus, accrual *int64) error
}

type AccrualBalanceRepo interface {
	AddAccrual(ctx context.Context, userID int64, orderNumber string, amount int64) error
}

type AccrualUseCase struct {
	client      AccrualClient
	orderRepo   AccrualOrderRepo
	balanceRepo AccrualBalanceRepo
	txManager   port.TxManager
}

func NewAccrualUseCase(client AccrualClient, orderRepo AccrualOrderRepo, balanceRepo AccrualBalanceRepo, txManager port.TxManager) *AccrualUseCase {
	return &AccrualUseCase{
		client:      client,
		orderRepo:   orderRepo,
		balanceRepo: balanceRepo,
		txManager:   txManager,
	}
}

func (uc *AccrualUseCase) GetPendingOrders(ctx context.Context) ([]entity.Order, error) {
	return uc.orderRepo.FindPending(ctx)
}

func (uc *AccrualUseCase) ProcessOrder(ctx context.Context, order entity.Order) error {
	resp, err := uc.client.GetOrderAccrual(ctx, order.Number)
	if err != nil {
		return err
	}
	if resp == nil {
		return nil
	}

	status := mapAccrualStatus(resp.Status)

	if status == entity.OrderStatusProcessed {
		accrualAmount := int64(resp.Accrual * 100)

		return uc.txManager.Do(ctx, func(ctx context.Context) error {
			if err := uc.orderRepo.UpdateStatus(ctx, order.Number, status, &accrualAmount); err != nil {
				return err
			}
			return uc.balanceRepo.AddAccrual(ctx, order.UserID, order.Number, accrualAmount)
		})
	}

	return uc.orderRepo.UpdateStatus(ctx, order.Number, status, nil)
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
