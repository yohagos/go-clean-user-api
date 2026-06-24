package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
	domain "github.com/yohagos/go-clean-user-api/internal/domain/repository"
)

type tokenRepository struct {
	db *sqlx.DB
}

func NewTokenRepository(db *sqlx.DB) domain.TokenRepository {
	return &tokenRepository{
		db: db,
	}
}

func (r *tokenRepository) Create(ctx context.Context, token *entity.Token) error {
	query := `
		INSERT INTO tokens (id, user_id, token, type, expired_at, revoked, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.Token,
		token.Type,
		token.ExpiredAt,
		token.Revoked,
		token.CreatedAt,
		token.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	return nil
}

func (r *tokenRepository) GetByToken(
	ctx context.Context,
	tokenStr string,
	tokenType entity.TokenType,
) (*entity.Token, error) {
	query := `
		SELECT id, user_id, token, type, expired_at, revoked, created_at, updated_at
		FROM tokens
		WHERE token = $1 AND type = $2
	`

	var token entity.Token
	err := r.db.GetContext(ctx, &token, query, tokenStr, tokenType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return &token, nil
}

func (r *tokenRepository) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE tokens SET revoked = true, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return entity.ErrInvalidToken
	}

	return nil
}

func (r *tokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE tokens SET revoked = true, updated_at = NOW()
		WHERE user_id = $1 AND revoked = false
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	return nil
}

func (r *tokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM tokens WHERE expired_at < NOW()`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	return nil
}
