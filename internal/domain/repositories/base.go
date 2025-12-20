package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListOptions struct {
	Offset  int
	Limit   int
	OrderBy string
	Order   string // asc or desc
}

func NewListOptions(page, perPage int) *ListOptions {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return &ListOptions{
		Offset:  (page - 1) * perPage,
		Limit:   perPage,
		OrderBy: "created_at",
		Order:   "desc",
	}
}

type BaseRepository[T any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) DB() *gorm.DB {
	return r.db
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, "id = ?", id).Error
}

func (r *BaseRepository[T]) FindByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepository[T]) FindAll(ctx context.Context, opts *ListOptions) ([]T, int64, error) {
	var entities []T
	var total int64

	query := r.db.WithContext(ctx).Model(new(T))
	query.Count(&total)

	if opts != nil {
		if opts.OrderBy != "" {
			query = query.Order(opts.OrderBy + " " + opts.Order)
		}
		query = query.Offset(opts.Offset).Limit(opts.Limit)
	}

	err := query.Find(&entities).Error
	return entities, total, err
}

func (r *BaseRepository[T]) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *BaseRepository[T]) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
