package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"notification-system/internal/adapter"
	"notification-system/internal/config"
	"notification-system/internal/queue"
	"notification-system/internal/repository"
	"notification-system/internal/worker"
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

	log.Info().Msg("starting notification worker")

	// Connect to database
	db, err := repository.NewDB(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("connected to database")

	// Connect to RabbitMQ
	rmq, err := queue.NewRabbitMQ(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}
	defer rmq.Close()
	log.Info().Msg("connected to rabbitmq")

	// Set prefetch count
	if err := rmq.Channel().Qos(cfg.RabbitMQ.PrefetchCount, 0, false); err != nil {
		log.Fatal().Err(err).Msg("failed to set QoS")
	}

	// Ensure exchange exists
	if err := rmq.DeclareExchange(queue.ExchangeName); err != nil {
		log.Fatal().Err(err).Msg("failed to declare exchange")
	}

	// Initialize repositories
	recipientRepo := repository.NewRecipientRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Load platform credentials
	twilioCfg, sendgridCfg, _, _ := config.LoadPlatformCredentials()

	// Initialize adapters
	// Use mock adapters for platforms without real credentials configured
	adapters := make(map[string]adapter.Sender)

	if twilioCfg.AccountSID != "" {
		adapters["sms"] = adapter.NewTwilioAdapter(twilioCfg)
		log.Info().Msg("using Twilio adapter for SMS")
	} else {
		adapters["sms"] = adapter.NewMockAdapter("sms")
		log.Info().Msg("using Mock adapter for SMS")
	}

	if sendgridCfg.APIKey != "" {
		adapters["email"] = adapter.NewSendGridAdapter(sendgridCfg)
		log.Info().Msg("using SendGrid adapter for email")
	} else {
		adapters["email"] = adapter.NewMockAdapter("email")
		log.Info().Msg("using Mock adapter for email")
	}

	// WhatsApp and Telegram always use mock for now
	adapters["whatsapp"] = adapter.NewMockAdapter("whatsapp")
	adapters["telegram"] = adapter.NewMockAdapter("telegram")
	log.Info().Msg("using Mock adapters for whatsapp and telegram")

	// Create consumer and worker
	consumer := queue.NewConsumer(rmq)
	w := worker.NewWorker(consumer, recipientRepo, messageRepo, adapters)

	// Context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Define queues to consume from
	queues := []struct {
		name       string
		routingKey string
	}{
		{name: "notification.sms", routingKey: queue.RoutingKeySMS},
		{name: "notification.email", routingKey: queue.RoutingKeyEmail},
		{name: "notification.whatsapp", routingKey: queue.RoutingKeyWhatsApp},
		{name: "notification.telegram", routingKey: queue.RoutingKeyTelegram},
	}

	// Start workers for each queue
	var wg sync.WaitGroup
	for _, q := range queues {
		wg.Add(1)
		go func(queueName, routingKey string) {
			defer wg.Done()
			if err := w.Start(ctx, queueName, routingKey); err != nil {
				log.Error().Err(err).Str("queue", queueName).Msg("worker stopped with error")
			}
		}(q.name, q.routingKey)
	}

	log.Info().Int("queues", len(queues)).Msg("all workers started")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info().Str("signal", sig.String()).Msg("shutting down worker")
	cancel()
	wg.Wait()
	log.Info().Msg("worker stopped")
}
