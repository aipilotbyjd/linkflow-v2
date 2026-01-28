package ai

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// VectorStoreNode handles vector database operations (upsert, search)
type VectorStoreNode struct{}

// NewVectorStoreNode creates a new vector store node
func NewVectorStoreNode() *VectorStoreNode {
	return &VectorStoreNode{}
}

func (n *VectorStoreNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	// Get operation
	operation, _ := params["operation"].(string)
	if operation == "" {
		operation = "upsert" // Default
	}

	// Get provider (pinecone, pgvector, etc)
	provider, _ := params["provider"].(string)
	if provider == "" {
		return nil, fmt.Errorf("vector store provider is required")
	}

	// Get credentials
	indexName, _ := params["index_name"].(string)

	// Mock implementation for now - replacing with actual DB calls in future
	// This structure ensures the node is ready for integration

	switch operation {
	case "upsert":
		// Get data to upsert
		var vectors []interface{}
		if v, ok := params["vectors"].([]interface{}); ok {
			vectors = v
		} else if _, ok := params["text"].(string); ok {
			// Auto-embed if just text is provided (would need embedding model config)
			// For now, assume vectors are passed from an embedding node
			return nil, fmt.Errorf("direct text upsert not yet supported, use embeddings node first")
		}

		if len(vectors) == 0 {
			return nil, fmt.Errorf("no vectors provided for upsert")
		}

		return types.JSON{
			"success":   true,
			"count":     len(vectors),
			"operation": "upsert",
			"provider":  provider,
			"index":     indexName,
		}, nil

	case "search":
		// Get query vector
		var vector []float64
		if v, ok := params["vector"].([]interface{}); ok {
			for _, val := range v {
				if f, ok := val.(float64); ok {
					vector = append(vector, f)
				}
			}
		}

		if k, ok := params["top_k"].(float64); ok && k > 0 {
			// topK = int(k)
		}

		if len(vector) == 0 {
			return nil, fmt.Errorf("query vector required for search")
		}

		// Mock search results
		return types.JSON{
			"matches": []map[string]interface{}{
				{
					"id":       "mock-doc-1",
					"score":    0.95,
					"metadata": map[string]interface{}{"text": "This is a relevant document"},
				},
				{
					"id":       "mock-doc-2",
					"score":    0.88,
					"metadata": map[string]interface{}{"text": "Another relevant document"},
				},
			},
			"count":     2,
			"operation": "search",
		}, nil

	case "delete":
		ids, _ := params["ids"].([]interface{})
		return types.JSON{
			"success":       true,
			"deleted_count": len(ids),
			"operation":     "delete",
		}, nil
	}

	return nil, fmt.Errorf("unknown operation: %s", operation)
}

func (n *VectorStoreNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "ai.vector_store",
		Name:        "Vector Store",
		Description: "Store and search vector embeddings (Pinecone, PgVector)",
		Category:    "ai",
		Version:     "1.0.0",
		Icon:        "Database01",
		Color:       "#10B981", // Emerald
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "provider",
				DisplayName: "Provider",
				Type:        "options",
				Required:    true,
				Default:     "pinecone",
				Options: []wtypes.ParamOption{
					{Name: "Pinecone", Value: "pinecone"},
					{Name: "PgVector", Value: "pgvector"},
					{Name: "Qdrant", Value: "qdrant"},
				},
			},
			{
				Name:        "credential",
				DisplayName: "Credential",
				Type:        "credential",
				Required:    true,
			},
			{
				Name:        "operation",
				DisplayName: "Operation",
				Type:        "options",
				Required:    true,
				Default:     "upsert",
				Options: []wtypes.ParamOption{
					{Name: "Upsert (Insert/Update)", Value: "upsert"},
					{Name: "Search (Query)", Value: "search"},
					{Name: "Delete", Value: "delete"},
				},
			},
			{
				Name:        "index_name",
				DisplayName: "Index Name",
				Type:        "string",
				Required:    true,
			},
			// Upsert params
			{
				Name:        "vectors",
				DisplayName: "Vectors",
				Type:        "json",
				Description: "Array of vectors to upsert (from Embeddings node)",
			},
			// Search params
			{
				Name:        "vector",
				DisplayName: "Query Vector",
				Type:        "json",
				Description: "Vector to search for",
			},
			{
				Name:        "top_k",
				DisplayName: "Top K",
				Type:        "number",
				Default:     5,
			},
		},
		Credentials: []string{"pinecone", "postgres", "qdrant"},
	}
}
