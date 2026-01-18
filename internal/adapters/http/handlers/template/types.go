package template

import (
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/template"
)

// Response DTOs

type TemplateResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Author      string    `json:"author"`
	IsFeatured  bool      `json:"featured"`
	UsageCount  int       `json:"usageCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TemplateDetailResponse struct {
	TemplateResponse
	Nodes       interface{} `json:"nodes"`
	Connections interface{} `json:"connections"`
	Settings    interface{} `json:"settings,omitempty"`
	Thumbnail   string      `json:"thumbnail,omitempty"`
}

type CategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon,omitempty"`
}

// Mappers

func ToTemplateResponse(t template.Template) TemplateResponse {
	description := ""
	if t.Description != nil {
		description = *t.Description
	}
	author := "LinkFlow"
	if t.Author != nil {
		author = *t.Author
	}
	tags := t.Tags
	if tags == nil {
		tags = []string{}
	}

	return TemplateResponse{
		ID:          t.ID.String(),
		Name:        t.Name,
		Description: description,
		Category:    t.Category,
		Tags:        tags,
		Author:      author,
		IsFeatured:  t.IsFeatured,
		UsageCount:  t.UsageCount,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func ToTemplateResponses(templates []template.Template) []TemplateResponse {
	responses := make([]TemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = ToTemplateResponse(t)
	}
	return responses
}

func ToTemplateDetailResponse(t *template.Template) TemplateDetailResponse {
	base := ToTemplateResponse(*t)
	thumbnail := ""
	if t.Thumbnail != nil {
		thumbnail = *t.Thumbnail
	}

	return TemplateDetailResponse{
		TemplateResponse: base,
		Nodes:            t.Nodes,
		Connections:      t.Connections,
		Settings:         t.Settings,
		Thumbnail:        thumbnail,
	}
}

func GetCategories() []CategoryResponse {
	return []CategoryResponse{
		{ID: "communication", Name: "Communication", Description: "Email, Slack, Teams integrations", Icon: "💬"},
		{ID: "devops", Name: "DevOps", Description: "CI/CD, GitHub, GitLab workflows", Icon: "🔧"},
		{ID: "crm", Name: "CRM", Description: "Salesforce, HubSpot integrations", Icon: "👥"},
		{ID: "data", Name: "Data", Description: "ETL, backup, sync workflows", Icon: "📊"},
		{ID: "marketing", Name: "Marketing", Description: "Marketing automation workflows", Icon: "📣"},
		{ID: "ecommerce", Name: "E-Commerce", Description: "Shopify, WooCommerce integrations", Icon: "🛒"},
	}
}
