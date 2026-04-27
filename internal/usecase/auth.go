package usecase

import (
	"context"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockery --name=UserRepository --output=../mocks --outpkg=mocks --filename=user_repository.go
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
}

type TokenGenerator interface {
	Generate(userID int64) (string, error)
}

type AuthUseCase struct {
	repo           UserRepository
	tokenGenerator TokenGenerator
}

func NewAuthUseCase(repo UserRepository, tokenGenerator TokenGenerator) *AuthUseCase {
	return &AuthUseCase{repo: repo, tokenGenerator: tokenGenerator}
}

func (uc *AuthUseCase) Register(ctx context.Context, login string, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	user := entity.User{
		Login:        login,
		PasswordHash: string(hash),
	}
	if err := uc.repo.Create(ctx, &user); err != nil {
		return "", err
	}
	return uc.tokenGenerator.Generate(user.ID)
}

func (uc *AuthUseCase) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := uc.repo.FindByLogin(ctx, login)
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", entity.ErrInvalidCredentials
	}

	return uc.tokenGenerator.Generate(user.ID)
}
