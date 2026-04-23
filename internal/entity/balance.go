package entity

import "errors"

type Balance struct {
	Current   int64
	Withdrawn int64
}

var (
	ErrBalanceInsufficientFunds = errors.New("insufficient funds")
	ErrWithdrawalAlreadyExists  = errors.New("withdrawal already exists")
	ErrAccrualAlreadyExists     = errors.New("accrual already exists")
)
