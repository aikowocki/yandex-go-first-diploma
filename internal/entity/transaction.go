package entity

import "time"

type TransactionType int

const (
	TransactionTypeAccrual    TransactionType = 0 // начисление балов
	TransactionTypeWithdrawal TransactionType = 1 // спсиание балов
)

type Transaction struct {
	ID          int64
	UserID      int64
	OrderNumber string
	Type        TransactionType
	Amount      int64
	CreatedAt   time.Time
}
