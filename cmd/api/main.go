package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/yohagos/go-clean-user-api/docs"

	"github.com/yohagos/go-clean-user-api/internal/config"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/handler"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/middleware"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/validators"
	"github.com/yohagos/go-clean-user-api/internal/domain/service"
	"github.com/yohagos/go-clean-user-api/internal/domain/usecase"
	"github.com/yohagos/go-clean-user-api/internal/logger"
	"github.com/yohagos/go-clean-user-api/internal/repository/postgres"
)

// @title User Management API
// @version 1.0
// @description This is a production-ready User Management API with JWT authentication
// @termsOfService https://example.com/terms/

// @contact.name API Support
// @contact.url https://example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	if err := logger.Init(cfg.Log.Level, cfg.Log.OutputPath, cfg.Log.Encoding); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Log.Info(
		"Starting API Server....",
		zap.Any("Server Configs", cfg.Server),
	)

	db, err := initDB(cfg)
	if err != nil {
		logger.Log.Fatal("Failed to initialize database: ", zap.Error(err))
	}
	defer db.Close()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validators.RegisterValidators(v)
		logger.Log.Info("Custom Validators registered successfully!!!")
	}

	userRepo := postgres.NewUserRepository(db)
	tokenRepo := postgres.NewTokenRepository(db)

	tokenConfig := service.TokenConfig{
		AccessTokenSecret:  cfg.JWT.AccessSecret,
		RefreshTokenSecret: cfg.JWT.RefreshSecret,
		AccessTokenTTL:     time.Duration(cfg.JWT.AccessTTL) * time.Second,
		RefreshTokenTTL:    time.Duration(cfg.JWT.RefreshTTL) * time.Second,
	}

	tokenService := service.NewTokenService(tokenRepo, userRepo, tokenConfig)
	authUseCase := usecase.NewAuthUseCase(userRepo, tokenRepo, tokenService)
	userUseCase := usecase.NewUserUseCase(userRepo)

	authHandler := handler.NewAuthHandler(authUseCase)
	userHandler := handler.NewUserHandler(userUseCase)
	healthHandler := handler.NewHealthHandler(db)

	router := setupRouter(cfg, authHandler, userHandler, healthHandler, authUseCase)

	srv := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	go func() {
		logger.Log.Info("Starting server on :8080", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server.", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("server forced to shutdown: ", zap.Error(err))
	}

	logger.Log.Info("Server exited properly")
}

func initDB(cfg *config.Config) (*sqlx.DB, error) {
	logger.Log.Info("Initializing database connection")

	db, err := sqlx.Open("postgres", cfg.Database.URL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Log.Fatal("Error: Pinging DB caused an error.", zap.Error(err))
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	logger.Log.Info("Database connected successfully")
	return db, nil
}

func setupRouter(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	healthHandler *handler.HealthHandler,
	authUseCase usecase.AuthUseCase,
) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimiterMiddleware())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", middleware.AuthMiddleware(authUseCase), authHandler.Logout)
	}

	api := router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(authUseCase))
	{
		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.GetAllUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	api.Use(middleware.AuthMiddleware(authUseCase))
	{
		health := api.Group("/health")
		{
			health.GET("", healthHandler.Check)
		}
	}

	return router
}
