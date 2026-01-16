package template

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"gorm.io/gorm"
)

type Template struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	Category    string         `gorm:"size:50;index" json:"category"`
	Tags        []string       `gorm:"type:text[];serializer:json" json:"tags"`
	Nodes       types.JSONArray `gorm:"type:jsonb;not null" json:"nodes"`
	Connections types.JSONArray `gorm:"type:jsonb;not null" json:"connections"`
	Settings    types.JSON     `gorm:"type:jsonb" json:"settings,omitempty"`
	Thumbnail   *string        `gorm:"type:text" json:"thumbnail,omitempty"`
	Author      *string        `gorm:"size:100" json:"author,omitempty"`
	IsPublic    bool           `gorm:"default:true" json:"is_public"`
	IsFeatured  bool           `gorm:"default:false" json:"is_featured"`
	UsageCount  int            `gorm:"default:0" json:"usage_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Template) TableName() string {
	return "templates"
}

func NewTemplate(name, category string, nodes, connections types.JSONArray) *Template {
	now := time.Now()
	return &Template{
		ID:          uuid.New(),
		Name:        name,
		Category:    category,
		Nodes:       nodes,
		Connections: connections,
		Tags:        []string{},
		IsPublic:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (t *Template) IncrementUsage() {
	t.UsageCount++
	t.UpdatedAt = time.Now()
}
