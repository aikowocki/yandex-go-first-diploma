package usecase

import (
	"context"
	"errors"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
)

//go:generate mockery --name=OrderRepository --output=../mocks --outpkg=mocks
type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	FindByNumber(ctx context.Context, number string) (*entity.Order, error)
	FindByUser(ctx context.Context, userID int64) ([]entity.Order, error)
}

type OrderUseCase struct {
	repo OrderRepository
}

func NewOrderUseCase(repo OrderRepository) *OrderUseCase {
	return &OrderUseCase{repo: repo}
}

func (uc *OrderUseCase) SubmitOrder(ctx context.Context, userID int64, number string) error {
	if !entity.ValidateLuhn(number) {
		return entity.ErrOrderNumberNotValid
	}

	order, err := uc.repo.FindByNumber(ctx, number)
	if err != nil && !errors.Is(err, entity.ErrOrderNotFound) {
		return err
	}
	if order != nil {
		if order.UserID != userID {
			return entity.ErrOrderNotBelongToUser // 409
		}
		return entity.ErrOrderAlreadySubmitted // 200
	}

	return uc.repo.Create(ctx, &entity.Order{
		UserID: userID,
		Number: number,
	}) // 202
}

func (uc *OrderUseCase) GetUserOrders(ctx context.Context, userID int64) ([]entity.Order, error) {
	return uc.repo.FindByUser(ctx, userID)
}
