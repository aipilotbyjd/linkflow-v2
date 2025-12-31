//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/linkflow-ai/linkflow/internal/api"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/linkflow-ai/linkflow/internal/pkg/oauth"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	appwire "github.com/linkflow-ai/linkflow/internal/wire"
	"gorm.io/gorm"
)

// App holds all initialized dependencies
type App struct {
	Config       *config.Config
	DB           *gorm.DB
	Redis        *pkgredis.Client
	Queue        *queue.Client
	JWTManager   *crypto.JWTManager
	OAuthManager *oauth.Manager
	Services     *api.Services
	Repos        *api.Repositories
}

// provideApp creates the App struct with all dependencies
func provideApp(
	cfg *config.Config,
	db *gorm.DB,
	redis *pkgredis.Client,
	queue *queue.Client,
	jwt *crypto.JWTManager,
	oauthMgr *oauth.Manager,
	services *api.Services,
	repos *api.Repositories,
) *App {
	return &App{
		Config:       cfg,
		DB:           db,
		Redis:        redis,
		Queue:        queue,
		JWTManager:   jwt,
		OAuthManager: oauthMgr,
		Services:     services,
		Repos:        repos,
	}
}

// InitializeApp wires all dependencies and returns the App
func InitializeApp() (*App, error) {
	wire.Build(
		appwire.AppSet,
		provideApp,
	)
	return nil, nil
}
