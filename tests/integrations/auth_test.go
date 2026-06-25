package integrations

import (
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/dto"
)

func (s *TestSuite) TestAuthRegister_Success() {
	email := "test_" + time.Now().Format("20060102150405") + "@example.de"
	body := map[string]interface{}{
		"email":    email,
		"name":     "Test User",
		"password": "abcdef",
	}
	w := s.makeRequest("POST", "/api/v1/auth/register", body, "")

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	s.parseResponse(w, &response)

	user, ok := response["user"].(map[string]interface{})
	s.Require().True(ok, "user field not found or not a map")
	assert.Equal(s.T(), email, user["email"])
	assert.Equal(s.T(), "Test User", user["name"])
}

func (s *TestSuite) TestAuthRegister_DuplicateEmail() {
	email := "test_" + time.Now().Format("20060102150405") + "@example.de"
	body1 := map[string]interface{}{
		"email":    email,
		"name":     "First User",
		"password": "123456",
	}
	s.makeRequest("POST", "/api/v1/auth/register", body1, "")

	body2 := map[string]interface{}{
		"email":    email,
		"name":     "Second User",
		"password": "123456",
	}
	w := s.makeRequest("POST", "/api/v1/auth/register", body2, "")

	assert.Equal(s.T(), http.StatusConflict, w.Code)

	var response dto.ErrorResponse
	s.parseResponse(w, &response)
	assert.Equal(s.T(), http.StatusConflict, w.Code)
}

func (s *TestSuite) TestAuthRegister_InvalidEmail() {
	email := "test_" + time.Now().Format("20060102150405") + "-example"
	body := map[string]interface{}{
		"email":    email,
		"name":     "First User",
		"password": "123456",
	}
	w := s.makeRequest("POST", "/api/v1/auth/register", body, "")

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *TestSuite) TestAuthLogin_Success() {
	email := "test_" + time.Now().Format("20060102150405") + "@example.de"
	registerBody := map[string]interface{}{
		"email":    email,
		"name":     "Login User",
		"password": "123456",
	}
	s.makeRequest("POST", "/api/v1/auth/register", registerBody, "")

	loginBody := map[string]interface{}{
		"email":    email,
		"password": "123456",
	}
	w := s.makeRequest("POST", "/api/v1/auth/login", loginBody, "")

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response dto.TokenResponse
	s.parseResponse(w, &response)

	assert.NotEmpty(s.T(), response.AccessToken)
	assert.NotEmpty(s.T(), response.RefreshToken)
	assert.Equal(s.T(), int64(900), response.ExpiresIn)

	s.AccessToken = response.AccessToken
	s.RefreshToken = response.RefreshToken
}

func (s *TestSuite) TestAuthLogin_WrongPassword() {
	email := "test_" + time.Now().Format("20060102150405") + "@example.de"
	registerBody := map[string]interface{}{
		"email":    email,
		"name":     "Login User",
		"password": "123456",
	}
	s.makeRequest("POST", "/api/v1/auth/register", registerBody, "")

	loginBody := map[string]interface{}{
		"email":    email,
		"password": "654321",
	}
	w := s.makeRequest("POST", "/api/v1/auth/login", loginBody, "")

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *TestSuite) TestAuthLogin_UserNotFound() {
	email := "test_" + time.Now().Format("20060102150405") + "@example.de"
	loginBody := map[string]interface{}{
		"email":    email,
		"password": "123456",
	}
	w := s.makeRequest("POST", "/api/v1/auth/login", loginBody, "")

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *TestSuite) TestAuthRefreshToken_Success() {
	s.TestAuthLogin_Success()

	refreshBody := map[string]interface{}{
		"refresh_token": s.RefreshToken,
	}
	w := s.makeRequest("POST", "/api/v1/auth/refresh", refreshBody, "")

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response dto.TokenResponse
	s.parseResponse(w, &response)

	assert.NotEmpty(s.T(), response.AccessToken)
	assert.NotEmpty(s.T(), response.RefreshToken)
}

func (s *TestSuite) TestAuthRefreshToken_Invalid() {
	refreshBody := map[string]interface{}{
		"refresh_token": "invalid-token",
	}
	w := s.makeRequest("POST", "/api/v1/auth/refresh", refreshBody, "")

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *TestSuite) TestAuthLogout_Success() {
	s.TestAuthLogin_Success()

	w := s.makeRequest("POST", "/api/v1/auth/logout", nil, s.AccessToken, s.RefreshToken)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	w2 := s.makeRequest("GET", "/api/v1/users", nil, s.AccessToken)
	assert.Equal(s.T(), http.StatusOK, w2.Code)
}

func (s *TestSuite) TestProtectedRoute_NoToken() {
	w := s.makeRequest("GET", "/api/v1/users", nil, "")

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)

	var response dto.ErrorResponse
	s.parseResponse(w, &response)
	assert.Equal(s.T(), "unauthorized", response.Error)
}

func (s *TestSuite) TestProtectedRoute_InvalidToken() {
	w := s.makeRequest("GET", "/api/v1/users", nil, "invalid-token")

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}
