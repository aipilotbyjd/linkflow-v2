package billing

// Plan represents a subscription plan (Make.com style)
type Plan struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	PriceMonthly  int64    `json:"price_monthly"` // in cents
	PriceYearly   int64    `json:"price_yearly"`  // in cents
	Currency      string   `json:"currency"`
	Features      []string `json:"features"`
	Limits        Limits   `json:"limits"`
	StripePriceID string   `json:"stripe_price_id,omitempty"`
	IsPerUser     bool     `json:"is_per_user"` // Teams plan is per-user
	Popular       bool     `json:"popular"`     // highlight on pricing page
}

// Limits defines plan limits (Make.com style - operations/credits based)
type Limits struct {
	// Operations (credits) - core billing unit
	OperationsPerMonth int `json:"operations_per_month"` // -1 = unlimited

	// AI Credits - separate pool for AI features
	AICreditsPerMonth int `json:"ai_credits_per_month"`

	// Scenarios (workflows)
	ActiveScenarios int `json:"active_scenarios"` // -1 = unlimited

	// Team
	TeamMembers int `json:"team_members"` // -1 = unlimited

	// Execution intervals
	MinIntervalMinutes int `json:"min_interval_minutes"` // minimum time between runs

	// Data & Storage
	DataTransferMB int `json:"data_transfer_mb"`
	FileStorageMB  int `json:"file_storage_mb"`
	RetentionDays  int `json:"retention_days"`

	// Features
	HasAPIAccess         bool `json:"has_api_access"`
	HasPriorityExecution bool `json:"has_priority_execution"`
	HasCustomVariables   bool `json:"has_custom_variables"`
	HasFullTextSearch    bool `json:"has_full_text_search"`
	HasTeamRoles         bool `json:"has_team_roles"`
	HasTemplateSharing   bool `json:"has_template_sharing"`
	HasSSO               bool `json:"has_sso"`
	HasAuditLogs         bool `json:"has_audit_logs"`
	Has24x7Support       bool `json:"has_24x7_support"`
	HasCustomFunctions   bool `json:"has_custom_functions"`
}

// OverageRates defines cost for exceeding limits
type OverageRates struct {
	OperationsPer1000  int64   `json:"operations_per_1000_cents"` // cost per 1000 extra ops
	AICreditsPerCredit int64   `json:"ai_credits_per_credit_cents"`
	StoragePerGB       int64   `json:"storage_per_gb_cents"`
	ExtraSeatMonthly   int64   `json:"extra_seat_monthly_cents"`
	OverageMultiplier  float64 `json:"overage_multiplier"`   // 1.25x for Zapier-style
	MaxOverageMultiple int     `json:"max_overage_multiple"` // Cap at 3x plan limit
}

// Default overage rates
var DefaultOverageRates = OverageRates{
	OperationsPer1000:  350,  // $3.50 per 1000 operations
	AICreditsPerCredit: 1,    // $0.01 per AI credit
	StoragePerGB:       100,  // $1.00 per GB
	ExtraSeatMonthly:   500,  // $5.00 per extra seat
	OverageMultiplier:  1.25, // 1.25x rate for overage (Zapier style)
	MaxOverageMultiple: 3,    // Max 3x plan limit before hard stop
}

// TaskFreeNodes - these node types don't count toward billing (Zapier 2025 style)
var TaskFreeNodes = map[string]bool{
	"logic.filter":     true,
	"logic.switch":     true,
	"logic.if":         true,
	"logic.router":     true,
	"transform.format": true,
	"transform.set":    true,
	"logic.delay":      true,
	"logic.wait":       true,
	"logic.loop":       true, // Loop itself is free, iterations count
	"logic.noop":       true,
	"logic.merge":      true,
	"logic.split":      true,
}

// IsTaskFreeNode checks if a node type is free (doesn't count toward billing)
func IsTaskFreeNode(nodeType string) bool {
	return TaskFreeNodes[nodeType]
}

// AIModelCredits defines credit costs for different AI models (token-based)
type AIModelCredits struct {
	InputPer1KTokens  int `json:"input_per_1k_tokens"`
	OutputPer1KTokens int `json:"output_per_1k_tokens"`
	PerImage          int `json:"per_image,omitempty"`
	PerMinuteAudio    int `json:"per_minute_audio,omitempty"`
}

