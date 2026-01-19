package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/mappers"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

type WorkflowRepository struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

func (r *WorkflowRepository) Create(ctx context.Context, wf *workflow.Workflow) error {
	model := mappers.WorkflowToModel(wf)
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *WorkflowRepository) Update(ctx context.Context, wf *workflow.Workflow) error {
	model := mappers.WorkflowToModel(wf)
	return postgres.GetTx(ctx, r.db).Save(model).Error
}

func (r *WorkflowRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&models.Workflow{}, "id = ?", id).Error
}

func (r *WorkflowRepository) FindByID(ctx context.Context, id uuid.UUID) (*workflow.Workflow, error) {
	var model models.Workflow
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, workflow.ErrWorkflowNotFound
		}
		return nil, err
	}
	return mappers.WorkflowToDomain(&model), nil
}

func (r *WorkflowRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *workflow.ListOptions) ([]workflow.Workflow, int64, error) {
	var modelList []models.Workflow
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).Where("workspace_id = ?", workspaceID)

	if opts != nil {
		if opts.Status != nil {
			query = query.Where("status = ?", *opts.Status)
		}
		if opts.Search != "" {
			query = query.Where("name ILIKE ?", "%"+opts.Search+"%")
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil && opts.ListOptions != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	query = query.Order("created_at DESC")

	if err := query.Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	workflows := make([]workflow.Workflow, len(modelList))
	for i, m := range modelList {
		workflows[i] = *mappers.WorkflowToDomain(&m)
	}

	return workflows, total, nil
}

func (r *WorkflowRepository) ExistsByName(ctx context.Context, workspaceID uuid.UUID, name string) (bool, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).
		Where("workspace_id = ? AND name = ?", workspaceID, name).
		Count(&count).Error
	return count > 0, err
}

func (r *WorkflowRepository) IncrementExecutionCount(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).
		Where("id = ?", id).
		UpdateColumn("execution_count", gorm.Expr("execution_count + 1")).
		UpdateColumn("last_executed_at", gorm.Expr("NOW()")).
		Error
}

func (r *WorkflowRepository) FindByFolderID(ctx context.Context, folderID uuid.UUID, opts *types.ListOptions) ([]workflow.Workflow, int64, error) {
	var modelList []models.Workflow
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).Where("project_id = ?", folderID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	query = query.Order("created_at DESC")

	if err := query.Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	workflows := make([]workflow.Workflow, len(modelList))
	for i, m := range modelList {
		workflows[i] = *mappers.WorkflowToDomain(&m)
	}

	return workflows, total, nil
}

func (r *WorkflowRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status workflow.Status) error {
	return postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *WorkflowRepository) CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).Where("workspace_id = ?", workspaceID).Count(&count).Error
	return count, err
}

func (r *WorkflowRepository) CountActiveByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).
		Where("workspace_id = ? AND status = ?", workspaceID, workflow.StatusActive).
		Count(&count).Error
	return count, err
}

