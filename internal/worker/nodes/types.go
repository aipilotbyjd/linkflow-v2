package nodes

// Re-export everything from core package
// This allows sub-packages to import core directly (no cycle)
// while consumers can still use nodes package

import (
	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// Type aliases for backward compatibility
type ExecutionContext = core.ExecutionContext
type Node = core.Node
type NodeMeta = core.NodeMeta
type Dependencies = core.Dependencies
type NodeWithDeps = core.NodeWithDeps
type Registry = core.Registry

// Re-export functions
var (
	Register              = core.Register
	SetGlobalDependencies = core.SetGlobalDependencies
	GetGlobalDependencies = core.GetGlobalDependencies
	Get                   = core.Get
	GetMeta               = core.GetMeta
	List                  = core.List
	ListByCategory        = core.ListByCategory
	ListAll               = core.ListAll
	Count                 = core.Count
)

// Re-export helper functions
var (
	GetString      = core.GetString
	GetInt         = core.GetInt
	GetFloat       = core.GetFloat
	GetBool        = core.GetBool
	GetMap         = core.GetMap
	GetArray       = core.GetArray
	GetStringArray = core.GetStringArray
	ToFloat        = core.ToFloat
	ToBool         = core.ToBool
	IsEmpty        = core.IsEmpty
	GetNestedValue = core.GetNestedValue
	ResolveValue   = core.ResolveValue
	FormatString   = core.FormatString
	MergeMap       = core.MergeMap
	CopyMap        = core.CopyMap
)
