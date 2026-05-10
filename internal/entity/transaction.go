package entity

import "time"

type TransactionType string

const (
	TransactionTypeAccrual    TransactionType = "ACCRUAL"    // начисление балов
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL" // списание балов
)

type Transaction struct {
	ID          int64 `gorm:"primaryKey;autoIncrement"`
	UserID      int64
	OrderNumber string
	Type        TransactionType `gorm:"type:transaction_type"`
	Amount      int64
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
