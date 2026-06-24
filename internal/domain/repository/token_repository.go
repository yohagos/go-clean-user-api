package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
)

type TokenRepository interface {
	Create(ctx context.Context, token *entity.Token) error
	GetByToken(ctx context.Context, token string, tokenType entity.TokenType) (*entity.Token, error)
	Revoke(ctx context.Context, tokenID uuid.UUID) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}
