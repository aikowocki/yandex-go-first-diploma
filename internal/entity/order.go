package entity

import "time"

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	ID        int64
	UserID    int64
	Number    string
	Status    OrderStatus
	Accrual   *int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

var doubled = [10]int{0, 2, 4, 6, 8, 1, 3, 5, 7, 9}

func ValidateLuhn(number string) bool {
	if len(number) == 0 {
		return false
	}
	sum := 0
	lastIdx := len(number) - 1
	for i := lastIdx; i >= 0; i-- {
		if number[i] < '0' || number[i] > '9' {
			return false
		}
		digit := int(number[i] - '0')
		if (lastIdx-i)%2 == 1 {
			digit = doubled[digit]
		}
		sum += int(digit)
	}

	return sum%10 == 0
}
