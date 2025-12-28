//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	appwire "github.com/linkflow-ai/linkflow/internal/wire"
	"gorm.io/gorm"
)

// SchedulerApp holds all initialized dependencies for scheduler
type SchedulerApp struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *pkgredis.Client
	Queue  *queue.Client
}

// provideSchedulerApp creates the SchedulerApp struct with all dependencies
func provideSchedulerApp(
	cfg *config.Config,
	db *gorm.DB,
	redis *pkgredis.Client,
	queue *queue.Client,
) *SchedulerApp {
	return &SchedulerApp{
		Config: cfg,
		DB:     db,
		Redis:  redis,
		Queue:  queue,
	}
}

// InitializeSchedulerApp wires all dependencies and returns the SchedulerApp
func InitializeSchedulerApp() (*SchedulerApp, error) {
	wire.Build(
		appwire.SchedulerSet,
		provideSchedulerApp,
	)
	return nil, nil
}
