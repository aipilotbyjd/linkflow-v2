package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

// NodeExecutionModel represents node execution in database
type NodeExecutionModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutionID  uuid.UUID  `gorm:"type:uuid;index;not null"`
	NodeID       string     `gorm:"size:100;not null"`
	NodeType     string     `gorm:"size:100;not null"`
	Status       string     `gorm:"size:20;not null;default:pending"`
	InputData    types.JSON `gorm:"type:jsonb"`
	OutputData   types.JSON `gorm:"type:jsonb"`
	ErrorMessage *string    `gorm:"type:text"`
	StartedAt    *time.Time
	CompletedAt  *time.Time
	DurationMs   *int
	CreatedAt    time.Time
}

func (NodeExecutionModel) TableName() string {
	return "node_executions"
}

type NodeExecutionRepository struct {
	db *gorm.DB
}

func NewNodeExecutionRepository(db *gorm.DB) *NodeExecutionRepository {
	return &NodeExecutionRepository{db: db}
}

func (r *NodeExecutionRepository) Create(ctx context.Context, nodeExec *execution.NodeExecution) error {
	model := r.toModel(nodeExec)
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *NodeExecutionRepository) Update(ctx context.Context, nodeExec *execution.NodeExecution) error {
	model := r.toModel(nodeExec)
	return postgres.GetTx(ctx, r.db).Save(model).Error
}

func (r *NodeExecutionRepository) FindByID(ctx context.Context, id uuid.UUID) (*execution.NodeExecution, error) {
	var model NodeExecutionModel
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *NodeExecutionRepository) FindByExecutionID(ctx context.Context, executionID uuid.UUID) ([]execution.NodeExecution, error) {
	var models []NodeExecutionModel
	if err := postgres.GetTx(ctx, r.db).Where("execution_id = ?", executionID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	nodes := make([]execution.NodeExecution, len(models))
	for i, m := range models {
		nodes[i] = *r.toDomain(&m)
	}
	return nodes, nil
}

func (r *NodeExecutionRepository) FindByExecutionAndNodeID(ctx context.Context, executionID uuid.UUID, nodeID string) (*execution.NodeExecution, error) {
	var model NodeExecutionModel
	if err := postgres.GetTx(ctx, r.db).First(&model, "execution_id = ? AND node_id = ?", executionID, nodeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *NodeExecutionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status execution.NodeStatus) error {
	return postgres.GetTx(ctx, r.db).Model(&NodeExecutionModel{}).
		Where("id = ?", id).
		Update("status", string(status)).Error
}

func (r *NodeExecutionRepository) DeleteByExecutionID(ctx context.Context, executionID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&NodeExecutionModel{}, "execution_id = ?", executionID).Error
}

func (r *NodeExecutionRepository) toModel(n *execution.NodeExecution) *NodeExecutionModel {
	return &NodeExecutionModel{
		ID:           n.ID,
		ExecutionID:  n.ExecutionID,
		NodeID:       n.NodeID,
		NodeType:     n.NodeType,
		Status:       string(n.Status),
		InputData:    n.InputData,
		OutputData:   n.OutputData,
		ErrorMessage: n.ErrorMessage,
		StartedAt:    n.StartedAt,
		CompletedAt:  n.CompletedAt,
		DurationMs:   n.DurationMs,
		CreatedAt:    n.CreatedAt,
	}
}

func (r *NodeExecutionRepository) toDomain(m *NodeExecutionModel) *execution.NodeExecution {
	return &execution.NodeExecution{
		ID:           m.ID,
		ExecutionID:  m.ExecutionID,
		NodeID:       m.NodeID,
		NodeType:     m.NodeType,
		Status:       execution.NodeStatus(m.Status),
		InputData:    m.InputData,
		OutputData:   m.OutputData,
		ErrorMessage: m.ErrorMessage,
		StartedAt:    m.StartedAt,
		CompletedAt:  m.CompletedAt,
		DurationMs:   m.DurationMs,
		CreatedAt:    m.CreatedAt,
	}
}
