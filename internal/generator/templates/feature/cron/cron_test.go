package worker

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewCronScheduler(t *testing.T) {
	s := NewCronScheduler()
	if s == nil {
		t.Fatal("Expected non-nil *CronScheduler")
	}
}

func TestCronScheduler_AddAndRun(t *testing.T) {
	s := NewCronScheduler()

	var count int64
	_, err := s.AddJob("@every 100ms", func() {
		atomic.AddInt64(&count, 1)
	})
	if err != nil {
		t.Fatalf("AddJob failed: %v", err)
	}

	s.Start()
	time.Sleep(350 * time.Millisecond)
	ctx := s.Stop()
	<-ctx.Done()

	if atomic.LoadInt64(&count) == 0 {
		t.Error("Expected cron job to run at least once")
	}
}

func TestCronScheduler_Entries(t *testing.T) {
	s := NewCronScheduler()
	s.AddJob("@every 1s", func() {}) //nolint:errcheck

	entries := s.Entries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}
