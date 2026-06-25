package integrations

import (
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/dto"
)

func (s *TestSuite) TestUserCreate_Success() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	body := map[string]interface{}{
		"email": "new_" + time.Now().Format("20060102150405") + "@example.de",
		"name":  "New User",
	}

	w := s.makeRequest("POST", "/api/v1/users", body, s.AccessToken)

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var response dto.UserResponse
	s.parseResponse(w, &response)
	assert.NotEmpty(s.T(), response.ID)
	assert.Equal(s.T(), body["email"], response.Email)
	assert.Equal(s.T(), body["name"], response.Name)
}

func (s *TestSuite) TestUserCreate_DuplicateEmail() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	userEmail := "duplicate_" + time.Now().Format("20060102150405") + "@example.de"

	body1 := map[string]interface{}{
		"email": userEmail,
		"name":  "First User",
	}
	s.makeRequest("POST", "/api/v1/users", body1, s.AccessToken)

	body2 := map[string]interface{}{
		"email": userEmail,
		"name":  "Second User",
	}
	w := s.makeRequest("POST", "/api/v1/users", body2, s.AccessToken)

	assert.Equal(s.T(), http.StatusConflict, w.Code)

	var response dto.ErrorResponse
	s.parseResponse(w, &response)
	assert.Equal(s.T(), "email already exists", response.Error)
}

func (s *TestSuite) TestUserCreate_InvalidEmail() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	body1 := map[string]interface{}{
		"email": "invalid-email",
		"name":  "invalid User",
	}
	w := s.makeRequest("POST", "/api/v1/users", body1, s.AccessToken)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	s.parseResponse(w, &response)
	assert.Equal(s.T(), "invalid request", response.Error)
}

func (s *TestSuite) TestUserCreate_Unauthorized() {
	body1 := map[string]interface{}{
		"email": "test@example.de",
		"name":  "unauthorized User",
	}
	w := s.makeRequest("POST", "/api/v1/users", body1, "")

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)

	var response dto.ErrorResponse
	s.parseResponse(w, &response)
	assert.Equal(s.T(), "unauthorized", response.Error)
}

func (s *TestSuite) TestUserGetByID_Success() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	userEmail := "get_" + time.Now().Format("20060102150405") + "@example.de"

	createBody := map[string]interface{}{
		"email": userEmail,
		"name":  "Get User",
	}
	wCreate := s.makeRequest("POST", "/api/v1/users", createBody, s.AccessToken)

	var createResponse dto.UserResponse
	s.parseResponse(wCreate, &createResponse)
	userID := createResponse.ID

	w := s.makeRequest("GET", "/api/v1/users/"+userID.String(), nil, s.AccessToken)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response dto.UserResponse
	s.parseResponse(w, &response)
	assert.Equal(s.T(), userID, response.ID)
	assert.Equal(s.T(), userEmail, response.Email)
	assert.Equal(s.T(), "Get User", response.Name)
}

func (s *TestSuite) TestUserGetByID_NotFound() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	w := s.makeRequest("GET", "/api/v1/users/00000000-0000-0000-0000-000000000000", nil, s.AccessToken)
	assert.Equal(s.T(), http.StatusNotFound, w.Code)

	var response dto.ErrorResponse
	s.parseResponse(w, &response)
	assert.Equal(s.T(), "user not found", response.Error)
}

func (s *TestSuite) TestUserGetByID_InvalidUUID() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	w := s.makeRequest("GET", "/api/v1/users/invalid-uuid", nil, s.AccessToken)
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	s.parseResponse(w, &response)
	assert.Equal(s.T(), "invalid user id", response.Error)
}

func (s *TestSuite) TestUserGetByID_Unauthorized() {
	w := s.makeRequest("GET", "/api/v1/users/00000000-0000-0000-0000-000000000000", nil, s.AccessToken)
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *TestSuite) TestUserGetAll_Success() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	for i := 0; i < 3; i++ {
		body := map[string]interface{}{
			"email": "list_" + time.Now().Format("20060102150405") + "@example.de",
			"name":  "List User",
		}
		s.makeRequest("POST", "/api/v1/users", body, s.AccessToken)
		time.Sleep(1 * time.Second)
	}

	w := s.makeRequest("GET", "/api/v1/users?limit=10&offset=0", nil, s.AccessToken)
	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response dto.UsersListResponse
	s.parseResponse(w, &response)

	assert.GreaterOrEqual(s.T(), len(response.Users), 3)
	assert.Equal(s.T(), 10, response.Limit)
	assert.Equal(s.T(), 0, response.Offset)
}

func (s *TestSuite) TestUserGetAll_WithPagination() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	for i := 0; i < 3; i++ {
		body := map[string]interface{}{
			"email": "Page_" + time.Now().Format("20060102150405") + "@example.de",
			"name":  "List User",
		}
		s.makeRequest("POST", "/api/v1/users", body, s.AccessToken)
		time.Sleep(1 * time.Millisecond)
	}

	w1 := s.makeRequest("GET", "/api/v1/users?limit=2&offset=0", nil, s.AccessToken)
	assert.Equal(s.T(), http.StatusOK, w1.Code)

	var response dto.UsersListResponse
	s.parseResponse(w1, &response)
	assert.LessOrEqual(s.T(), len(response.Users), 2)
	assert.Equal(s.T(), 2, response.Limit)
	assert.Equal(s.T(), 0, response.Offset)
}

func (s *TestSuite) TestUserGetAll_Unauthorized() {
	w1 := s.makeRequest("GET", "/api/v1/users?limit=2&offset=0", nil, "")
	assert.Equal(s.T(), http.StatusUnauthorized, w1.Code)
}

