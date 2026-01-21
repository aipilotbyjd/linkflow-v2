package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// BrowserNode handles browser automation (navigate, scrape, click)
type BrowserNode struct{}

// NewBrowserNode creates a new browser node
func NewBrowserNode() *BrowserNode {
	return &BrowserNode{}
}

func (n *BrowserNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	action, _ := params["action"].(string)
	url, _ := params["url"].(string)

	if action == "" {
		action = "scrape" // Default
	}
	if url == "" && action != "close" {
		return nil, fmt.Errorf("URL is required")
	}

	// In a real implementation, this would connect to a headless browser service (Playwright/Chrome)
	// For now, we stub the functionality to define the interface

	switch action {
	case "scrape":
		selector, _ := params["selector"].(string)
		if selector == "" {
			selector = "body"
		}

		// Mock scraping result
		return types.JSON{
			"url":       url,
			"title":     "Example Domain",
			"content":   "This is a mocked scraping result for " + url,
			"selector":  selector,
			"text":      "Example content extracted from selector",
			"timestamp": time.Now().Format(time.RFC3339),
		}, nil

	case "screenshot":
		fullPage, _ := params["full_page"].(bool)
		
		// Mock screenshot result
		return types.JSON{
			"url":       url,
			"image_url": "https://storage.example.com/screenshots/mock-123.png",
			"format":    "png",
			"full_page": fullPage,
			"size_bytes": 102400,
		}, nil

	case "extract_data":
		// Mock structured data extraction
		return types.JSON{
			"url": url,
			"data": []map[string]interface{}{
				{"title": "Item 1", "price": "$10.00"},
				{"title": "Item 2", "price": "$20.00"},
			},
			"count": 2,
		}, nil
	}

	return nil, fmt.Errorf("unknown action: %s", action)
}

func (n *BrowserNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "action.browser",
		Name:        "Browser Automation",
		Description: "Automate web browsing (scrape, screenshot, interact)",
		Category:    "action",
		Version:     "1.0.0",
		Icon:        "globe",
		Color:       "#3B82F6", // Blue
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "action",
				DisplayName: "Action",
				Type:        "options",
				Required:    true,
				Default:     "scrape",
				Options: []wtypes.ParamOption{
					{Name: "Scrape Content", Value: "scrape"},
					{Name: "Take Screenshot", Value: "screenshot"},
					{Name: "Extract Structured Data", Value: "extract_data"},
					{Name: "Navigate", Value: "navigate"},
				},
			},
			{
				Name:        "url",
				DisplayName: "URL",
				Type:        "string",
				Required:    true,
				Description: "Website URL to visit",
			},
			{
				Name:        "selector",
				DisplayName: "CSS Selector",
				Type:        "string",
				Description: "CSS selector to target specific element",
			},
			{
				Name:        "full_page",
				DisplayName: "Full Page",
				Type:        "boolean",
				Default:     false,
			},
			{
				Name:        "wait_for",
				DisplayName: "Wait For Selector",
				Type:        "string",
				Description: "Wait for this selector to appear before acting",
			},
		},
	}
}
