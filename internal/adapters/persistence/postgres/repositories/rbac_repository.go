package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/mappers"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
	"gorm.io/gorm"
)

type RBACRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) *RBACRepository {
	return &RBACRepository{db: db}
}

// Role Management
func (r *RBACRepository) CreateRole(ctx context.Context, role *rbac.Role) error {
	model := mappers.ToModelRole(role)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *RBACRepository) UpdateRole(ctx context.Context, role *rbac.Role) error {
	model := mappers.ToModelRole(role)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *RBACRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Role{}, id).Error
}

func (r *RBACRepository) GetRole(ctx context.Context, id uuid.UUID) (*rbac.Role, error) {
	var model models.Role
	// Preload permissions
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&model, id).Error; err != nil {
		return nil, err
	}
	return mappers.ToDomainRole(&model), nil
}

func (r *RBACRepository) GetRoleByName(ctx context.Context, workspaceID *uuid.UUID, name string) (*rbac.Role, error) {
	var model models.Role
	query := r.db.WithContext(ctx).Preload("Permissions").Where("name = ?", name)
	if workspaceID == nil {
		query = query.Where("workspace_id IS NULL")
	} else {
		query = query.Where("workspace_id = ?", workspaceID)
	}

	if err := query.First(&model).Error; err != nil {
		return nil, err
	}
	return mappers.ToDomainRole(&model), nil
}

func (r *RBACRepository) ListRoles(ctx context.Context, workspaceID uuid.UUID) ([]rbac.Role, error) {
	var models []models.Role
	// Fetch system roles (NULL workspace_id) AND workspace roles
	if err := r.db.WithContext(ctx).Preload("Permissions").
		Where("workspace_id = ? OR workspace_id IS NULL", workspaceID).
		Find(&models).Error; err != nil {
		return nil, err
	}
	roles := make([]rbac.Role, len(models))
	for i, m := range models {
		roles[i] = *mappers.ToDomainRole(&m)
	}
	return roles, nil
}

func (r *RBACRepository) ListSystemRoles(ctx context.Context) ([]rbac.Role, error) {
	var models []models.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").
		Where("workspace_id IS NULL").
		Find(&models).Error; err != nil {
		return nil, err
	}
	roles := make([]rbac.Role, len(models))
	for i, m := range models {
		roles[i] = *mappers.ToDomainRole(&m)
	}
	return roles, nil
}

// Permission Management
func (r *RBACRepository) ListPermissions(ctx context.Context) ([]rbac.Permission, error) {
	var models []models.Permission
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	perms := make([]rbac.Permission, len(models))
	for i, m := range models {
		perms[i] = mappers.ToDomainPermission(&m)
	}
	return perms, nil
}

func (r *RBACRepository) GetPermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]rbac.Permission, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, roleID).Error; err != nil {
		return nil, err
	}
	perms := make([]rbac.Permission, len(role.Permissions))
	for i, p := range role.Permissions {
		perms[i] = mappers.ToDomainPermission(&p)
	}
	return perms, nil
}

func (r *RBACRepository) AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []string) error {
	var role models.Role
	if err := r.db.WithContext(ctx).First(&role, roleID).Error; err != nil {
		return err
	}

	var perms []models.Permission
	if err := r.db.WithContext(ctx).Where("id IN ?", permissionIDs).Find(&perms).Error; err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(&role).Association("Permissions").Replace(perms)
}

func (r *RBACRepository) RemovePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []string) error {
	var role models.Role
	if err := r.db.WithContext(ctx).First(&role, roleID).Error; err != nil {
		return err
	}
	var perms []models.Permission
	if err := r.db.WithContext(ctx).Where("id IN ?", permissionIDs).Find(&perms).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&role).Association("Permissions").Delete(perms)
}
