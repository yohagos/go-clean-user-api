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
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/yohagos/go-clean-user-api/internal/config"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/handler"
	"github.com/yohagos/go-clean-user-api/internal/delivery/http/middleware"
	"github.com/yohagos/go-clean-user-api/internal/domain/usecase"
	"github.com/yohagos/go-clean-user-api/internal/logger"
	"github.com/yohagos/go-clean-user-api/internal/repository/postgres"
	"go.uber.org/zap"
)

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
		"Starting User API Server",
		zap.String("mode", cfg.Server.Mode),
		zap.String("port", cfg.Server.Port),
	)

	db, err := initDB(cfg)
	if err != nil {
		logger.Log.Fatal("Failed to initialize database: ", zap.Error(err))
	}
	defer db.Close()

	userRepo := postgres.NewUserRepository(db)
	userUseCase := usecase.NewUserUseCase(userRepo)
	userHandler := handler.NewUserHandler(userUseCase)

	router := setupRouter(cfg, userHandler)

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

func setupRouter(cfg *config.Config, userHandler *handler.UserHandler) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimiterMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	v1 := router.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.GetAllUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	return router
}
