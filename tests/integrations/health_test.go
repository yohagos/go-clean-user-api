package integrations

import (
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
)

func (s *TestSuite) TestHealthCheck_Success() {
	email := "health_" + time.Now().Format("20060102150405") + "@example.de"
	s.createAndLoginUser(email, "Test123!@#")

	w := s.makeRequest("GET", "/api/v1/health", nil, s.AccessToken)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	s.parseResponse(w, &response)
	assert.Equal(s.T(), "healthy", response["status"])
	assert.Equal(s.T(), "connected", response["database"])
	assert.NotEmpty(s.T(), response["uptime"])
	assert.NotEmpty(s.T(), response["timestamp"])
}

func (s *TestSuite) TestHealthCheck_Unauthorized() {
	w := s.makeRequest("GET", "/api/v1/health", nil, "")
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *TestSuite) TestHealthCheck_InvalidToken() {
	w := s.makeRequest("GET", "/api/v1/health", nil, "invalid-token")
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}
