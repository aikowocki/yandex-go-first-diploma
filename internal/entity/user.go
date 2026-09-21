package entity

import (
	"errors"
	"time"
)

type User struct {
	ID           int64 `gorm:"primaryKey;autoIncrement"`
	Login        string
	PasswordHash string
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

//GORM Если бы таблица не называлась users (structName -> snake_case -> множественное число) можем переопределить так
//func (User) TableName() string {
//	return "app_users"
//}

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
