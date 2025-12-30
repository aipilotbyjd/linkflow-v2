package wire

import (
	"github.com/google/wire"
)

// AppSet combines all provider sets for API server
var AppSet = wire.NewSet(
	InfraSet,
	CryptoSet,
	RepositorySet,
	ServiceSet,
	ProvideServices,
	ProvideRepositories,
)

// WorkerSet provides dependencies for worker service
var WorkerSet = wire.NewSet(
	InfraSet,
	CryptoSet,
	WorkerRepositorySet,
	WorkerServiceSet,
)

// SchedulerSet provides dependencies for scheduler service
var SchedulerSet = wire.NewSet(
	InfraSet,
)
