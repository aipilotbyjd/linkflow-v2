package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

// VersionModel represents workflow version in database
type VersionModel struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WorkflowID  uuid.UUID       `gorm:"type:uuid;index;not null"`
	Version     int             `gorm:"not null"`
	Nodes       types.JSONArray `gorm:"type:jsonb"`
	Connections types.JSONArray `gorm:"type:jsonb"`
	Settings    types.JSON      `gorm:"type:jsonb"`
	CreatedBy   *uuid.UUID      `gorm:"type:uuid"`
	CreatedAt   time.Time
}

func (VersionModel) TableName() string {
	return "workflow_versions"
}

type VersionRepository struct {
	db *gorm.DB
}

func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

func (r *VersionRepository) Create(ctx context.Context, version *workflow.Version) error {
	model := &VersionModel{
		ID:          version.ID,
		WorkflowID:  version.WorkflowID,
		Version:     version.Version,
		Nodes:       version.Nodes,
		Connections: version.Connections,
		Settings:    version.Settings,
		CreatedBy:   version.CreatedBy,
		CreatedAt:   version.CreatedAt,
	}
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *VersionRepository) FindByID(ctx context.Context, id uuid.UUID) (*workflow.Version, error) {
	var model VersionModel
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, workflow.ErrVersionNotFound
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *VersionRepository) FindByWorkflowID(ctx context.Context, workflowID uuid.UUID, opts *types.ListOptions) ([]workflow.Version, int64, error) {
	var models []VersionModel
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&VersionModel{}).Where("workflow_id = ?", workflowID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("version DESC")
	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, 0, err
	}

	versions := make([]workflow.Version, len(models))
	for i, m := range models {
		versions[i] = *r.toDomain(&m)
	}
	return versions, total, nil
}

func (r *VersionRepository) FindByWorkflowAndVersion(ctx context.Context, workflowID uuid.UUID, version int) (*workflow.Version, error) {
	var model VersionModel
	if err := postgres.GetTx(ctx, r.db).First(&model, "workflow_id = ? AND version = ?", workflowID, version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, workflow.ErrVersionNotFound
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *VersionRepository) FindLatestByWorkflowID(ctx context.Context, workflowID uuid.UUID) (*workflow.Version, error) {
	var model VersionModel
	if err := postgres.GetTx(ctx, r.db).Where("workflow_id = ?", workflowID).Order("version DESC").First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, workflow.ErrVersionNotFound
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *VersionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&VersionModel{}, "id = ?", id).Error
}

func (r *VersionRepository) DeleteByWorkflowID(ctx context.Context, workflowID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&VersionModel{}, "workflow_id = ?", workflowID).Error
}

func (r *VersionRepository) toDomain(m *VersionModel) *workflow.Version {
	return &workflow.Version{
		ID:          m.ID,
		WorkflowID:  m.WorkflowID,
		Version:     m.Version,
		Nodes:       types.JSONArray(m.Nodes),
		Connections: types.JSONArray(m.Connections),
		Settings:    m.Settings,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
	}
}
