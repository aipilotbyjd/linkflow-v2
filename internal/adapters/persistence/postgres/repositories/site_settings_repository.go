package repositories

import (
	"context"

	"github.com/lib/pq"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/core/domain/sitesettings"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SiteSettingsRepository struct {
	db *gorm.DB
}

func NewSiteSettingsRepository(db *gorm.DB) *SiteSettingsRepository {
	return &SiteSettingsRepository{db: db}
}

// Get returns the site settings, creating default if not exists
func (r *SiteSettingsRepository) Get(ctx context.Context) (*sitesettings.SiteSettings, error) {
	var settings sitesettings.SiteSettings

	err := postgres.GetTx(ctx, r.db).First(&settings, "id = ?", 1).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default settings
			settings = *sitesettings.NewSiteSettings()
			if err := postgres.GetTx(ctx, r.db).Create(&settings).Error; err != nil {
				return nil, err
			}
			return &settings, nil
		}
		return nil, err
	}

	return &settings, nil
}

// Update updates the site settings
func (r *SiteSettingsRepository) Update(ctx context.Context, settings *sitesettings.SiteSettings) error {
	// Use upsert to ensure the record exists
	return postgres.GetTx(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"disabled_nodes", "updated_at"}),
	}).Create(settings).Error
}

// GetDisabledNodes returns just the disabled nodes list
func (r *SiteSettingsRepository) GetDisabledNodes(ctx context.Context) ([]string, error) {
	settings, err := r.Get(ctx)
	if err != nil {
		return nil, err
	}
	return settings.GetDisabledNodes(), nil
}

// SetDisabledNodes updates only the disabled nodes list
func (r *SiteSettingsRepository) SetDisabledNodes(ctx context.Context, nodes []string) error {
	settings, err := r.Get(ctx)
	if err != nil {
		return err
	}

	settings.SetDisabledNodes(nodes)

	return postgres.GetTx(ctx, r.db).Model(&sitesettings.SiteSettings{}).
		Where("id = ?", 1).
		Update("disabled_nodes", pq.StringArray(nodes)).Error
}
