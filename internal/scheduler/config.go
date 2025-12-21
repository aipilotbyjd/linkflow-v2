package scheduler

import "time"

type Config struct {
	// Polling
	PollInterval time.Duration
	BatchSize    int

	// Rate Limiting
	GlobalRateLimit   int // per minute
	WorkspaceLimit    int // per minute per workspace

	// Leader Election
	LeaderKey string
	LeaderTTL time.Duration

	// Recovery
	StaleThreshold  time.Duration
	CleanupInterval time.Duration
	RetentionDays   int

	// Shutdown
	ShutdownTimeout time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		PollInterval:      time.Second,
		BatchSize:         100,
		GlobalRateLimit:   1000,
		WorkspaceLimit:    100,
		LeaderKey:         "scheduler:leader",
		LeaderTTL:         30 * time.Second,
		StaleThreshold:    10 * time.Minute,
		CleanupInterval:   time.Hour,
		RetentionDays:     30,
		ShutdownTimeout:   30 * time.Second,
	}
}

func (c *Config) Validate() error {
	if c.PollInterval <= 0 {
		c.PollInterval = time.Second
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 100
	}
	if c.GlobalRateLimit <= 0 {
		c.GlobalRateLimit = 1000
	}
	if c.WorkspaceLimit <= 0 {
		c.WorkspaceLimit = 100
	}
	if c.LeaderTTL <= 0 {
		c.LeaderTTL = 30 * time.Second
	}
	if c.StaleThreshold <= 0 {
		c.StaleThreshold = 10 * time.Minute
	}
	if c.RetentionDays <= 0 {
		c.RetentionDays = 30
	}
	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = 30 * time.Second
	}
	return nil
}
