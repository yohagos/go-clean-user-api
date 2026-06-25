package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yohagos/go-clean-user-api/internal/delivery/http/handler"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/middleware"
	"github.com/yohagos/go-clean-user-api/internal/domain/entity"
	"github.com/yohagos/go-clean-user-api/internal/domain/service"
	"github.com/yohagos/go-clean-user-api/internal/domain/usecase"
	"github.com/yohagos/go-clean-user-api/internal/logger"
	"github.com/yohagos/go-clean-user-api/internal/repository/postgres"
	"go.uber.org/zap"
)

type TestSuite struct {
	suite.Suite

	DB        *sqlx.DB
	Router    *gin.Engine
	Container testcontainers.Container

	AuthHandler *handler.AuthHandler
	UserHandler *handler.UserHandler

	AuthUseCase usecase.AuthUseCase
	UserUseCase usecase.UserUseCase

	TokenService service.TokenService

	TestUser entity.User

	AccessToken  string
	RefreshToken string
}

func (s *TestSuite) SetupSuite() {
	logger.Log = zap.NewNop()

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)
	s.Container = container

	dsn := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		dsn = "postgres://testuser:testpass@host.docker.internal:5432/testdb?sslmode=disable"
		db, err = sqlx.Connect("postgres", dsn)
		s.Require().NoError(err)
	}
	s.DB = db

	s.runMigrations()

	userRepo := postgres.NewUserRepository(db)
	tokenRepo := postgres.NewTokenRepository(db)

	tokenConfig := service.TokenConfig{
		AccessTokenSecret:  "test-access-secret",
		RefreshTokenSecret: "test-refresh-secret",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
	}

	s.TokenService = service.NewTokenService(tokenRepo, userRepo, tokenConfig)
	s.AuthUseCase = usecase.NewAuthUseCase(userRepo, tokenRepo, s.TokenService)
	s.UserUseCase = usecase.NewUserUseCase(userRepo)
	s.AuthHandler = handler.NewAuthHandler(s.AuthUseCase)
	s.UserHandler = handler.NewUserHandler(s.UserUseCase)

	s.Router = s.setupRouter()
}

func (s *TestSuite) TearDownSuite() {
	if s.Container != nil {
		s.Container.Terminate(context.Background())
	}

	if s.DB != nil {
		s.DB.Close()
	}
}

func (s *TestSuite) SetupTest() {
	s.clearDatabase()
	s.AccessToken = ""
	s.RefreshToken = ""

	s.Router = s.setupRouter()
}

func (s *TestSuite) runMigrations() {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(100) NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tokens (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token TEXT NOT NULL UNIQUE,
			type VARCHAR(10) NOT NULL CHECK (type IN ('access', 'refresh')),
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			revoked BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tokens_token ON tokens(token)`,
	}

	for _, migration := range migrations {
		_, err := s.DB.Exec(migration)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			s.Require().NoError(err)
		}
	}
}

func (s *TestSuite) clearDatabase() {
	s.DB.Exec("TRUNCATE TABLE tokens CASCADE")
	s.DB.Exec("TRUNCATE TABLE users CASCADE")
}

func (s *TestSuite) setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "router works"})
	})

	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.CORSMiddleware())

	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", s.AuthHandler.Register)
		auth.POST("/login", s.AuthHandler.Login)
		auth.POST("/refresh", s.AuthHandler.RefreshToken)
		auth.POST("/logout", middleware.AuthMiddleware(s.AuthUseCase), s.AuthHandler.Logout)
	}

	userGroup := router.Group("/api/v1/users")
	userGroup.Use(middleware.AuthMiddleware(s.AuthUseCase))
	{
		userGroup.POST("", s.UserHandler.CreateUser)
		userGroup.GET("", s.UserHandler.GetAllUsers)
		userGroup.GET("/:id", s.UserHandler.GetUser)
		userGroup.PUT("/:id", s.UserHandler.UpdateUser)
		userGroup.DELETE("/:id", s.UserHandler.DeleteUser)
	}

	return router
}

func (s *TestSuite) makeRequest(method, path string, body any, token string, refreshToken ...string) *httptest.ResponseRecorder {
	var jsonBody []byte
	if body != nil {
		jsonBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if len(refreshToken) > 0 {
		req.Header.Set("X-Refresh-Token", refreshToken[0])
	}

	w := httptest.NewRecorder()
	s.Router.ServeHTTP(w, req)

	return w
}

func (s *TestSuite) parseResponse(w *httptest.ResponseRecorder, target any) {
	json.Unmarshal(w.Body.Bytes(), target)
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestDatabaseConnection() {
	var result int
	err := s.DB.Get(&result, "SELECT 1")
	s.Require().NoError(err)
	assert.Equal(s.T(), 1, result)
}
