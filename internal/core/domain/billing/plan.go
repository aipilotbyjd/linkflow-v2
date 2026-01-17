package billing

type Plan struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	PriceMonthly  int64    `json:"price_monthly"`
	PriceYearly   int64    `json:"price_yearly"`
	Currency      string   `json:"currency"`
	Features      []string `json:"features"`
	Limits        Limits   `json:"limits"`
	StripePriceID string   `json:"stripe_price_id,omitempty"`
}

type Limits struct {
	Workflows           int `json:"workflows"`
	ExecutionsPerMonth  int `json:"executions_per_month"`
	TeamMembers         int `json:"team_members"`
	Credentials         int `json:"credentials"`
	RetentionDays       int `json:"retention_days"`
	WebhooksPerWorkflow int `json:"webhooks_per_workflow"`
}

var FreePlan = Plan{
	ID:           "free",
	Name:         "Free",
	Description:  "For individuals getting started",
	PriceMonthly: 0,
	PriceYearly:  0,
	Currency:     "USD",
	Features:     []string{"5 workflows", "1,000 executions/month", "Community support"},
	Limits: Limits{
		Workflows:           5,
		ExecutionsPerMonth:  1000,
		TeamMembers:         1,
		Credentials:         5,
		RetentionDays:       7,
		WebhooksPerWorkflow: 1,
	},
}

var ProPlan = Plan{
	ID:           "pro",
	Name:         "Pro",
	Description:  "For professionals and small teams",
	PriceMonthly: 2900,
	PriceYearly:  29000,
	Currency:     "USD",
	Features:     []string{"Unlimited workflows", "50,000 executions/month", "Priority support"},
	Limits: Limits{
		Workflows:           -1,
		ExecutionsPerMonth:  50000,
		TeamMembers:         10,
		Credentials:         50,
		RetentionDays:       30,
		WebhooksPerWorkflow: 10,
	},
}

var EnterprisePlan = Plan{
	ID:           "enterprise",
	Name:         "Enterprise",
	Description:  "For large organizations",
	PriceMonthly: 0,
	PriceYearly:  0,
	Currency:     "USD",
	Features:     []string{"Unlimited everything", "SSO", "Dedicated support", "SLA"},
	Limits: Limits{
		Workflows:           -1,
		ExecutionsPerMonth:  -1,
		TeamMembers:         -1,
		Credentials:         -1,
		RetentionDays:       365,
		WebhooksPerWorkflow: -1,
	},
}

func GetPlan(id string) *Plan {
	switch id {
	case "free":
		return &FreePlan
	case "pro":
		return &ProPlan
	case "enterprise":
		return &EnterprisePlan
	default:
		return &FreePlan
	}
}

func GetAllPlans() []Plan {
	return []Plan{FreePlan, ProPlan, EnterprisePlan}
}