// AIModelCosts maps AI models to their credit costs
var AIModelCosts = map[string]AIModelCredits{
	// OpenAI
	"gpt-4":         {InputPer1KTokens: 30, OutputPer1KTokens: 60},
	"gpt-4-turbo":   {InputPer1KTokens: 10, OutputPer1KTokens: 30},
	"gpt-4o":        {InputPer1KTokens: 5, OutputPer1KTokens: 15},
	"gpt-4o-mini":   {InputPer1KTokens: 1, OutputPer1KTokens: 3},
	"gpt-3.5-turbo": {InputPer1KTokens: 1, OutputPer1KTokens: 2},
	"dall-e-3":      {PerImage: 40},
	"dall-e-2":      {PerImage: 20},
	"whisper-1":     {PerMinuteAudio: 6},
	"tts-1":         {InputPer1KTokens: 15},
	"tts-1-hd":      {InputPer1KTokens: 30},

	// Anthropic
	"claude-3-opus":     {InputPer1KTokens: 15, OutputPer1KTokens: 75},
	"claude-3-sonnet":   {InputPer1KTokens: 3, OutputPer1KTokens: 15},
	"claude-3-haiku":    {InputPer1KTokens: 1, OutputPer1KTokens: 5},
	"claude-3.5-sonnet": {InputPer1KTokens: 3, OutputPer1KTokens: 15},

	// Google
	"gemini-pro":   {InputPer1KTokens: 1, OutputPer1KTokens: 2},
	"gemini-ultra": {InputPer1KTokens: 5, OutputPer1KTokens: 15},

	// Default for unknown models
	"default": {InputPer1KTokens: 5, OutputPer1KTokens: 10},
}

// CalculateAICredits calculates credits for AI usage
func CalculateAICredits(model string, inputTokens, outputTokens, images, audioMinutes int) int {
	costs, ok := AIModelCosts[model]
	if !ok {
		costs = AIModelCosts["default"]
	}

	credits := 0
	credits += (inputTokens / 1000) * costs.InputPer1KTokens
	credits += (outputTokens / 1000) * costs.OutputPer1KTokens
	credits += images * costs.PerImage
	credits += audioMinutes * costs.PerMinuteAudio

	// Minimum 1 credit for any AI operation
	if credits < 1 && (inputTokens > 0 || outputTokens > 0 || images > 0 || audioMinutes > 0) {
		credits = 1
	}

	return credits
}

// ===== PLANS (Make.com style) =====

var FreePlan = Plan{
	ID:           "free",
	Name:         "Free",
	Description:  "For individuals exploring automation",
	PriceMonthly: 0,
	PriceYearly:  0,
	Currency:     "USD",
	Popular:      false,
	IsPerUser:    false,
	Features: []string{
		"1,000 operations/month",
		"2 active scenarios",
		"No-code visual builder",
		"2,000+ app integrations",
		"Community support",
		"15-minute minimum interval",
	},
	Limits: Limits{
		OperationsPerMonth:   1000,
		AICreditsPerMonth:    100,
		ActiveScenarios:      2,
		TeamMembers:          1,
		MinIntervalMinutes:   15,
		DataTransferMB:       100,
		FileStorageMB:        500,
		RetentionDays:        7,
		HasAPIAccess:         false,
		HasPriorityExecution: false,
		HasCustomVariables:   false,
		HasFullTextSearch:    false,
		HasTeamRoles:         false,
		HasTemplateSharing:   false,
		HasSSO:               false,
		HasAuditLogs:         false,
		Has24x7Support:       false,
		HasCustomFunctions:   false,
	},
}

var CorePlan = Plan{
	ID:           "core",
	Name:         "Core",
	Description:  "For freelancers and solopreneurs",
	PriceMonthly: 900,  // $9
	PriceYearly:  9000, // $90 (2 months free)
	Currency:     "USD",
	Popular:      false,
	IsPerUser:    false,
	Features: []string{
		"10,000 operations/month",
		"Unlimited active scenarios",
		"1-minute minimum interval",
		"API access",
		"Email support",
		"500 AI credits/month",
	},
	Limits: Limits{
		OperationsPerMonth:   10000,
		AICreditsPerMonth:    500,
		ActiveScenarios:      -1,
		TeamMembers:          1,
		MinIntervalMinutes:   1,
		DataTransferMB:       1000,
		FileStorageMB:        2000,
		RetentionDays:        30,
		HasAPIAccess:         true,
		HasPriorityExecution: false,
		HasCustomVariables:   false,
		HasFullTextSearch:    false,
		HasTeamRoles:         false,
		HasTemplateSharing:   false,
		HasSSO:               false,
		HasAuditLogs:         false,
		Has24x7Support:       false,
		HasCustomFunctions:   false,
	},
}