func (s *TestSuite) TestUserUpdate_Success() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	userEmail := "update_" + time.Now().Format("20060102150405") + "@example.de"

	createUser := map[string]interface{}{
		"email": userEmail,
		"name":  "Old name",
	}
	wCreate := s.makeRequest("POST", "/api/v1/users", createUser, s.AccessToken)

	var createResponse dto.UserResponse
	s.parseResponse(wCreate, &createResponse)

	userID := createResponse.ID

	updateBody := map[string]interface{}{
		"email": "update_" + time.Now().Format("20060102150405") + "@example.de",
		"name":  "New name",
	}

	wUpdate := s.makeRequest("PUT", "/api/v1/users/"+userID.String(), updateBody, s.AccessToken)
	assert.Equal(s.T(), http.StatusOK, wUpdate.Code)

	var response dto.UserResponse
	s.parseResponse(wUpdate, &response)
	assert.Equal(s.T(), userID, response.ID)
	assert.Equal(s.T(), updateBody["email"], response.Email)
	assert.Equal(s.T(), updateBody["name"], response.Name)
}

func (s *TestSuite) TestUserUpdate_Partial() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	userEmail := "update_" + time.Now().Format("20060102150405") + "@example.de"

	createUser := map[string]interface{}{
		"email": userEmail,
		"name":  "Original name",
	}
	wCreate := s.makeRequest("POST", "/api/v1/users", createUser, s.AccessToken)

	var createResponse dto.UserResponse
	s.parseResponse(wCreate, &createResponse)

	userID := createResponse.ID

	updateBody := map[string]interface{}{
		"name": "Only Name Changed",
	}

	wUpdate := s.makeRequest("PUT", "/api/v1/users/"+userID.String(), updateBody, s.AccessToken)
	assert.Equal(s.T(), http.StatusOK, wUpdate.Code)

	var response dto.UserResponse
	s.parseResponse(wUpdate, &response)
	assert.Equal(s.T(), userID, response.ID)
	assert.Equal(s.T(), userEmail, response.Email)
	assert.Equal(s.T(), updateBody["name"], response.Name)
}

func (s *TestSuite) TestUserUpdate_NotFound() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	updateBody := map[string]interface{}{
		"email": "test@example.de",
		"name":  "Test",
	}
	w := s.makeRequest("PUT", "/api/v1/users/00000000-0000-0000-0000-000000000000", updateBody, s.AccessToken)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)

	var response dto.ErrorResponse
	s.parseResponse(w, &response)
	assert.Equal(s.T(), "user not found", response.Error)
}

func (s *TestSuite) TestUserUpdate_DuplicateEmail() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	email1 := "dup1_" + time.Now().Format("20060102150405") + "@example.de"
	email2 := "dup2_" + time.Now().Format("20060102150405") + "@example.de"

	body1 := map[string]interface{}{
		"email": email1,
		"name":  "User One",
	}
	w1 := s.makeRequest("POST", "/api/v1/users", body1, s.AccessToken)
	var resp1 dto.UserResponse
	s.parseResponse(w1, &resp1)

	body2 := map[string]interface{}{
		"email": email2,
		"name":  "User Two",
	}
	s.makeRequest("POST", "/api/v1/users", body2, s.AccessToken)

	updateBody := map[string]interface{}{
		"email": email2,
	}
	w := s.makeRequest("PUT", "/api/v1/users/"+resp1.ID.String(), updateBody, s.AccessToken)

	assert.Equal(s.T(), http.StatusConflict, w.Code)

	var response dto.ErrorResponse
	s.parseResponse(w, &response)
	assert.Equal(s.T(), "email already exists", response.Error)
}

func (s *TestSuite) TestUserUpdate_Unauthorized() {
	updateBody := map[string]interface{}{
		"email": "test@example.de",
		"name":  "Test",
	}
	w := s.makeRequest("PUT", "/api/v1/users/00000000-0000-0000-0000-000000000000", updateBody, "")

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *TestSuite) TestUserDelete_Success() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	userEmail := "delete_" + time.Now().Format("20060102150405") + "@example.de"

	createUser := map[string]interface{}{
		"email": userEmail,
		"name":  "name",
	}
	wCreate := s.makeRequest("POST", "/api/v1/users", createUser, s.AccessToken)

	var createResponse dto.UserResponse
	s.parseResponse(wCreate, &createResponse)
	userID := createResponse.ID

	w := s.makeRequest("DELETE", "/api/v1/users/"+userID.String(), nil, s.AccessToken)
	assert.Equal(s.T(), http.StatusNoContent, w.Code)

	wGet := s.makeRequest("GET", "/api/v1/users/"+userID.String(), nil, s.AccessToken)
	assert.Equal(s.T(), http.StatusNotFound, wGet.Code)
}

func (s *TestSuite) TestUserDelete_NotFound() {
	email := "user_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "123456")

	w := s.makeRequest("DELETE", "/api/v1/users/00000000-0000-0000-0000-000000000000", nil, s.AccessToken)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)

	var response dto.ErrorResponse
	s.parseResponse(w, &response)
	assert.Equal(s.T(), "user not found", response.Error)
}

func (s *TestSuite) TestUserDelete_Unauthorized() {
	w := s.makeRequest("DELETE", "/api/v1/users/00000000-0000-0000-0000-000000000000", nil, "")

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *TestSuite) createAndLoginUser(email, password string) {
	registerBody := map[string]interface{}{
		"email":    email,
		"name":     "Test User",
		"password": password,
	}
	s.makeRequest("POST", "/api/v1/auth/register", registerBody, "")

	loginBody := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	w := s.makeRequest("POST", "/api/v1/auth/login", loginBody, "")

	var response dto.TokenResponse
	s.parseResponse(w, &response)
	s.AccessToken = response.AccessToken
	s.RefreshToken = response.RefreshToken
}
