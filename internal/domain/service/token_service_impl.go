package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/dto"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
	domain "github.com/yohagos/go-clean-user-api/internal/domain/repository"
)

type tokenService struct {
	tokenRepo     domain.TokenRepository
	userRepo      domain.UserRepository
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewTokenService(
	tokenRepo domain.TokenRepository,
	userRepo domain.UserRepository,
	config TokenConfig,
) TokenService {
	return &tokenService{
		tokenRepo:     tokenRepo,
		userRepo:      userRepo,
		accessSecret:  config.AccessTokenSecret,
		refreshSecret: config.RefreshTokenSecret,
		accessTTL:     config.AccessTokenTTL,
		refreshTTL:    config.RefreshTokenTTL,
	}
}

func (s *tokenService) GenerateTokenPair(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.TokenResponse, error) {
	accToken, accJTI, err := s.generateJWT(userID, s.accessSecret, s.accessTTL)
	if err != nil {
		return nil, err
	}

	refToken, refJTI, err := s.generateJWT(userID, s.refreshSecret, s.refreshTTL)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	accessTokenEntity := &entity.Token{
		ID:        accJTI,
		UserID:    userID,
		Token:     accToken,
		Type:      entity.AccessToken,
		ExpiredAt: now.Add(s.accessTTL),
		Revoked:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	refreshTokenEntity := &entity.Token{
		ID:        refJTI,
		UserID:    userID,
		Token:     refToken,
		Type:      entity.RefreshToken,
		ExpiredAt: now.Add(s.refreshTTL),
		Revoked:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.tokenRepo.Create(ctx, accessTokenEntity); err != nil {
		return nil, err
	}

	if err := s.tokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accToken,
		RefreshToken: refToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *tokenService) generateJWT(
	userID uuid.UUID,
	secret string,
	ttl time.Duration,
) (string, uuid.UUID, error) {
	jti := uuid.New()
	now := time.Now()
	claims := TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", uuid.Nil, err
	}

	return tokenString, jti, nil
}

func (s *tokenService) ValidateAccessToken(token string) (*uuid.UUID, error) {
	return s.validateToken(token, s.accessSecret)
}

func (s *tokenService) ValidateRefreshToken(token string) (*uuid.UUID, error) {
	return s.validateToken(token, s.refreshSecret)
}

func (s *tokenService) validateToken(tokenString, secret string) (*uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, entity.ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, entity.ErrInvalidToken
	}
	return &claims.UserID, nil
}

func (s *tokenService) RevokeToken(ctx context.Context, tokenString string, tokenType entity.TokenType) error {
	token, err := s.tokenRepo.GetByToken(ctx, tokenString, tokenType)
	if err != nil {
		return err
	}
	return s.tokenRepo.Revoke(ctx, token.ID)
}

func (s *tokenService) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return s.tokenRepo.RevokeAllUserTokens(ctx, userID)
}
