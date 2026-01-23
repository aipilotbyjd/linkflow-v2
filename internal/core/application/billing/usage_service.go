package billing

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// UsageService handles all usage tracking and enforcement
type UsageService struct {
	usageRepo        billing.UsageRepository
	subscriptionRepo billing.SubscriptionRepository
	alertService     *AlertService
	mu               sync.RWMutex
	cache            map[uuid.UUID]*cachedUsage // workspace -> usage cache
	rolloverCache    map[uuid.UUID][]*billing.CreditRollover
	byokCache        map[uuid.UUID][]*billing.BYOKConfig
}

type cachedUsage struct {
	usage     *billing.Usage
	limits    *billing.Limits
	planID    string
	expiresAt time.Time
}

// Errors
var (
	ErrOperationsExceeded   = errors.New("operations limit exceeded for this billing period")
	ErrAICreditsExceeded    = errors.New("AI credits limit exceeded for this billing period")
	ErrStorageExceeded      = errors.New("storage limit exceeded")
	ErrScenariosExceeded    = errors.New("active scenarios limit exceeded")
	ErrDataTransferExceeded = errors.New("data transfer limit exceeded")
	ErrFeatureNotAvailable  = errors.New("feature not available on your plan")
	ErrNoActiveSubscription = errors.New("no active subscription found")
)

func NewUsageService(usageRepo billing.UsageRepository, subscriptionRepo billing.SubscriptionRepository) *UsageService {
	return &UsageService{
		usageRepo:        usageRepo,
		subscriptionRepo: subscriptionRepo,
		alertService:     NewAlertService(),
		cache:            make(map[uuid.UUID]*cachedUsage),
		rolloverCache:    make(map[uuid.UUID][]*billing.CreditRollover),
		byokCache:        make(map[uuid.UUID][]*billing.BYOKConfig),
	}
}

