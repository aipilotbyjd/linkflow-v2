package gateway

import (
	"context"
	"sort"

	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
)

// Router handles intelligent model routing
type Router struct {
	defaultModel string
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		defaultModel: "gpt-4o-mini",
	}
}

// Route determines the best model for a request
func (r *Router) Route(ctx context.Context, req *ai.RouterRequest, providers map[ai.Provider]ai.ProviderAdapter) (*ai.RouterResponse, error) {
	candidates := r.getCandidateModels(req)

	if len(candidates) == 0 {
		// Fall back to default
		model, _ := ai.GetModel(r.defaultModel)
		return &ai.RouterResponse{
			SelectedModel:    model,
			SelectedProvider: model.Provider,
			Reason:           "default model selected, no candidates matched criteria",
		}, nil
	}

	// Score and rank candidates
	scored := r.scoreModels(candidates, req)

	// Sort by score (descending)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Filter to available providers
	for _, s := range scored {
		if _, ok := providers[s.model.Provider]; ok {
			return &ai.RouterResponse{
				SelectedModel:    s.model,
				SelectedProvider: s.model.Provider,
				Reason:           s.reason,
				Alternatives:     r.getAlternatives(scored, s.model.ID),
			}, nil
		}
	}

	// No available provider, return first candidate
	return &ai.RouterResponse{
		SelectedModel:    scored[0].model,
		SelectedProvider: scored[0].model.Provider,
		Reason:           "selected best match (provider may not be configured)",
		Alternatives:     r.getAlternatives(scored, scored[0].model.ID),
	}, nil
}

type scoredModel struct {
	model  ai.Model
	score  float64
	reason string
}

func (r *Router) getCandidateModels(req *ai.RouterRequest) []ai.Model {
	var candidates []ai.Model

	for _, model := range ai.ModelRegistry {
		// Filter by provider preference
		if len(req.PreferredProviders) > 0 {
			found := false
			for _, p := range req.PreferredProviders {
				if model.Provider == p {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by cost
		if req.MaxCostPer1M > 0 {
			avgCost := (model.InputPricePer1M + model.OutputPricePer1M) / 2
			if avgCost > req.MaxCostPer1M {
				continue
			}
		}

		// Filter by capabilities
		if req.RequireVision && !model.SupportsVision {
			continue
		}
		if req.RequireTools && !model.SupportsTools {
			continue
		}

		candidates = append(candidates, model)
	}

	return candidates
}

func (r *Router) scoreModels(models []ai.Model, req *ai.RouterRequest) []scoredModel {
	scored := make([]scoredModel, len(models))

	for i, model := range models {
		score := 50.0 // Base score
		reason := "balanced selection"

		// Cost scoring (lower is better for cost-conscious)
		avgCost := (model.InputPricePer1M + model.OutputPricePer1M) / 2
		if avgCost < 1.0 {
			score += 20
			reason = "cost-effective model"
		} else if avgCost < 5.0 {
			score += 10
		} else if avgCost > 20.0 {
			score -= 10
		}

		// Speed preference
		if req.PreferSpeed {
			// Smaller/faster models get bonus
			if model.ID == "gpt-4o-mini" || model.ID == "claude-3-5-haiku-20241022" || model.ID == "gemini-1.5-flash" {
				score += 25
				reason = "fast model for speed preference"
			}
		}

		// Quality preference
		if req.PreferQuality {
			// Larger/better models get bonus
			if model.ID == "gpt-4o" || model.ID == "claude-3-5-sonnet-20241022" || model.ID == "o1" {
				score += 25
				reason = "high-quality model for quality preference"
			}
		}

		// Capability bonus
		if model.SupportsVision && req.RequireVision {
			score += 10
		}
		if model.SupportsTools && req.RequireTools {
			score += 10
		}
		if model.SupportsJSON {
			score += 5
		}

		// Context window bonus for long conversations
		if len(req.Messages) > 10 && model.ContextWindow > 100000 {
			score += 10
			reason = "large context for long conversation"
		}

		scored[i] = scoredModel{
			model:  model,
			score:  score,
			reason: reason,
		}
	}

	return scored
}

func (r *Router) getAlternatives(scored []scoredModel, selectedID string) []ai.Model {
	var alternatives []ai.Model
	count := 0

	for _, s := range scored {
		if s.model.ID != selectedID && count < 3 {
			alternatives = append(alternatives, s.model)
			count++
		}
	}

	return alternatives
}

// EstimateComplexity estimates the complexity of a request
func (r *Router) EstimateComplexity(messages []ai.Message) string {
	totalLength := 0
	hasImages := false
	hasCode := false

	for _, msg := range messages {
		text := msg.GetText()
		totalLength += len(text)

		if msg.HasImage() {
			hasImages = true
		}

		// Simple heuristic for code detection
		if containsCode(text) {
			hasCode = true
		}
	}

	if hasImages {
		return "high"
	}
	if hasCode || totalLength > 5000 {
		return "high"
	}
	if totalLength > 1000 || len(messages) > 5 {
		return "medium"
	}
	return "low"
}

func containsCode(text string) bool {
	codeIndicators := []string{
		"```", "def ", "func ", "function ", "class ", "import ",
		"const ", "let ", "var ", "package ", "public ", "private ",
	}

	for _, indicator := range codeIndicators {
		if len(text) > len(indicator) {
			for i := 0; i <= len(text)-len(indicator); i++ {
				if text[i:i+len(indicator)] == indicator {
					return true
				}
			}
		}
	}
	return false
}
