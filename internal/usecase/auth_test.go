package usecase

import (
	"context"
	"testing"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/aikowocki/yandex-go-first-diploma/internal/mocks"
	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_Success(t *testing.T) {
	repo := new(mocks.UserRepository)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)

	jwtManager := auth.NewJWTManager("secret")
	uc := NewAuthUseCase(repo, jwtManager)

	token, err := uc.Register(context.Background(), "user", "pass")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	repo.AssertExpectations(t)
}

func TestRegister_Duplicate(t *testing.T) {
	repo := new(mocks.UserRepository)
	repo.On("Create", mock.Anything, mock.Anything).Return(entity.ErrUserExists)

	jwtManager := auth.NewJWTManager("secret")
	uc := NewAuthUseCase(repo, jwtManager)

	_, err := uc.Register(context.Background(), "user", "pass")

	assert.ErrorIs(t, err, entity.ErrUserExists)
}

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.MinCost)
	user := &entity.User{ID: 1, Login: "user", PasswordHash: string(hash)}

	repo := new(mocks.UserRepository)
	repo.On("FindByLogin", mock.Anything, "user").Return(user, nil)

	jwtManager := auth.NewJWTManager("secret")
	uc := NewAuthUseCase(repo, jwtManager)

	token, err := uc.Login(context.Background(), "user", "pass")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	repo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := new(mocks.UserRepository)
	repo.On("FindByLogin", mock.Anything, "user").Return(nil, entity.ErrUserNotFound)

	jwtManager := auth.NewJWTManager("secret")
	uc := NewAuthUseCase(repo, jwtManager)

	_, err := uc.Login(context.Background(), "user", "pass")

	assert.ErrorIs(t, err, entity.ErrUserNotFound)
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.MinCost)
	user := &entity.User{ID: 1, Login: "user", PasswordHash: string(hash)}

	repo := new(mocks.UserRepository)
	repo.On("FindByLogin", mock.Anything, "user").Return(user, nil)

	jwtManager := auth.NewJWTManager("secret")
	uc := NewAuthUseCase(repo, jwtManager)

	_, err := uc.Login(context.Background(), "user", "wrongpass")

	assert.ErrorIs(t, err, entity.ErrInvalidCredentials)
}