// CheckAndConsumeOperations checks if operations are available and consumes them atomically
// Supports: task-free nodes, rollover credits, overage billing
func (s *UsageService) CheckAndConsumeOperations(ctx context.Context, workspaceID uuid.UUID, count int64, nodeType string) error {
	// Check if this is a task-free node (Zapier 2025 style)
	if billing.IsTaskFreeNode(nodeType) {
		return nil // Free, don't count
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	usage, limits, planID, err := s.getUsageAndLimitsWithPlan(ctx, workspaceID)
	if err != nil {
		return err
	}

	// -1 means unlimited
	if limits.OperationsPerMonth < 0 {
		usage.IncrementOperations(count)
		_ = s.usageRepo.Update(ctx, usage)
		s.updateCacheWithPlan(workspaceID, usage, limits, planID)
		return nil
	}

	limit := int64(limits.OperationsPerMonth)
	maxOverage := limit * int64(billing.DefaultOverageRates.MaxOverageMultiple)

	// First, try to consume from rollover credits
	remaining := count
	if rollovers := s.rolloverCache[workspaceID]; len(rollovers) > 0 {
		for _, r := range rollovers {
			if remaining <= 0 {
				break
			}
			consumed := r.ConsumeCredits(int(remaining))
			remaining -= int64(consumed)
		}
	}

	// Check if we have capacity (including overage)
	newTotal := usage.Operations + remaining
	if newTotal > maxOverage {
		// Hard stop - exceeded max overage
		s.alertService.TriggerAlert(workspaceID, billing.AlertTypeOperations, 100, usage.Operations, limit)
		return ErrOperationsExceeded
	}

	// Check if entering overage territory
	if newTotal > limit {
		// Allow but mark as overage
		usage.IncrementOperations(remaining)
		if err := s.usageRepo.Update(ctx, usage); err != nil {
			return err
		}
		// Trigger overage alert
		s.alertService.TriggerOverageAlert(workspaceID, billing.AlertTypeOperations, newTotal-limit)
	} else {
		usage.IncrementOperations(remaining)
		if err := s.usageRepo.Update(ctx, usage); err != nil {
			return err
		}
	}

	// Check and trigger threshold alerts
	percentage := float64(usage.Operations) / float64(limit) * 100
	s.alertService.CheckThresholds(workspaceID, billing.AlertTypeOperations, percentage, usage.Operations, limit)

	s.updateCacheWithPlan(workspaceID, usage, limits, planID)
	return nil
}

// CheckAndConsumeAICredits checks and consumes AI credits
// Supports: BYOK bypass, token-based calculation, rollover, overage
func (s *UsageService) CheckAndConsumeAICredits(ctx context.Context, workspaceID uuid.UUID, model string, inputTokens, outputTokens, images, audioMinutes int) error {
	// Check if using BYOK - if so, don't charge AI credits
	if s.isBYOKActive(workspaceID, model) {
		return nil // Using their own key, no charge
	}

	// Calculate credits based on model and token usage
	credits := int64(billing.CalculateAICredits(model, inputTokens, outputTokens, images, audioMinutes))
	if credits == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	usage, limits, planID, err := s.getUsageAndLimitsWithPlan(ctx, workspaceID)
	if err != nil {
		return err
	}

	// -1 means unlimited
	if limits.AICreditsPerMonth < 0 {
		usage.IncrementAICredits(credits)
		_ = s.usageRepo.Update(ctx, usage)
		s.updateCacheWithPlan(workspaceID, usage, limits, planID)
		return nil
	}

	limit := int64(limits.AICreditsPerMonth)
	maxOverage := limit * int64(billing.DefaultOverageRates.MaxOverageMultiple)

	// Try rollover AI credits first
	remaining := credits
	if rollovers := s.rolloverCache[workspaceID]; len(rollovers) > 0 {
		for _, r := range rollovers {
			if remaining <= 0 {
				break
			}
			consumed := r.ConsumeAICredits(int(remaining))
			remaining -= int64(consumed)
		}
	}

	newTotal := usage.AICreditsUsed + remaining
	if newTotal > maxOverage {
		s.alertService.TriggerAlert(workspaceID, billing.AlertTypeAICredits, 100, usage.AICreditsUsed, limit)
		return ErrAICreditsExceeded
	}

	usage.IncrementAICredits(remaining)
	if err := s.usageRepo.Update(ctx, usage); err != nil {
		return err
	}

	// Check thresholds
	percentage := float64(usage.AICreditsUsed) / float64(limit) * 100
	s.alertService.CheckThresholds(workspaceID, billing.AlertTypeAICredits, percentage, usage.AICreditsUsed, limit)

	s.updateCacheWithPlan(workspaceID, usage, limits, planID)
	return nil
}

// isBYOKActive checks if workspace has active BYOK for the model's provider
func (s *UsageService) isBYOKActive(workspaceID uuid.UUID, model string) bool {
	s.mu.RLock()
	configs := s.byokCache[workspaceID]
	s.mu.RUnlock()

	if len(configs) == 0 {
		return false
	}

	// Determine provider from model
	var provider billing.AIProvider
	switch {
	case containsPrefix(model, "gpt", "dall-e", "whisper", "tts", "o1"):
		provider = billing.ProviderOpenAI
	case containsPrefix(model, "claude"):
		provider = billing.ProviderAnthropic
	case containsPrefix(model, "gemini"):
		provider = billing.ProviderGoogle
	default:
		return false
	}

	return billing.IsBYOKEnabled(configs, provider)
}

func containsPrefix(s string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// CheckOperationsAvailable checks without consuming (for pre-flight check)
func (s *UsageService) CheckOperationsAvailable(ctx context.Context, workspaceID uuid.UUID, count int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	usage, limits, err := s.getUsageAndLimitsFromCache(ctx, workspaceID)
	if err != nil {
		return err
	}

	if limits.OperationsPerMonth > 0 {
		if usage.Operations+count > int64(limits.OperationsPerMonth) {
			return ErrOperationsExceeded
		}
	}
	return nil
}

// CheckAICreditsAvailable checks without consuming
func (s *UsageService) CheckAICreditsAvailable(ctx context.Context, workspaceID uuid.UUID, credits int64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	usage, limits, err := s.getUsageAndLimitsFromCache(ctx, workspaceID)
	if err != nil {
		return err
	}

	if limits.AICreditsPerMonth > 0 {
		if usage.AICreditsUsed+credits > int64(limits.AICreditsPerMonth) {
			return ErrAICreditsExceeded
		}
	}
	return nil
}

// CheckFeatureAccess verifies if a feature is available on the plan
func (s *UsageService) CheckFeatureAccess(ctx context.Context, workspaceID uuid.UUID, feature string) error {
	_, limits, err := s.getUsageAndLimitsFromCache(ctx, workspaceID)
	if err != nil {
		return err
	}

	switch feature {
	case "api_access":
		if !limits.HasAPIAccess {
			return ErrFeatureNotAvailable
		}
	case "priority_execution":
		if !limits.HasPriorityExecution {
			return ErrFeatureNotAvailable
		}
	case "custom_variables":
		if !limits.HasCustomVariables {
			return ErrFeatureNotAvailable
		}
	case "full_text_search":
		if !limits.HasFullTextSearch {
			return ErrFeatureNotAvailable
		}
	case "team_roles":
		if !limits.HasTeamRoles {
			return ErrFeatureNotAvailable
		}
	case "template_sharing":
		if !limits.HasTemplateSharing {
			return ErrFeatureNotAvailable
		}
	case "sso":
		if !limits.HasSSO {
			return ErrFeatureNotAvailable
		}
	case "audit_logs":
		if !limits.HasAuditLogs {
			return ErrFeatureNotAvailable
		}
	case "custom_functions":
		if !limits.HasCustomFunctions {
			return ErrFeatureNotAvailable
		}
	}
	return nil
}

// CheckActiveScenarios checks if more scenarios can be activated
func (s *UsageService) CheckActiveScenarios(ctx context.Context, workspaceID uuid.UUID, currentActive int) error {
	_, limits, err := s.getUsageAndLimitsFromCache(ctx, workspaceID)
	if err != nil {
		return err
	}

	if limits.ActiveScenarios > 0 && currentActive >= limits.ActiveScenarios {
		return ErrScenariosExceeded
	}
	return nil
}

// GetMinInterval returns minimum execution interval in minutes
func (s *UsageService) GetMinInterval(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	_, limits, err := s.getUsageAndLimitsFromCache(ctx, workspaceID)
	if err != nil {
		return 15, err // default to free plan interval
	}
	return limits.MinIntervalMinutes, nil
}

// GetUsageStatus returns current usage status for display
func (s *UsageService) GetUsageStatus(ctx context.Context, workspaceID uuid.UUID) (*billing.UsageLimitStatus, error) {
	usage, limits, err := s.getUsageAndLimits(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return usage.CheckLimits(limits), nil
}

// TrackDataTransfer records data transfer
func (s *UsageService) TrackDataTransfer(ctx context.Context, workspaceID uuid.UUID, bytes int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	usage, limits, err := s.getUsageAndLimits(ctx, workspaceID)
	if err != nil {
		return err
	}

	mb := bytes / (1024 * 1024)
	if limits.DataTransferMB > 0 && usage.DataTransferMB+mb > int64(limits.DataTransferMB) {
		return ErrDataTransferExceeded
	}

	usage.IncrementDataTransfer(mb)
	return s.usageRepo.Update(ctx, usage)
}

// ResetMonthlyUsage resets usage for new billing period
func (s *UsageService) ResetMonthlyUsage(ctx context.Context, workspaceID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	usage := billing.NewUsage(workspaceID, periodStart, periodEnd)
	if err := s.usageRepo.Create(ctx, usage); err != nil {
		return err
	}

	// Invalidate cache
	delete(s.cache, workspaceID)
	return nil
}

// CalculateOverage calculates overage charges
func (s *UsageService) CalculateOverage(ctx context.Context, workspaceID uuid.UUID) (*OverageDetails, error) {
	usage, limits, err := s.getUsageAndLimits(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	details := &OverageDetails{}

	if limits.OperationsPerMonth > 0 && usage.Operations > int64(limits.OperationsPerMonth) {
		details.ExtraOperations = int(usage.Operations - int64(limits.OperationsPerMonth))
		details.OperationsCharge = billing.CalculateOverage(details.ExtraOperations, 0, 0, nil)
	}

	if limits.AICreditsPerMonth > 0 && usage.AICreditsUsed > int64(limits.AICreditsPerMonth) {
		details.ExtraAICredits = int(usage.AICreditsUsed - int64(limits.AICreditsPerMonth))
		details.AICreditsCharge = billing.CalculateOverage(0, details.ExtraAICredits, 0, nil)
	}

	details.TotalCharge = details.OperationsCharge + details.AICreditsCharge
	return details, nil
}

// OverageDetails contains overage calculation results
type OverageDetails struct {
	ExtraOperations  int   `json:"extra_operations"`
	OperationsCharge int64 `json:"operations_charge_cents"`
	ExtraAICredits   int   `json:"extra_ai_credits"`
	AICreditsCharge  int64 `json:"ai_credits_charge_cents"`
	TotalCharge      int64 `json:"total_charge_cents"`
}

// Internal helpers

func (s *UsageService) getUsageAndLimits(ctx context.Context, workspaceID uuid.UUID) (*billing.Usage, *billing.Limits, error) {
	// Get subscription - use free plan if no subscription exists
	planID := "free"
	sub, err := s.subscriptionRepo.FindByWorkspaceID(ctx, workspaceID)
	if err == nil && sub != nil && sub.IsActive() {
		planID = sub.PlanID
	}

	// Get plan limits (defaults to free plan if not found)
	plan := billing.GetPlan(planID)
	limits := &plan.Limits

	// Get or create usage for current period
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	usage, err := s.usageRepo.FindByWorkspaceAndPeriod(ctx, workspaceID, periodStart, periodEnd)
	if err != nil {
		// Create new usage record
		usage = billing.NewUsage(workspaceID, periodStart, periodEnd)
		if err := s.usageRepo.Create(ctx, usage); err != nil {
			return nil, nil, err
		}
	}

	return usage, limits, nil
}

func (s *UsageService) getUsageAndLimitsFromCache(ctx context.Context, workspaceID uuid.UUID) (*billing.Usage, *billing.Limits, error) {
	// Check cache first
	if cached, ok := s.cache[workspaceID]; ok && time.Now().Before(cached.expiresAt) {
		return cached.usage, cached.limits, nil
	}

	// Cache miss, fetch from DB
	return s.getUsageAndLimits(ctx, workspaceID)
}

func (s *UsageService) updateCacheWithPlan(workspaceID uuid.UUID, usage *billing.Usage, limits *billing.Limits, planID string) {
	s.cache[workspaceID] = &cachedUsage{
		usage:     usage,
		limits:    limits,
		planID:    planID,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
}

func (s *UsageService) getUsageAndLimitsWithPlan(ctx context.Context, workspaceID uuid.UUID) (*billing.Usage, *billing.Limits, string, error) {
	// Get subscription - use free plan if no subscription exists
	planID := "free"
	sub, err := s.subscriptionRepo.FindByWorkspaceID(ctx, workspaceID)
	if err == nil && sub != nil && sub.IsActive() {
		planID = sub.PlanID
	}

	// Get plan limits (defaults to free plan if not found)
	plan := billing.GetPlan(planID)
	limits := &plan.Limits

	// Get or create usage for current period
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	usage, err := s.usageRepo.FindByWorkspaceAndPeriod(ctx, workspaceID, periodStart, periodEnd)
	if err != nil {
		// Create new usage record
		usage = billing.NewUsage(workspaceID, periodStart, periodEnd)
		if err := s.usageRepo.Create(ctx, usage); err != nil {
			return nil, nil, "", err
		}
	}

	return usage, limits, planID, nil
}

// SetBYOKConfig caches BYOK configuration for a workspace
func (s *UsageService) SetBYOKConfig(workspaceID uuid.UUID, configs []*billing.BYOKConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byokCache[workspaceID] = configs
}

// SetRolloverCredits caches rollover credits for a workspace
func (s *UsageService) SetRolloverCredits(workspaceID uuid.UUID, rollovers []*billing.CreditRollover) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rolloverCache[workspaceID] = rollovers
}

// GetUsageDashboard returns comprehensive usage data for dashboard
func (s *UsageService) GetUsageDashboard(ctx context.Context, workspaceID uuid.UUID) (*UsageDashboard, error) {
	usage, limits, planID, err := s.getUsageAndLimitsWithPlan(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	plan := billing.GetPlan(planID)

	// Calculate projections
	now := time.Now()
	daysInMonth := float64(time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day())
	daysPassed := float64(now.Day())

	var projectedOps, projectedAI int64
	if daysPassed > 0 {
		dailyRate := float64(usage.Operations) / daysPassed
		projectedOps = int64(dailyRate * daysInMonth)

		dailyAIRate := float64(usage.AICreditsUsed) / daysPassed
		projectedAI = int64(dailyAIRate * daysInMonth)
	}

	// Calculate rollover available
	var rolloverOps, rolloverAI int
	s.mu.RLock()
	rollovers := s.rolloverCache[workspaceID]
	s.mu.RUnlock()
	for _, r := range rollovers {
		rolloverOps += r.RemainingCredits()
		rolloverAI += r.RemainingAICredits()
	}

	// Calculate overage
	var overageOps, overageAI int64
	var overageCharge int64
	if limits.OperationsPerMonth > 0 && usage.Operations > int64(limits.OperationsPerMonth) {
		overageOps = usage.Operations - int64(limits.OperationsPerMonth)
		overageCharge += int64(float64(overageOps/1000) * float64(billing.DefaultOverageRates.OperationsPer1000) * billing.DefaultOverageRates.OverageMultiplier)
	}
	if limits.AICreditsPerMonth > 0 && usage.AICreditsUsed > int64(limits.AICreditsPerMonth) {
		overageAI = usage.AICreditsUsed - int64(limits.AICreditsPerMonth)
		overageCharge += int64(float64(overageAI) * float64(billing.DefaultOverageRates.AICreditsPerCredit) * billing.DefaultOverageRates.OverageMultiplier)
	}

	return &UsageDashboard{
		PlanID:   planID,
		PlanName: plan.Name,

		// Operations
		OperationsUsed:      usage.Operations,
		OperationsLimit:     int64(limits.OperationsPerMonth),
		OperationsPercent:   safePercent(usage.Operations, int64(limits.OperationsPerMonth)),
		OperationsProjected: projectedOps,
		OperationsRollover:  int64(rolloverOps),

		// AI Credits
		AICreditsUsed:      usage.AICreditsUsed,
		AICreditsLimit:     int64(limits.AICreditsPerMonth),
		AICreditsPercent:   safePercent(usage.AICreditsUsed, int64(limits.AICreditsPerMonth)),
		AICreditsProjected: projectedAI,
		AICreditsRollover:  int64(rolloverAI),

		// Overage
		OverageOperations:  overageOps,
		OverageAICredits:   overageAI,
		OverageChargeCents: overageCharge,

		// Period
		PeriodStart:   usage.PeriodStart,
		PeriodEnd:     usage.PeriodEnd,
		DaysRemaining: int(daysInMonth - daysPassed),
	}, nil
}

// UsageDashboard contains comprehensive usage data
type UsageDashboard struct {
	PlanID   string `json:"plan_id"`
	PlanName string `json:"plan_name"`

	// Operations
	OperationsUsed      int64   `json:"operations_used"`
	OperationsLimit     int64   `json:"operations_limit"`
	OperationsPercent   float64 `json:"operations_percent"`
	OperationsProjected int64   `json:"operations_projected"`
	OperationsRollover  int64   `json:"operations_rollover"`

	// AI Credits
	AICreditsUsed      int64   `json:"ai_credits_used"`
	AICreditsLimit     int64   `json:"ai_credits_limit"`
	AICreditsPercent   float64 `json:"ai_credits_percent"`
	AICreditsProjected int64   `json:"ai_credits_projected"`
	AICreditsRollover  int64   `json:"ai_credits_rollover"`

	// Overage
	OverageOperations  int64 `json:"overage_operations"`
	OverageAICredits   int64 `json:"overage_ai_credits"`
	OverageChargeCents int64 `json:"overage_charge_cents"`

	// Period
	PeriodStart   time.Time `json:"period_start"`
	PeriodEnd     time.Time `json:"period_end"`
	DaysRemaining int       `json:"days_remaining"`
}

func safePercent(used, limit int64) float64 {
	if limit <= 0 {
		return 0
	}
	return float64(used) / float64(limit) * 100
}
