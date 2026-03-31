package service

import (
	"context"
	"log/slog"
	"time"

	"tracker/internal/metrics"
	"tracker/internal/repository"
)

type StatusUpdater struct {
	challenges *repository.ChallengeRepository
}

func NewStatusUpdater(c *repository.ChallengeRepository) *StatusUpdater {
	return &StatusUpdater{challenges: c}
}

// Start runs the status updater once immediately and then every hour.
func (u *StatusUpdater) Start(ctx context.Context) {
	u.run()
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				u.run()
			}
		}
	}()
}

func (u *StatusUpdater) run() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	activated, err := u.challenges.ActivateUpcoming(ctx)
	if err != nil {
		slog.Error("activate upcoming challenges", "error", err)
	} else if activated > 0 {
		slog.Info("activated upcoming challenges", "count", activated)
	}

	finished, err := u.challenges.FinishExpired(ctx)
	if err != nil {
		slog.Error("finish expired challenges", "error", err)
	} else if finished > 0 {
		slog.Info("finished expired challenges", "count", finished)
	}

	// Update active challenges gauge
	if count, err := u.challenges.CountActive(ctx); err == nil {
		metrics.ActiveChallengesTotal.Set(float64(count))
	}
}