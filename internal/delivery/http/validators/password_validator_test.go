package validators

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Password string `validate:"password_strength"`
}

func TestPasswordStrengthValidator(t *testing.T) {
	v := validator.New()
	v.RegisterValidation("password_strength", PasswordStrengthValidator)

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"Valid password", "Test123!@#", true},
		{"too short", "Tes3!", false},
		{"No uppercase", "test123!@#", false},
		{"No lowercase ", "TEST123!@#", false},
		{"No number ", "Testdsdf!@#", false},
		{"No special ", "Test12312313", false},
		{"Only letters ", "Testtesttest", false},
		{"Valid with spaces ", "Test 123!@#", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(TestStruct{Password: tt.password})
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
