package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/analytics"
	"github.com/mohdrashid9678/xlink/internal/auth"
	"github.com/mohdrashid9678/xlink/internal/cache"
	"github.com/mohdrashid9678/xlink/internal/database"
	"github.com/mohdrashid9678/xlink/internal/handlers"
	"github.com/mohdrashid9678/xlink/internal/middleware"
	"github.com/mohdrashid9678/xlink/internal/repository"
	"github.com/mohdrashid9678/xlink/internal/routes"
	"github.com/mohdrashid9678/xlink/internal/server"
	"github.com/mohdrashid9678/xlink/internal/service"
	"github.com/mohdrashid9678/xlink/pkg/config"
	"github.com/mohdrashid9678/xlink/pkg/logger"
	"github.com/mohdrashid9678/xlink/pkg/metrics"
	"github.com/mohdrashid9678/xlink/pkg/tracer"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatalf("application terminated with error: %v", err)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	appLogger := logger.New(logger.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	})

	// Initialize OpenTelemetry Distributed Tracer
	_, shutdownTracer, err := tracer.InitTracer(ctx, tracer.Config{
		ServiceName:   cfg.ServiceName,
		Endpoint:      cfg.OTelEndpoint,
		SamplingRatio: cfg.OTelSamplingRatio,
	})
	if err != nil {
		appLogger.Warn("failed to initialize OpenTelemetry tracer", slog.Any("error", err))
	} else {
		defer func() {
			if err := shutdownTracer(context.Background()); err != nil {
				appLogger.Warn("error shutting down tracer", slog.Any("error", err))
			}
		}()
	}

	if cfg.AutoMigrate {
		if err := database.RunMigrations(cfg.DBURL); err != nil {
			return fmt.Errorf("database migrations failed: %w", err)
		}
	}

	db, err := database.NewPostgres(cfg.DBURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	urlCache, redisClient, closeCache := cache.NewURLCacheStack(ctx, cfg.RedisURL, appLogger)
	defer closeCache()

	urlRepository := repository.NewPostgresURLRepository(db.Pool)
	userRepository := repository.NewPostgresUserRepository(db.Pool)
	refreshTokenRepository := repository.NewPostgresRefreshTokenRepository(db.Pool)
	analyticsRepository := repository.NewPostgresAnalyticsRepository(db.Pool)

	// Initialize Analytics Pipeline
	extractor := analytics.NewExtractor()
	var producer analytics.StreamProducer = analytics.NewNoopProducer()
	if redisClient != nil {
		rawRedis := redisClient.Raw()
		producer = analytics.NewRedisStreamProducer(rawRedis)
		consumer := analytics.NewStreamConsumer(rawRedis, analyticsRepository, appLogger, analytics.ConsumerConfig{})
		if err := consumer.Start(ctx); err != nil {
			appLogger.Warn("failed to start analytics stream consumer", slog.Any("error", err))
		} else {
			defer consumer.Stop()
		}
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSigningKey)
	urlService := service.NewURLService(urlRepository, urlCache)
	authService := service.NewAuthService(userRepository, refreshTokenRepository, jwtManager)
	analyticsService := service.NewAnalyticsService(urlRepository, analyticsRepository)

	urlHandler := handlers.NewURLHandler(urlService, extractor, producer)
	authHandler := handlers.NewAuthHandler(authService)
	healthHandler := handlers.NewHealthHandler(db.Pool, redisClient)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(
		middleware.OpenTelemetryTracing(cfg.ServiceName),
		middleware.TraceResponseHeader(),
		middleware.RequestID(),
		middleware.StructuredLogger(appLogger),
		middleware.PrometheusMetrics(metrics.DefaultMetrics),
		middleware.Recovery(appLogger),
	)

	routes.RegisterRoutes(router, urlHandler, authHandler, healthHandler, analyticsHandler, middleware.RequireAuth(jwtManager))
	if cfg.PProfEnabled {
		routes.RegisterPProfRoutes(router)
	}

	srv := server.New(router, server.Config{
		Port: cfg.Port,
	})

	return srv.Serve(ctx)
}