// AdvancedSearch performs advanced search on workflows
func (r *WorkflowRepository) AdvancedSearch(ctx context.Context, workspaceID uuid.UUID, opts *workflow.AdvancedSearchOptions) ([]workflow.Workflow, int64, error) {
	var modelList []models.Workflow
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).Where("workspace_id = ?", workspaceID)

	// Text search
	if opts.Query != "" {
		searchQuery := "%" + opts.Query + "%"
		searchConditions := r.db.Where("1 = 0") // Start with false

		for _, field := range opts.SearchIn {
			switch field {
			case "name":
				searchConditions = searchConditions.Or("name ILIKE ?", searchQuery)
			case "description":
				searchConditions = searchConditions.Or("description ILIKE ?", searchQuery)
			case "tags":
				searchConditions = searchConditions.Or("? = ANY(tags)", opts.Query)
			case "nodes":
				// Search in nodes JSON array for node names or types
				searchConditions = searchConditions.Or("nodes::text ILIKE ?", searchQuery)
			}
		}
		query = query.Where(searchConditions)
	}

	// Status filter (multiple)
	if len(opts.Status) > 0 {
		query = query.Where("status IN ?", opts.Status)
	}

	// Tags filter (match any)
	if len(opts.Tags) > 0 {
		query = query.Where("tags && ?", "{"+joinStrings(opts.Tags, ",")+"}")
	}

	// Tags filter (match all)
	if len(opts.TagsAll) > 0 {
		query = query.Where("tags @> ?", "{"+joinStrings(opts.TagsAll, ",")+"}")
	}

	// Node types filter
	if len(opts.NodeTypes) > 0 {
		nodeConditions := r.db.Where("1 = 0")
		for _, nodeType := range opts.NodeTypes {
			// Search for node type in the nodes JSON array
			nodeConditions = nodeConditions.Or("nodes @> ?", `[{"type":"`+nodeType+`"}]`)
		}
		query = query.Where(nodeConditions)
	}

	// Category filter
	if opts.Category != "" {
		query = query.Where("category = ?", opts.Category)
	}

	// Favorite filter
	if opts.IsFavorite != nil {
		query = query.Where("is_favorite = ?", *opts.IsFavorite)
	}

	// Folder filter
	if opts.FolderID != nil {
		query = query.Where("project_id = ?", *opts.FolderID)
	}

	// Created by filter
	if opts.CreatedBy != nil {
		query = query.Where("created_by = ?", *opts.CreatedBy)
	}

	// Date filters
	if opts.CreatedAfter != nil {
		query = query.Where("created_at >= to_timestamp(?)", *opts.CreatedAfter)
	}
	if opts.CreatedBefore != nil {
		query = query.Where("created_at <= to_timestamp(?)", *opts.CreatedBefore)
	}
	if opts.UpdatedAfter != nil {
		query = query.Where("updated_at >= to_timestamp(?)", *opts.UpdatedAfter)
	}
	if opts.UpdatedBefore != nil {
		query = query.Where("updated_at <= to_timestamp(?)", *opts.UpdatedBefore)
	}
	if opts.ExecutedAfter != nil {
		query = query.Where("last_executed_at >= to_timestamp(?)", *opts.ExecutedAfter)
	}
	if opts.ExecutedBefore != nil {
		query = query.Where("last_executed_at <= to_timestamp(?)", *opts.ExecutedBefore)
	}

	// Execution count filters
	if opts.MinExecutions != nil {
		query = query.Where("execution_count >= ?", *opts.MinExecutions)
	}
	if opts.MaxExecutions != nil {
		query = query.Where("execution_count <= ?", *opts.MaxExecutions)
	}

	// Has error workflow filter
	if opts.HasErrors != nil {
		if *opts.HasErrors {
			query = query.Where("error_workflow_id IS NOT NULL")
		} else {
			query = query.Where("error_workflow_id IS NULL")
		}
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	sortBy := opts.SortBy
	if sortBy == "" {
		sortBy = "updated_at"
	}
	sortOrder := opts.SortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Validate sort column to prevent SQL injection
	validSortColumns := map[string]bool{
		"name": true, "created_at": true, "updated_at": true,
		"execution_count": true, "last_executed_at": true, "status": true,
	}
	if !validSortColumns[sortBy] {
		sortBy = "updated_at"
	}

	query = query.Order(sortBy + " " + sortOrder)

	// Pagination
	if opts.ListOptions != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	if err := query.Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	workflows := make([]workflow.Workflow, len(modelList))
	for i, m := range modelList {
		workflows[i] = *mappers.WorkflowToDomain(&m)
	}

	return workflows, total, nil
}

// GetNodeTypesInWorkspace returns all unique node types used in workflows
func (r *WorkflowRepository) GetNodeTypesInWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]string, error) {
	var results []string

	// Extract unique node types from all workflows in the workspace
	err := postgres.GetTx(ctx, r.db).Raw(`
		SELECT DISTINCT jsonb_array_elements(nodes)->>'type' as node_type
		FROM workflows
		WHERE workspace_id = ? AND deleted_at IS NULL
		AND jsonb_array_elements(nodes)->>'type' IS NOT NULL
		ORDER BY node_type
	`, workspaceID).Pluck("node_type", &results).Error

	return results, err
}

// GetTagsInWorkspace returns all unique tags used in workflows
func (r *WorkflowRepository) GetTagsInWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]string, error) {
	var results []string

	err := postgres.GetTx(ctx, r.db).Raw(`
		SELECT DISTINCT unnest(tags) as tag
		FROM workflows
		WHERE workspace_id = ? AND deleted_at IS NULL AND tags IS NOT NULL
		ORDER BY tag
	`, workspaceID).Pluck("tag", &results).Error

	return results, err
}

// GetCategoriesInWorkspace returns all unique categories used in workflows
func (r *WorkflowRepository) GetCategoriesInWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]string, error) {
	var results []string

	err := postgres.GetTx(ctx, r.db).Model(&models.Workflow{}).
		Where("workspace_id = ? AND category IS NOT NULL AND category != ''", workspaceID).
		Distinct("category").
		Order("category").
		Pluck("category", &results).Error

	return results, err
}

// Helper to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
