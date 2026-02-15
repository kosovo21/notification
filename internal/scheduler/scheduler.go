package scheduler

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"notification-system/internal/repository"
	"notification-system/internal/service"
)

// Scheduler polls for scheduled messages and publishes them when due.
type Scheduler struct {
	messageRepo repository.MessageRepository
	msgService  *service.MessageService
	interval    time.Duration
	batchSize   int
}

// NewScheduler creates a new Scheduler.
func NewScheduler(
	messageRepo repository.MessageRepository,
	msgService *service.MessageService,
	interval time.Duration,
	batchSize int,
) *Scheduler {
	if interval == 0 {
		interval = 10 * time.Second
	}
	if batchSize == 0 {
		batchSize = 50
	}
	return &Scheduler{
		messageRepo: messageRepo,
		msgService:  msgService,
		interval:    interval,
		batchSize:   batchSize,
	}
}

// Start begins the scheduler polling loop. Blocks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	log.Info().
		Dur("interval", s.interval).
		Int("batch_size", s.batchSize).
		Msg("scheduler started")

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("scheduler stopped")
			return
		case <-ticker.C:
			s.scan(ctx)
		}
	}
}

// scan finds due scheduled messages and publishes them.
func (s *Scheduler) scan(ctx context.Context) {
	messages, err := s.messageRepo.GetScheduledMessages(ctx, time.Now(), s.batchSize)
	if err != nil {
		log.Error().Err(err).Msg("scheduler: failed to get scheduled messages")
		return
	}

	if len(messages) == 0 {
		return
	}

	log.Info().Int("count", len(messages)).Msg("scheduler: found scheduled messages")

	for i := range messages {
		if err := s.msgService.PublishMessage(ctx, &messages[i]); err != nil {
			log.Error().Err(err).
				Str("message_id", messages[i].ID.String()).
				Msg("scheduler: failed to publish scheduled message")
		}
	}
}
