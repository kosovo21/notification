package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-system/internal/cache"
	"notification-system/internal/config"
	"notification-system/internal/queue"

	"notification-system/internal/repository"
	"notification-system/internal/router"
	"notification-system/internal/scheduler"
	"notification-system/internal/service"
	"notification-system/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	log := logger.Get()

	log.Info().Msg("starting notification API server")

	// Connect to database
	db, err := repository.NewDB(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("connected to database")

	// Connect to Redis
	rdb, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer rdb.Close()
	log.Info().Msg("connected to redis")

	// Connect to RabbitMQ
	rmq, err := queue.NewRabbitMQ(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}
	defer rmq.Close()
	log.Info().Msg("connected to rabbitmq")

	// Initialize publisher
	publisher := queue.NewPublisher(rmq)

	// Ensure exchange exists
	if err := rmq.DeclareExchange(queue.ExchangeName); err != nil {
		log.Fatal().Err(err).Msg("failed to declare exchange")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	recipientRepo := repository.NewRecipientRepository(db)

	// Initialize message service (for scheduler)
	msgService := service.NewMessageService(db, messageRepo, recipientRepo, publisher)

	// Build router
	r := router.NewRouter(router.Deps{
		DB:            db,
		UserRepo:      userRepo,
		MessageRepo:   messageRepo,
		RecipientRepo: recipientRepo,
		RedisClient:   rdb,
		RateLimit:     cfg.RateLimit,
		Publisher:     publisher,
	})

	// Start scheduler
	schedCtx, schedCancel := context.WithCancel(context.Background())
	defer schedCancel()
	sched := scheduler.NewScheduler(messageRepo, msgService, 10*time.Second, 50)
	go sched.Start(schedCtx)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("HTTP server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info().Str("signal", sig.String()).Msg("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server stopped")
}
