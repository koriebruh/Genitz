package worker

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"
)

// CronScheduler wraps cron.Cron with a convenient API.
type CronScheduler struct {
	c *cron.Cron
}

// NewCronScheduler creates a cron.Cron scheduler with seconds precision.
func NewCronScheduler() *CronScheduler {
	c := cron.New(cron.WithSeconds())
	log.Println("✅ Cron scheduler initialized")
	return &CronScheduler{c: c}
}

// AddJob registers a named function to run on the given cron expression.
// Example spec: "0 * * * * *" (every minute at :00 seconds)
func (s *CronScheduler) AddJob(spec string, fn func()) (cron.EntryID, error) {
	return s.c.AddFunc(spec, fn)
}

// Start begins executing registered cron jobs in the background.
func (s *CronScheduler) Start() {
	s.c.Start()
}

// Stop gracefully shuts down the scheduler and waits for running jobs to finish.
func (s *CronScheduler) Stop() context.Context {
	return s.c.Stop()
}

// Entries returns all registered cron entries.
func (s *CronScheduler) Entries() []cron.Entry {
	return s.c.Entries()
}
