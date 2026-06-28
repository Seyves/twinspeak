package scheduler

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

type RenewalService interface {
	RenewSubscriptionsRoutine(ctx context.Context, now time.Time) error
}

type Scheduler struct {
	renewalService RenewalService
	interval       time.Duration
	ticker         *time.Ticker
	stopChan       chan struct{}
	doneChan       chan struct{}
	running        bool
}

func New(renewalService RenewalService, interval time.Duration) *Scheduler {
	return &Scheduler{
		renewalService: renewalService,
		interval:       interval,
		stopChan:       make(chan struct{}),
		doneChan:       make(chan struct{}),
		running:        false,
	}
}

func (s *Scheduler) Start(ctx context.Context, runImmediately bool) {
	if s.running {
		log.Warn("Scheduler already running, ignoring Start() call")
		return
	}

	s.running = true
	s.ticker = time.NewTicker(s.interval)

	log.Infof("Starting subscription renewal scheduler with interval: %s", s.interval)

	go func() {
		defer close(s.doneChan)
		defer s.recoverFromPanic()

		if runImmediately {
			log.Info("Running initial subscription renewal cycle")
			s.executeRenewalCycle(ctx)
		}

		for {
			select {
			case <-s.ticker.C:
				s.executeRenewalCycle(ctx)
			case <-s.stopChan:
				log.Info("Scheduler received stop signal, shutting down gracefully")
				s.ticker.Stop()
				return
			case <-ctx.Done():
				log.Info("Scheduler context cancelled, shutting down")
				s.ticker.Stop()
				return
			}
		}
	}()
}

func (s *Scheduler) Stop(timeout time.Duration) error {
	if !s.running {
		return nil
	}

	log.Info("Stopping scheduler...")
	close(s.stopChan)

	select {
	case <-s.doneChan:
		log.Info("Scheduler stopped successfully")
		s.running = false
		return nil
	case <-time.After(timeout):
		log.Warn("Scheduler stop timeout exceeded, forcing shutdown")
		s.running = false
		return fmt.Errorf("scheduler shutdown timeout after %s", timeout)
	}
}

func (s *Scheduler) executeRenewalCycle(ctx context.Context) {
	defer s.recoverFromPanic()

	startTime := time.Now()
	log.Info("Starting subscription renewal cycle")

	err := s.renewalService.RenewSubscriptionsRoutine(ctx, startTime)
	duration := time.Since(startTime)

	if err != nil {
		log.Errorf("Subscription renewal cycle failed after %s: %s", duration, err.Error())
	} else {
		log.Infof("Subscription renewal cycle completed successfully in %s", duration)
	}
}

func (s *Scheduler) recoverFromPanic() {
	if r := recover(); r != nil {
		log.Errorf("Scheduler panic recovered: %v\nStack trace:\n%s", r, string(debug.Stack()))
	}
}
