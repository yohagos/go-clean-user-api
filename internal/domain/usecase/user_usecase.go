package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
)

type UserUseCase interface {
	CreateUser(ctx context.Context, email, name string) (*entity.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetAllUsers(ctx context.Context, limit, offset int) ([]*entity.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, email, name string) (*entity.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
