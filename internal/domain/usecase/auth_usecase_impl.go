package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/dto"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
	domain "github.com/yohagos/go-clean-user-api/internal/domain/repository"
	"github.com/yohagos/go-clean-user-api/internal/domain/service"
	"github.com/yohagos/go-clean-user-api/internal/logger"
)

type authUseCase struct {
	userRepo     domain.UserRepository
	tokenRepo    domain.TokenRepository
	tokenService service.TokenService
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	tokenRepo domain.TokenRepository,
	tokenService service.TokenService,
) AuthUseCase {
	return &authUseCase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		tokenService: tokenService,
	}
}

func (uc *authUseCase) Register(
	ctx context.Context,
	req *entity.RegisterRequest,
) (*entity.User, *dto.TokenResponse, error) {
	exists, err := uc.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, entity.ErrEmailExists
	}

	user := &entity.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Name:      req.Name,
		Password:  req.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := user.HashPassword(); err != nil {
		return nil, nil, err
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	tokens, err := uc.tokenService.GenerateTokenPair(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (uc *authUseCase) Login(
	ctx context.Context,
	req *entity.LoginRequest,
) (*dto.TokenResponse, error) {
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, entity.ErrUnauthorized
	}

	if err := user.CheckPassword(req.Password); err != nil {
		logger.Log.Error("CheckPassword for Login thrown an error: " + err.Error())
		return nil, entity.ErrUnauthorized
	}

	tokens, err := uc.tokenService.GenerateTokenPair(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (uc *authUseCase) RefreshToken(
	ctx context.Context,
	refreshToken string,
) (*dto.TokenResponse, error) {
	userID, err := uc.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	storedToken, err := uc.tokenRepo.GetByToken(ctx, refreshToken, entity.RefreshToken)
	if err != nil {
		return nil, entity.ErrInvalidToken
	}

	if storedToken.Revoked {
		return nil, entity.ErrRevokedToken
	}

	if time.Now().After(storedToken.ExpiredAt) {
		return nil, entity.ErrExpiredToken
	}

	if err := uc.tokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		return nil, err
	}

	newTokens, err := uc.tokenService.GenerateTokenPair(ctx, *userID)
	if err != nil {
		return nil, err
	}
	return newTokens, nil
}

func (uc *authUseCase) Logout(
	ctx context.Context,
	accessToken, refreshToken string,
) error {
	if err := uc.tokenService.RevokeToken(ctx, accessToken, entity.AccessToken); err != nil {
		return err
	}

	if err := uc.tokenService.RevokeToken(ctx, refreshToken, entity.RefreshToken); err != nil {
		return err
	}

	return nil
}

func (uc *authUseCase) ValidateToken(
	ctx context.Context,
	token string,
) (*uuid.UUID, error) {
	return uc.tokenService.ValidateAccessToken(token)
}
