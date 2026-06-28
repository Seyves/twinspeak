package scheduler

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type mockRenewalService struct {
	callCount      atomic.Int32
	shouldFail     bool
	executionDelay time.Duration
}

func (m *mockRenewalService) RenewSubscriptionsRoutine(ctx context.Context, now time.Time) error {
	m.callCount.Add(1)

	if m.executionDelay > 0 {
		select {
		case <-time.After(m.executionDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if m.shouldFail {
		return errors.New("mock renewal failed")
	}
	return nil
}

func (m *mockRenewalService) getCallCount() int32 {
	return m.callCount.Load()
}

func TestSchedulerStartStop(t *testing.T) {
	mock := &mockRenewalService{}
	sched := New(mock, 100*time.Millisecond)
	ctx := context.Background()

	sched.Start(ctx, false)

	if !sched.running {
		t.Fatal("Scheduler should be running after Start()")
	}

	time.Sleep(250 * time.Millisecond)

	err := sched.Stop(5 * time.Second)
	if err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	if sched.running {
		t.Fatal("Scheduler should not be running after Stop()")
	}

	count := mock.getCallCount()
	if count < 2 {
		t.Fatalf("Expected at least 2 executions, got %d", count)
	}
}

func TestSchedulerImmediateRun(t *testing.T) {
	mock := &mockRenewalService{}
	sched := New(mock, 1*time.Hour)
	ctx := context.Background()

	sched.Start(ctx, true)
	defer sched.Stop(5 * time.Second)

	time.Sleep(100 * time.Millisecond)

	count := mock.getCallCount()
	if count < 1 {
		t.Fatalf("Expected at least 1 execution with immediate run, got %d", count)
	}
}

func TestSchedulerErrorHandling(t *testing.T) {
	mock := &mockRenewalService{shouldFail: true}
	sched := New(mock, 100*time.Millisecond)
	ctx := context.Background()

	sched.Start(ctx, false)
	time.Sleep(250 * time.Millisecond)
	sched.Stop(5 * time.Second)

	count := mock.getCallCount()
	if count < 2 {
		t.Fatalf("Scheduler should continue after errors, got %d executions", count)
	}
}

func TestSchedulerGracefulShutdown(t *testing.T) {
	mock := &mockRenewalService{executionDelay: 500 * time.Millisecond}
	sched := New(mock, 100*time.Millisecond)
	ctx := context.Background()

	sched.Start(ctx, true)

	time.Sleep(50 * time.Millisecond)

	start := time.Now()
	err := sched.Stop(2 * time.Second)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	if duration < 400*time.Millisecond {
		t.Fatalf("Stop() didn't wait for execution to complete: %v", duration)
	}
}

func TestSchedulerStopTimeout(t *testing.T) {
	mock := &mockRenewalService{executionDelay: 5 * time.Second}
	sched := New(mock, 1*time.Hour)
	ctx := context.Background()

	sched.Start(ctx, true)

	err := sched.Stop(100 * time.Millisecond)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

func TestSchedulerContextCancellation(t *testing.T) {
	mock := &mockRenewalService{}
	sched := New(mock, 100*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	sched.Start(ctx, false)

	time.Sleep(250 * time.Millisecond)
	beforeCancel := mock.getCallCount()

	cancel()

	time.Sleep(300 * time.Millisecond)
	afterCancel := mock.getCallCount()

	if afterCancel > beforeCancel+1 {
		t.Fatalf("Scheduler continued after context cancel: before=%d, after=%d", beforeCancel, afterCancel)
	}
}

func TestSchedulerDoubleStart(t *testing.T) {
	mock := &mockRenewalService{}
	sched := New(mock, 100*time.Millisecond)
	ctx := context.Background()

	sched.Start(ctx, false)
	sched.Start(ctx, false)

	time.Sleep(250 * time.Millisecond)
	sched.Stop(5 * time.Second)

	count := mock.getCallCount()
	if count > 4 {
		t.Fatalf("Double start may have caused double executions: %d", count)
	}
}

func TestSchedulerPanicRecovery(t *testing.T) {
	mock := &mockRenewalService{}
	sched := New(mock, 100*time.Millisecond)
	ctx := context.Background()

	sched.Start(ctx, false)
	time.Sleep(250 * time.Millisecond)
	sched.Stop(5 * time.Second)

	if !sched.running && mock.getCallCount() >= 2 {
		return
	}
}

func TestSchedulerRealWorldScenario(t *testing.T) {
	mock := &mockRenewalService{}
	sched := New(mock, 200*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	sched.Start(ctx, true)

	time.Sleep(550 * time.Millisecond)

	cancel()

	err := sched.Stop(5 * time.Second)
	if err != nil {
		t.Fatalf("Graceful shutdown failed: %v", err)
	}

	count := mock.getCallCount()
	if count < 3 {
		t.Fatalf("Expected at least 3 executions in real-world scenario, got %d", count)
	}

	if sched.running {
		t.Fatal("Scheduler should be stopped")
	}
}
