package marketplace

import "time"

// MarketplaceTemplate represents a marketplace template
type MarketplaceTemplate struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	LongDescription string    `json:"longDescription,omitempty"`
	Category        string    `json:"category"`
	Tags            []string  `json:"tags"`
	Author          Author    `json:"author"`
	Version         string    `json:"version"`
	Downloads       int       `json:"downloads"`
	Rating          float64   `json:"rating"`
	RatingCount     int       `json:"ratingCount"`
	Featured        bool      `json:"featured"`
	Verified        bool      `json:"verified"`
	Price           float64   `json:"price"`
	Currency        string    `json:"currency"`
	Screenshots     []string  `json:"screenshots,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Author represents a template author
type Author struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar,omitempty"`
	Verified bool   `json:"verified"`
}

// Rating represents a template rating
type Rating struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Score     int       `json:"score"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// RatingStats represents rating statistics
type RatingStats struct {
	Average      float64        `json:"average"`
	Total        int            `json:"total"`
	Distribution map[int]int    `json:"distribution"`
}

// GetMarketplaceTemplates returns sample marketplace templates
func GetMarketplaceTemplates() []MarketplaceTemplate {
	return []MarketplaceTemplate{
		{
			ID:          "mkt-1",
			Name:        "Advanced Slack Integration",
			Description: "Full-featured Slack integration with channels, DMs, and reactions",
			Category:    "communication",
			Tags:        []string{"slack", "messaging", "notifications"},
			Author:      Author{ID: "author-1", Name: "LinkFlow Team", Verified: true},
			Version:     "2.1.0",
			Downloads:   3500,
			Rating:      4.8,
			RatingCount: 127,
			Featured:    true,
			Verified:    true,
			Price:       0,
			Currency:    "USD",
			CreatedAt:   time.Now().AddDate(0, -6, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -14),
		},
		{
			ID:          "mkt-2",
			Name:        "AI Document Processor",
			Description: "Process documents with GPT-4 and extract structured data",
			Category:    "ai",
			Tags:        []string{"ai", "gpt", "documents", "extraction"},
			Author:      Author{ID: "author-2", Name: "AI Labs", Verified: true},
			Version:     "1.5.0",
			Downloads:   2100,
			Rating:      4.7,
			RatingCount: 89,
			Featured:    true,
			Verified:    true,
			Price:       9.99,
			Currency:    "USD",
			CreatedAt:   time.Now().AddDate(0, -4, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -7),
		},
		{
			ID:          "mkt-3",
			Name:        "E-commerce Order Flow",
			Description: "Complete order processing workflow for Shopify and WooCommerce",
			Category:    "integration",
			Tags:        []string{"ecommerce", "shopify", "woocommerce", "orders"},
			Author:      Author{ID: "author-3", Name: "Commerce Pro", Verified: false},
			Version:     "3.0.0",
			Downloads:   1800,
			Rating:      4.5,
			RatingCount: 63,
			Featured:    true,
			Verified:    false,
			Price:       19.99,
			Currency:    "USD",
			CreatedAt:   time.Now().AddDate(0, -8, 0),
			UpdatedAt:   time.Now().AddDate(0, -1, 0),
		},
	}
}
