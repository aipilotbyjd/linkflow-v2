package template

type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

var DefaultCategories = []Category{
	{ID: "marketing", Name: "Marketing", Description: "Marketing automation templates", Icon: "megaphone", Color: "#FF6B6B"},
	{ID: "sales", Name: "Sales", Description: "Sales automation templates", Icon: "chart-line", Color: "#4ECDC4"},
	{ID: "hr", Name: "HR", Description: "Human resources templates", Icon: "users", Color: "#45B7D1"},
	{ID: "finance", Name: "Finance", Description: "Finance and accounting templates", Icon: "dollar-sign", Color: "#96CEB4"},
	{ID: "development", Name: "Development", Description: "Developer workflow templates", Icon: "code", Color: "#9B59B6"},
	{ID: "support", Name: "Support", Description: "Customer support templates", Icon: "headphones", Color: "#F39C12"},
	{ID: "data", Name: "Data", Description: "Data processing templates", Icon: "database", Color: "#1ABC9C"},
	{ID: "communication", Name: "Communication", Description: "Team communication templates", Icon: "message-circle", Color: "#E74C3C"},
	{ID: "other", Name: "Other", Description: "Other templates", Icon: "folder", Color: "#95A5A6"},
}

func GetCategory(id string) *Category {
	for _, c := range DefaultCategories {
		if c.ID == id {
			return &c
		}
	}
	return nil
}

func GetAllCategories() []Category {
	return DefaultCategories
}
