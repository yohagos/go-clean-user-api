package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
)

type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	Name  string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
}

type UpdateUserRequest struct {
	Email string `json:"email" binding:"omitempty,email" example:"updated@example.com"`
	Name  string `json:"name" binding:"omitempty,min=2,max=100" example:"Jane Doe"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id" example:"123abcde-5678-1234-abcd-123456abcdef"`
	Email     string    `json:"email" example:"user@example.com"`
	Name      string    `json:"name" example:"Jphn Doe"`
	CreatedAt time.Time `json:"created_at" example:"2026-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-01-01T00:00:00Z"`
}

type UsersListResponse struct {
	Users      []UserResponse `json:"users"`
	TotalCount int            `json:"total_count"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
}

type ErrorResponse struct {
	Error   string `json:"error" example:"user not found"`
	Code    int    `json:"code" example:"404"`
	Details string `json:"details,omitempty"`
}

func ToUserResponse(user *entity.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func ToUserResponseList(users []*entity.User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToUserResponse(user)
	}
	return responses
}
