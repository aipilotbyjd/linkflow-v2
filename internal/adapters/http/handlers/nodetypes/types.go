package nodetypes

// CategoryResponse represents a node category with its count
type CategoryResponse struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
