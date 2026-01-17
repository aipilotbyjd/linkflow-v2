package template

import "time"

// Template represents a workflow template
type Template struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Author      string    `json:"author"`
	Version     string    `json:"version"`
	Downloads   int       `json:"downloads"`
	Rating      float64   `json:"rating"`
	RatingCount int       `json:"ratingCount"`
	Featured    bool      `json:"featured"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TemplateCategory represents a template category
type TemplateCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Count       int    `json:"count"`
	Icon        string `json:"icon,omitempty"`
}

// TemplateRepository defines the template repository interface
type TemplateRepository interface {
	FindAll(limit, offset int) ([]Template, int64, error)
	FindByID(id string) (*Template, error)
	FindFeatured(limit int) ([]Template, error)
	FindByCategory(category string, limit, offset int) ([]Template, int64, error)
	Search(query string, limit, offset int) ([]Template, int64, error)
	GetCategories() ([]TemplateCategory, error)
}

// GetStaticTemplates returns sample templates for demo purposes
func GetStaticTemplates() []Template {
	return []Template{
		{
			ID:          "tpl-1",
			Name:        "Slack to Email Notifier",
			Description: "Forward important Slack messages to email",
			Category:    "communication",
			Tags:        []string{"slack", "email", "notifications"},
			Author:      "LinkFlow",
			Version:     "1.0.0",
			Downloads:   1250,
			Rating:      4.8,
			RatingCount: 45,
			Featured:    true,
			CreatedAt:   time.Now().AddDate(0, -3, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -7),
		},
		{
			ID:          "tpl-2",
			Name:        "GitHub to Slack Reporter",
			Description: "Get Slack notifications for GitHub events",
			Category:    "devops",
			Tags:        []string{"github", "slack", "ci-cd"},
			Author:      "LinkFlow",
			Version:     "1.2.0",
			Downloads:   980,
			Rating:      4.6,
			RatingCount: 32,
			Featured:    true,
			CreatedAt:   time.Now().AddDate(0, -6, 0),
			UpdatedAt:   time.Now().AddDate(0, -1, 0),
		},
		{
			ID:          "tpl-3",
			Name:        "CRM Lead Sync",
			Description: "Sync leads between Salesforce and HubSpot",
			Category:    "crm",
			Tags:        []string{"salesforce", "hubspot", "leads"},
			Author:      "LinkFlow",
			Version:     "2.0.0",
			Downloads:   750,
			Rating:      4.5,
			RatingCount: 28,
			Featured:    true,
			CreatedAt:   time.Now().AddDate(0, -4, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -14),
		},
		{
			ID:          "tpl-4",
			Name:        "Data Backup to S3",
			Description: "Automatically backup data to AWS S3",
			Category:    "data",
			Tags:        []string{"aws", "s3", "backup"},
			Author:      "LinkFlow",
			Version:     "1.1.0",
			Downloads:   620,
			Rating:      4.7,
			RatingCount: 21,
			Featured:    false,
			CreatedAt:   time.Now().AddDate(0, -2, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -3),
		},
	}
}
