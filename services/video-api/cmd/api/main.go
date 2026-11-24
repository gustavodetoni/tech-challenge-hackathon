package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/fiap-x/video-platform/video-api/internal/auth"
	"github.com/fiap-x/video-platform/video-api/internal/config"
	appdb "github.com/fiap-x/video-platform/video-api/internal/db"
	"github.com/fiap-x/video-platform/video-api/internal/middleware"
	"github.com/fiap-x/video-platform/video-api/internal/video"

	_ "github.com/fiap-x/video-platform/video-api/docs"
)

// @title FIAP Video API
// @version 0.1.0
// @description Foundation endpoints for the FIAP X video platform.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()
	databaseURL := cfg.DatabaseURL
	if databaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	if err := appdb.Migrate(db); err != nil { logger.Error("database migration failed", "error", err); os.Exit(1) }
	videoAPI, err := video.New(db, cfg)
	if err != nil { logger.Error("MinIO client initialization failed", "error", err); os.Exit(1) }

	router := newRouter(db, cfg, videoAPI)
	address := cfg.HTTPAddr
	server := &http.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("video API started", "address", address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("video API stopped")
}

func newRouter(db *gorm.DB, cfg config.Config, videos *video.API) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{AllowOrigins: config.AllowedOrigins(cfg.CORSOrigin), AllowMethods: []string{"GET", "POST", "OPTIONS"}, AllowHeaders: []string{"Authorization", "Content-Type"}, AllowCredentials: false, MaxAge: 12 * time.Hour}))
	router.GET("/health/live", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "live"}) })
	router.GET("/health/ready", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	api := router.Group("/api/v1")
	{
		// Ping godoc
		// @Summary Check that the API is available
		// @Tags foundation
		// @Produce json
		// @Success 200 {object} map[string]string
		// @Router /ping [get]
		api.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
	}
	protected := api.Group("/videos", middleware.Auth(auth.New(cfg.IssuerURL, cfg.Audience, cfg.JWKSURL)))
	protected.POST("/uploads", videos.Create)
	protected.POST("/:id/confirm", videos.Confirm)
	protected.GET("", videos.List)
	protected.GET("/:id", videos.Get)
	protected.GET("/:id/download", videos.Download)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" { return value }
	return fallback
}
