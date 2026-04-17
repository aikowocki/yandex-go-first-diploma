package usecase

import (
	"context"
	"testing"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/aikowocki/yandex-go-first-diploma/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubmitOrder_InvalidLuhn(t *testing.T) {
	repo := new(mocks.OrderRepository)
	uc := NewOrderUseCase(repo)

	err := uc.SubmitOrder(context.Background(), 1, "12345")

	assert.ErrorIs(t, err, entity.ErrOrderNumberNotValid)
}

func TestSubmitOrder_NewOrder(t *testing.T) {
	repo := new(mocks.OrderRepository)
	repo.On("FindByNumber", mock.Anything, "79927398713").Return(nil, entity.ErrOrderNotFound)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)

	uc := NewOrderUseCase(repo)
	err := uc.SubmitOrder(context.Background(), 1, "79927398713")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSubmitOrder_AlreadySubmittedByUser(t *testing.T) {
	order := &entity.Order{UserID: 1, Number: "79927398713"}

	repo := new(mocks.OrderRepository)
	repo.On("FindByNumber", mock.Anything, "79927398713").Return(order, nil)

	uc := NewOrderUseCase(repo)
	err := uc.SubmitOrder(context.Background(), 1, "79927398713")

	assert.ErrorIs(t, err, entity.ErrOrderAlreadySubmitted)
}

func TestSubmitOrder_BelongsToOtherUser(t *testing.T) {
	order := &entity.Order{UserID: 2, Number: "79927398713"}

	repo := new(mocks.OrderRepository)
	repo.On("FindByNumber", mock.Anything, "79927398713").Return(order, nil)

	uc := NewOrderUseCase(repo)
	err := uc.SubmitOrder(context.Background(), 1, "79927398713")

	assert.ErrorIs(t, err, entity.ErrOrderNotBelongToUser)
}
