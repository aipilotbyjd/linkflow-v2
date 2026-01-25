package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	noteModel "github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models/note"
	domainNote "github.com/linkflow-ai/linkflow/internal/core/domain/note"
	"gorm.io/gorm"
)

type NoteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) Create(ctx context.Context, c *domainNote.Note) error {
	model := noteToModel(c)
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *NoteRepository) Update(ctx context.Context, c *domainNote.Note) error {
	model := noteToModel(c)
	return postgres.GetTx(ctx, r.db).Save(model).Error
}

func (r *NoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&noteModel.Note{}, "id = ?", id).Error
}

func (r *NoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainNote.Note, error) {
	var model noteModel.Note
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domainNote.ErrNotFound
		}
		return nil, err
	}
	return noteToDomain(&model), nil
}

func (r *NoteRepository) FindByWorkflow(ctx context.Context, workflowID uuid.UUID, opts *domainNote.ListOptions) ([]*domainNote.Note, int64, error) {
	var modelList []noteModel.Note
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&noteModel.Note{}).Where("workflow_id = ?", workflowID)

	// Apply filters
	if opts != nil {
		if opts.NodeID != nil {
			query = query.Where("node_id = ?", *opts.NodeID)
		}
		if opts.ResolvedOnly {
			query = query.Where("resolved = ?", true)
		}
		if opts.UnresolvedOnly {
			query = query.Where("resolved = ?", false)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil && opts.ListOptions != nil {
		query = query.Offset(opts.ListOptions.Offset).Limit(opts.ListOptions.Limit)
	}

	query = query.Order("created_at DESC")

	if err := query.Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	notes := make([]*domainNote.Note, len(modelList))
	for i, m := range modelList {
		notes[i] = noteToDomain(&m)
	}

	return notes, total, nil
}

func (r *NoteRepository) FindByNode(ctx context.Context, workflowID uuid.UUID, nodeID string) ([]*domainNote.Note, error) {
	var modelList []noteModel.Note

	query := postgres.GetTx(ctx, r.db).Model(&noteModel.Note{}).
		Where("workflow_id = ? AND node_id = ?", workflowID, nodeID).
		Order("created_at DESC")

	if err := query.Find(&modelList).Error; err != nil {
		return nil, err
	}

	notes := make([]*domainNote.Note, len(modelList))
	for i, m := range modelList {
		notes[i] = noteToDomain(&m)
	}

	return notes, nil
}

func (r *NoteRepository) CountByWorkflow(ctx context.Context, workflowID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&noteModel.Note{}).
		Where("workflow_id = ?", workflowID).
		Count(&count).Error
	return count, err
}

func (r *NoteRepository) CountUnresolvedByWorkflow(ctx context.Context, workflowID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&noteModel.Note{}).
		Where("workflow_id = ? AND resolved = ?", workflowID, false).
		Count(&count).Error
	return count, err
}

// Mapper functions
func noteToModel(c *domainNote.Note) *noteModel.Note {
	return &noteModel.Note{
		ID:          c.ID,
		WorkspaceID: c.WorkspaceID,
		WorkflowID:  c.WorkflowID,
		UserID:      c.UserID,
		NodeID:      c.NodeID,
		Content:     c.Content,
		Resolved:    c.Resolved,
		ResolvedAt:  c.ResolvedAt,
		ResolvedBy:  c.ResolvedBy,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func noteToDomain(m *noteModel.Note) *domainNote.Note {
	return &domainNote.Note{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		WorkflowID:  m.WorkflowID,
		UserID:      m.UserID,
		NodeID:      m.NodeID,
		Content:     m.Content,
		Resolved:    m.Resolved,
		ResolvedAt:  m.ResolvedAt,
		ResolvedBy:  m.ResolvedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
