package recovery

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Cleanup struct {
	db            *gorm.DB
	retentionDays int
	interval      time.Duration
}

func NewCleanup(db *gorm.DB, retentionDays int) *Cleanup {
	return &Cleanup{
		db:            db,
		retentionDays: retentionDays,
		interval:      time.Hour,
	}
}

func (c *Cleanup) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.cleanup(ctx)
		}
	}
}

func (c *Cleanup) cleanup(ctx context.Context) {
	cutoff := time.Now().AddDate(0, 0, -c.retentionDays)

	// Cleanup old executions
	result := c.db.WithContext(ctx).
		Exec("DELETE FROM executions WHERE created_at < ? AND status IN ('completed', 'failed', 'cancelled')", cutoff)

	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to cleanup old executions")
		return
	}

	if result.RowsAffected > 0 {
		log.Info().
			Int64("deleted", result.RowsAffected).
			Int("retention_days", c.retentionDays).
			Msg("Cleaned up old executions")
	}

	// Cleanup old node executions
	result = c.db.WithContext(ctx).
		Exec(`DELETE FROM node_executions WHERE execution_id NOT IN (SELECT id FROM executions)`)

	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to cleanup orphaned node executions")
		return
	}

	if result.RowsAffected > 0 {
		log.Info().Int64("deleted", result.RowsAffected).Msg("Cleaned up orphaned node executions")
	}
}

func (c *Cleanup) CleanupOnce(ctx context.Context) {
	c.cleanup(ctx)
}
