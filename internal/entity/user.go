package entity

import (
	"errors"
	"time"
)

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