var ProPlan = Plan{
	ID:           "pro",
	Name:         "Pro",
	Description:  "For growing businesses",
	PriceMonthly: 1600,  // $16
	PriceYearly:  16000, // $160 (2 months free)
	Currency:     "USD",
	Popular:      true,
	IsPerUser:    false,
	Features: []string{
		"10,000 operations/month",
		"Unlimited active scenarios",
		"Priority execution",
		"Custom variables",
		"Full-text execution search",
		"Priority email support",
		"1,000 AI credits/month",
	},
	Limits: Limits{
		OperationsPerMonth:   10000,
		AICreditsPerMonth:    1000,
		ActiveScenarios:      -1,
		TeamMembers:          1,
		MinIntervalMinutes:   1,
		DataTransferMB:       5000,
		FileStorageMB:        5000,
		RetentionDays:        60,
		HasAPIAccess:         true,
		HasPriorityExecution: true,
		HasCustomVariables:   true,
		HasFullTextSearch:    true,
		HasTeamRoles:         false,
		HasTemplateSharing:   false,
		HasSSO:               false,
		HasAuditLogs:         false,
		Has24x7Support:       false,
		HasCustomFunctions:   false,
	},
}

var TeamsPlan = Plan{
	ID:           "teams",
	Name:         "Teams",
	Description:  "For SMB teams collaborating on automation",
	PriceMonthly: 2900,  // $29 per user
	PriceYearly:  29000, // $290 per user (2 months free)
	Currency:     "USD",
	Popular:      false,
	IsPerUser:    true,
	Features: []string{
		"10,000 operations/month per user",
		"Everything in Pro",
		"Team roles & permissions",
		"Shared scenario templates",
		"Team dashboard",
		"2,000 AI credits/month per user",
	},
	Limits: Limits{
		OperationsPerMonth:   10000, // per user
		AICreditsPerMonth:    2000,  // per user
		ActiveScenarios:      -1,
		TeamMembers:          -1,
		MinIntervalMinutes:   1,
		DataTransferMB:       10000,
		FileStorageMB:        10000,
		RetentionDays:        90,
		HasAPIAccess:         true,
		HasPriorityExecution: true,
		HasCustomVariables:   true,
		HasFullTextSearch:    true,
		HasTeamRoles:         true,
		HasTemplateSharing:   true,
		HasSSO:               false,
		HasAuditLogs:         false,
		Has24x7Support:       false,
		HasCustomFunctions:   false,
	},
}

var EnterprisePlan = Plan{
	ID:           "enterprise",
	Name:         "Enterprise",
	Description:  "For organizations with critical automation needs",
	PriceMonthly: 0, // Custom
	PriceYearly:  0, // Custom
	Currency:     "USD",
	Popular:      false,
	IsPerUser:    true,
	Features: []string{
		"Custom operations volume",
		"Everything in Teams",
		"SSO / SAML",
		"Audit logs",
		"Custom functions",
		"24/7 priority support",
		"Dedicated success manager",
		"SLA guarantee",
		"Unlimited AI credits",
	},
	Limits: Limits{
		OperationsPerMonth:   -1, // Custom/Unlimited
		AICreditsPerMonth:    -1, // Unlimited
		ActiveScenarios:      -1,
		TeamMembers:          -1,
		MinIntervalMinutes:   1,
		DataTransferMB:       -1,
		FileStorageMB:        -1,
		RetentionDays:        365,
		HasAPIAccess:         true,
		HasPriorityExecution: true,
		HasCustomVariables:   true,
		HasFullTextSearch:    true,
		HasTeamRoles:         true,
		HasTemplateSharing:   true,
		HasSSO:               true,
		HasAuditLogs:         true,
		Has24x7Support:       true,
		HasCustomFunctions:   true,
	},
}

// GetPlan returns a plan by ID
func GetPlan(id string) *Plan {
	switch id {
	case "free":
		return &FreePlan
	case "core":
		return &CorePlan
	case "pro":
		return &ProPlan
	case "teams":
		return &TeamsPlan
	case "enterprise":
		return &EnterprisePlan
	default:
		return &FreePlan
	}
}

// GetAllPlans returns all available plans
func GetAllPlans() []Plan {
	return []Plan{FreePlan, CorePlan, ProPlan, TeamsPlan, EnterprisePlan}
}

// GetPublicPlans returns plans shown on pricing page
func GetPublicPlans() []Plan {
	return []Plan{FreePlan, CorePlan, ProPlan, TeamsPlan}
}

// CalculateTeamPrice calculates price for Teams plan based on seats
func CalculateTeamPrice(seats int, yearly bool) int64 {
	if yearly {
		return TeamsPlan.PriceYearly * int64(seats)
	}
	return TeamsPlan.PriceMonthly * int64(seats)
}

// CalculateOverage calculates overage charges
func CalculateOverage(extraOps, extraAI, extraStorageMB int, rates *OverageRates) int64 {
	if rates == nil {
		rates = &DefaultOverageRates
	}

	var total int64
	if extraOps > 0 {
		total += (int64(extraOps) / 1000) * rates.OperationsPer1000
	}
	if extraAI > 0 {
		total += int64(extraAI) * rates.AICreditsPerCredit
	}
	if extraStorageMB > 0 {
		total += (int64(extraStorageMB) / 1024) * rates.StoragePerGB
	}
	return total
}
