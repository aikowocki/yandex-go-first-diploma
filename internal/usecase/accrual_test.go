package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAccrualClient struct{ mock.Mock }

func (m *mockAccrualClient) GetOrderAccrual(ctx context.Context, number string) (*AccrualResult, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccrualResult), args.Error(1)
}

type mockAccrualOrderRepo struct{ mock.Mock }

func (m *mockAccrualOrderRepo) FindPending(ctx context.Context) ([]entity.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entity.Order), args.Error(1)
}

func (m *mockAccrualOrderRepo) UpdateStatus(ctx context.Context, number string, status entity.OrderStatus, accrual *int64) error {
	args := m.Called(ctx, number, status, accrual)
	return args.Error(0)
}

type mockAccrualBalanceRepo struct{ mock.Mock }

func (m *mockAccrualBalanceRepo) AddAccrual(ctx context.Context, userID int64, orderNumber string, amount int64) error {
	args := m.Called(ctx, userID, orderNumber, amount)
	return args.Error(0)
}

func TestProcessOrder(t *testing.T) {
	orderNumber := "79927398713"
	tests := []struct {
		name           string
		accrualResp    *AccrualResult
		accrualErr     error
		wantStatus     entity.OrderStatus
		wantAccrual    *int64
		wantAddAccrual bool
		wantErr        bool
	}{
		{
			name:           "processed with accrual",
			accrualResp:    &AccrualResult{Order: orderNumber, Status: "PROCESSED", Accrual: 500.0},
			wantStatus:     entity.OrderStatusProcessed,
			wantAccrual:    new(int64(50000)),
			wantAddAccrual: true,
		},
		{
			name:        "processing",
			accrualResp: &AccrualResult{Order: orderNumber, Status: "PROCESSING"},
			wantStatus:  entity.OrderStatusProcessing,
		},
		{
			name:        "invalid",
			accrualResp: &AccrualResult{Order: orderNumber, Status: "INVALID"},
			wantStatus:  entity.OrderStatusInvalid,
		},
		{
			name:        "no content",
			accrualResp: nil,
		},
		{
			name:       "client error",
			accrualErr: errors.New("connection refused"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := new(mockAccrualClient)
			orderRepo := new(mockAccrualOrderRepo)
			balanceRepo := new(mockAccrualBalanceRepo)

			order := entity.Order{UserID: 1, Number: orderNumber}

			client.On("GetOrderAccrual", mock.Anything, order.Number).
				Return(tt.accrualResp, tt.accrualErr)

			if tt.wantStatus != "" {
				orderRepo.On("UpdateStatus", mock.Anything, order.Number, tt.wantStatus, tt.wantAccrual).
					Return(nil)
			}
			if tt.wantAddAccrual {
				balanceRepo.On("AddAccrual", mock.Anything, int64(1), order.Number, *tt.wantAccrual).
					Return(nil)
			}

			uc := NewAccrualUseCase(client, orderRepo, balanceRepo, &mockTxManager{})
			err := uc.ProcessOrder(context.Background(), order)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			client.AssertExpectations(t)
			orderRepo.AssertExpectations(t)
			balanceRepo.AssertExpectations(t)
		})
	}
}
