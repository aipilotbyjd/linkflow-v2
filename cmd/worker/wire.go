//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	appwire "github.com/linkflow-ai/linkflow/internal/wire"
	"gorm.io/gorm"
)

// WorkerApp holds all initialized dependencies for worker
type WorkerApp struct {
	Config        *config.Config
	DB            *gorm.DB
	Redis         *pkgredis.Client
	Queue         *queue.Client
	Encryptor     *crypto.Encryptor
	WorkflowSvc   *services.WorkflowService
	ExecutionSvc  *services.ExecutionService
	CredentialSvc *services.CredentialService
	BillingSvc    *services.BillingService
	OAuthSvc      *services.OAuthService
}

// provideWorkerApp creates the WorkerApp struct with all dependencies
func provideWorkerApp(
	cfg *config.Config,
	db *gorm.DB,
	redis *pkgredis.Client,
	queue *queue.Client,
	encryptor *crypto.Encryptor,
	workflowSvc *services.WorkflowService,
	executionSvc *services.ExecutionService,
	credentialSvc *services.CredentialService,
	billingSvc *services.BillingService,
	oauthSvc *services.OAuthService,
) *WorkerApp {
	return &WorkerApp{
		Config:        cfg,
		DB:            db,
		Redis:         redis,
		Queue:         queue,
		Encryptor:     encryptor,
		WorkflowSvc:   workflowSvc,
		ExecutionSvc:  executionSvc,
		CredentialSvc: credentialSvc,
		BillingSvc:    billingSvc,
		OAuthSvc:      oauthSvc,
	}
}

// InitializeWorkerApp wires all dependencies and returns the WorkerApp
func InitializeWorkerApp() (*WorkerApp, error) {
	wire.Build(
		appwire.WorkerSet,
		provideWorkerApp,
	)
	return nil, nil
}
