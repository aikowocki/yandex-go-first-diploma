package entity

import "time"

type TransactionType string

const (
	TransactionTypeAccrual    TransactionType = "ACCRUAL"    // начисление балов
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL" // списание балов
)

type Transaction struct {
	ID          int64
	UserID      int64
	OrderNumber string
	Type        TransactionType
	Amount      int64
	CreatedAt   time.Time
}
