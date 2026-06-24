package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/yohagos/go-clean-user-api/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Password  string    `json:"password" db:"password"`
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

func (u *User) HashPassword() error {
	logger.Log.Info(
		"Hashing User Password",
		zap.Any("User Object", u),
	)
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(u.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	u.Password = string(hashed)
	logger.Log.Info(
		"User Password was hashed successfully",
		zap.Any("Hashed User Passw rod", u.Password),
	)
	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(u.Password),
		[]byte(password),
	)
}
