package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (u *User) Validate() error {
	if u.Email == "" {
		return ErrInvalidEmail
	}
	if u.Name == "" {
		return ErrInvalidName
	}
	return nil
}
