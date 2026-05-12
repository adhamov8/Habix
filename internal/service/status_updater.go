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
	badgeSvc   *BadgeService
}

func NewStatusUpdater(c *repository.ChallengeRepository) *StatusUpdater {
	return &StatusUpdater{challenges: c}
}

// подключаем выдачу значков при автоматическом завершении
func (u *StatusUpdater) SetBadgeService(bs *BadgeService) {
	u.badgeSvc = bs
}

// запускаем обновление статусов сразу и потом каждый час
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

	finishedIDs, err := u.challenges.FinishExpired(ctx)
	if err != nil {
		slog.Error("finish expired challenges", "error", err)
	} else if len(finishedIDs) > 0 {
		slog.Info("finished expired challenges", "count", len(finishedIDs))
		if u.badgeSvc != nil {
			for _, id := range finishedIDs {
				u.badgeSvc.ProcessFinishedChallenge(ctx, id)
			}
		}
	}

	if count, err := u.challenges.CountActive(ctx); err == nil {
		metrics.ActiveChallengesTotal.Set(float64(count))
	}
}
