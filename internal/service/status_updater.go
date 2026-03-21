package service

import (
	"context"
	"log"
	"time"

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
		log.Printf("status_updater: activate upcoming: %v", err)
	} else if activated > 0 {
		log.Printf("status_updater: activated %d challenges", activated)
	}

	finished, err := u.challenges.FinishExpired(ctx)
	if err != nil {
		log.Printf("status_updater: finish expired: %v", err)
	} else if finished > 0 {
		log.Printf("status_updater: finished %d challenges", finished)
	}
}