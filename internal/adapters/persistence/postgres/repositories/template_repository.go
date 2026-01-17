package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/template"
	"gorm.io/gorm"
)

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) Create(ctx context.Context, t *template.Template) error {
	return postgres.GetTx(ctx, r.db).Create(t).Error
}

func (r *TemplateRepository) Update(ctx context.Context, t *template.Template) error {
	return postgres.GetTx(ctx, r.db).Save(t).Error
}

func (r *TemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&template.Template{}, "id = ?", id).Error
}

func (r *TemplateRepository) FindByID(ctx context.Context, id uuid.UUID) (*template.Template, error) {
	var t template.Template
	if err := postgres.GetTx(ctx, r.db).First(&t, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, template.ErrTemplateNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepository) FindAll(ctx context.Context, opts *template.ListOptions) ([]template.Template, int64, error) {
	var templates []template.Template
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&template.Template{})

	if opts != nil {
		if opts.IsPublic != nil {
			query = query.Where("is_public = ?", *opts.IsPublic)
		}
		if opts.IsFeatured != nil {
			query = query.Where("is_featured = ?", *opts.IsFeatured)
		}
		if opts.Category != "" {
			query = query.Where("category = ?", opts.Category)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil && opts.ListOptions != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	if err := query.Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *TemplateRepository) FindByCategory(ctx context.Context, category string, opts *template.ListOptions) ([]template.Template, int64, error) {
	var templates []template.Template
	var total int64

	query := postgres.GetTx(ctx, r.db).Model(&template.Template{}).Where("category = ?", category)

	if opts != nil && opts.IsPublic != nil {
		query = query.Where("is_public = ?", *opts.IsPublic)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil && opts.ListOptions != nil {
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	if err := query.Order("usage_count DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *TemplateRepository) FindFeatured(ctx context.Context, limit int) ([]template.Template, error) {
	var templates []template.Template
	if err := postgres.GetTx(ctx, r.db).
		Where("is_featured = ? AND is_public = ?", true, true).
		Order("usage_count DESC").
		Limit(limit).
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *TemplateRepository) Search(ctx context.Context, query string, opts *template.ListOptions) ([]template.Template, int64, error) {
	var templates []template.Template
	var total int64

	dbQuery := postgres.GetTx(ctx, r.db).Model(&template.Template{}).
		Where("is_public = ?", true).
		Where("name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")

	if opts != nil && opts.Category != "" {
		dbQuery = dbQuery.Where("category = ?", opts.Category)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if opts != nil && opts.ListOptions != nil {
		dbQuery = dbQuery.Offset(opts.Offset).Limit(opts.Limit)
	}

	if err := dbQuery.Order("usage_count DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *TemplateRepository) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Model(&template.Template{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).
		Error
}
