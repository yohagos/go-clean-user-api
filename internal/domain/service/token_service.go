package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/dto"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
)

type TokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
}

type TokenService interface {
	GenerateTokenPair(ctx context.Context, userID uuid.UUID) (*dto.TokenResponse, error)
	ValidateAccessToken(token string) (*uuid.UUID, error)
	ValidateRefreshToken(token string) (*uuid.UUID, error)
	RevokeToken(ctx context.Context, token string, tokenType entity.TokenType) error
	RevokeAllUserTokens(ctx context.Context,userID uuid.UUID) error
}
