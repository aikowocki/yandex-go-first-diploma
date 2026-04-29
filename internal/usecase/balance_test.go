package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/aikowocki/yandex-go-first-diploma/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTxManager struct{}

func (m *mockTxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestWithdraw_InvalidLuhn(t *testing.T) {
	repo := new(mocks.BalanceRepository)
	uc := NewBalanceUseCase(repo, &mockTxManager{})

	err := uc.Withdraw(context.Background(), 1, "12345", 100)

	assert.ErrorIs(t, err, entity.ErrOrderNumberNotValid)
}

func TestWithdraw_Success(t *testing.T) {
	repo := new(mocks.BalanceRepository)
	repo.On("LockByUserId", mock.Anything, int64(1)).Return(nil)
	repo.On("GetBalance", mock.Anything, int64(1)).Return(entity.Balance{Current: 500}, nil)
	repo.On("Withdraw", mock.Anything, int64(1), "79927398713", int64(100)).Return(nil)

	uc := NewBalanceUseCase(repo, &mockTxManager{})
	err := uc.Withdraw(context.Background(), 1, "79927398713", 100)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	repo := new(mocks.BalanceRepository)
	repo.On("LockByUserId", mock.Anything, int64(1)).Return(nil)
	repo.On("GetBalance", mock.Anything, int64(1)).Return(entity.Balance{Current: 50}, nil)

	uc := NewBalanceUseCase(repo, &mockTxManager{})
	err := uc.Withdraw(context.Background(), 1, "79927398713", 100)

	assert.ErrorIs(t, err, entity.ErrBalanceInsufficientFunds)
	repo.AssertExpectations(t)
}

func TestWithdraw_LockError(t *testing.T) {
	repo := new(mocks.BalanceRepository)
	repo.On("LockByUserId", mock.Anything, int64(1)).Return(errors.New("failed"))

	uc := NewBalanceUseCase(repo, &mockTxManager{})
	err := uc.Withdraw(context.Background(), 1, "79927398713", 100)

	assert.Error(t, err)
	repo.AssertExpectations(t)
}
