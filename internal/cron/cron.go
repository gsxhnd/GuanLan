package cron

import (
	"context"
	"log"

	"github.com/gsxhnd/guanlan/internal/biz"
	"github.com/robfig/cron/v3"
)

// Scheduler wraps robfig/cron for api process jobs.
type Scheduler struct {
	cron *cron.Cron
	biz  *biz.Services
}

// New creates a cron scheduler. spec examples: "0 18 * * *" (daily 18:00).
func New(bizSvc *biz.Services, dailySyncSpec string) (*Scheduler, error) {
	if dailySyncSpec == "" {
		dailySyncSpec = "0 18 * * *"
	}
	c := cron.New()
	s := &Scheduler{cron: c, biz: bizSvc}
	_, err := c.AddFunc(dailySyncSpec, func() {
		ctx := context.Background()
		n, err := s.biz.Task.EnqueueDailySync(ctx)
		if err != nil {
			log.Printf("cron daily-sync enqueue: %v", err)
			return
		}
		log.Printf("cron daily-sync enqueued %d tasks", n)
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() context.Context {
	return s.cron.Stop()
}
