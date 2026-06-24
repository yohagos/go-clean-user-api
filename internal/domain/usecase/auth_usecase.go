package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/dto"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
)

type AuthUseCase interface {
	Register(ctx context.Context, req *entity.RegisterRequest) (*entity.User, *dto.TokenResponse, error)
	Login(ctx context.Context, req *entity.LoginRequest) (*dto.TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
	ValidateToken(ctx context.Context, token string) (*uuid.UUID, error)
}
