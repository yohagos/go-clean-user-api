package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
	domain "github.com/yohagos/go-clean-user-api/internal/domain/repository"
)

type userUseCase struct {
	userRepo domain.UserRepository
}

func NewUserUseCase(repo domain.UserRepository) UserUseCase {
	return &userUseCase{
		userRepo: repo,
	}
}

func (uc *userUseCase) CreateUser(ctx context.Context, email, name string) (*entity.User, error) {
	if email == "" {
		return nil, entity.ErrInvalidEmail
	}
	if name == "" {
		return nil, entity.ErrInvalidName
	}

	exists, err := uc.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, entity.ErrEmailExists
	}

	user := &entity.User{
		ID:        uuid.New(),
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *userUseCase) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *userUseCase) GetAllUsers(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return uc.userRepo.GetAll(ctx, limit, offset)
}

func (uc *userUseCase) UpdateUser(
	ctx context.Context,
	id uuid.UUID,
	email, name string,
) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if email != "" {
		if email != user.Email {
			exists, err := uc.userRepo.ExistsByEmail(ctx, email)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, entity.ErrEmailExists
			}
			user.Email = email
		}
	}
	if name != "" {
		user.Name = name
	}

	user.UpdatedAt = time.Now()

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *userUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.userRepo.Delete(ctx, id)
}
