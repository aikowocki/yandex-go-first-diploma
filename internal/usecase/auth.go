package usecase

import (
	"context"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockery --name=UserRepository --output=../mocks --outpkg=mocks
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
}

type AuthUseCase struct {
	repo       UserRepository
	jwtManager *auth.JWTManager
}

func NewAuthUseCase(repo UserRepository, jwtManager *auth.JWTManager) *AuthUseCase {
	return &AuthUseCase{repo: repo, jwtManager: jwtManager}
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
	return uc.jwtManager.Generate(user.ID)
}

func (uc *AuthUseCase) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := uc.repo.FindByLogin(ctx, login)
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", entity.ErrInvalidCredentials
	}

	return uc.jwtManager.Generate(user.ID)
}
