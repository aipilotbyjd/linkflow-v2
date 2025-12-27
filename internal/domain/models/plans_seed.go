package models

import "encoding/json"

// DefaultPlans returns the default pricing plans (Make.com style)
func DefaultPlans() []Plan {
	return []Plan{
		{
			ID:   PlanFree,
			Name: "Free",
			Tier: "free",

			PriceMonthly: 0,
			PriceYearly:  0,

			CreditsIncluded:   1000,
			CreditsMax:        1000,
			CreditOverageCost: 0, // No overage on free

			ExecutionsLimit:     100,
			WorkflowsLimit:      5,
			MembersLimit:        1,
			CredentialsLimit:    5,
			SchedulesLimit:      2,
			WebhooksLimit:       2,
			ExecutionTimeout:    30,
			MaxNodesPerWorkflow: 20,
			RetentionDays:       7,
			LogRetentionDays:    3,

			Features:    mustJSON(FreePlanFeatures()),
			Description: strPtr("Perfect for getting started with automation"),
			IsActive:    true,
			IsPublic:    true,
			SortOrder:   1,
		},
		{
			ID:   PlanStarter,
			Name: "Starter",
			Tier: "starter",

			PriceMonthly: 900,   // $9
			PriceYearly:  8640,  // $86.40 (20% off)

			CreditsIncluded:   10000,
			CreditsMax:        50000,
			CreditOverageCost: 225, // $2.25 per 1000 extra credits (25% markup)

			ExecutionsLimit:     1000,
			WorkflowsLimit:      20,
			MembersLimit:        3,
			CredentialsLimit:    20,
			SchedulesLimit:      10,
			WebhooksLimit:       10,
			ExecutionTimeout:    60,
			MaxNodesPerWorkflow: 50,
			RetentionDays:       30,
			LogRetentionDays:    7,

			Features:    mustJSON(StarterPlanFeatures()),
			Description: strPtr("For freelancers and solopreneurs"),
			IsActive:    true,
			IsPublic:    true,
			SortOrder:   2,
		},
		{
			ID:   PlanPro,
			Name: "Pro",
			Tier: "pro",

			PriceMonthly: 2900,   // $29
			PriceYearly:  27840,  // $278.40 (20% off)

			CreditsIncluded:   50000,
			CreditsMax:        500000,
			CreditOverageCost: 175, // $1.75 per 1000 extra credits (25% markup)

			ExecutionsLimit:     10000,
			WorkflowsLimit:      100,
			MembersLimit:        10,
			CredentialsLimit:    100,
			SchedulesLimit:      50,
			WebhooksLimit:       50,
			ExecutionTimeout:    300, // 5 minutes
			MaxNodesPerWorkflow: 100,
			RetentionDays:       90,
			LogRetentionDays:    30,

			Features:    mustJSON(ProPlanFeatures()),
			Description: strPtr("For growing businesses"),
			IsActive:    true,
			IsPublic:    true,
			SortOrder:   3,
		},
		{
			ID:   PlanBusiness,
			Name: "Business",
			Tier: "business",

			PriceMonthly: 9900,   // $99
			PriceYearly:  95040,  // $950.40 (20% off)

			CreditsIncluded:   200000,
			CreditsMax:        2000000,
			CreditOverageCost: 125, // $1.25 per 1000 extra credits (25% markup)

			ExecutionsLimit:     100000,
			WorkflowsLimit:      500,
			MembersLimit:        50,
			CredentialsLimit:    500,
			SchedulesLimit:      200,
			WebhooksLimit:       200,
			ExecutionTimeout:    1800, // 30 minutes
			MaxNodesPerWorkflow: 200,
			RetentionDays:       365,
			LogRetentionDays:    90,

			Features:    mustJSON(BusinessPlanFeatures()),
			Description: strPtr("For teams with advanced needs"),
			IsActive:    true,
			IsPublic:    true,
			SortOrder:   4,
		},
		{
			ID:   PlanEnterprise,
			Name: "Enterprise",
			Tier: "enterprise",

			PriceMonthly: 0, // Custom pricing
			PriceYearly:  0,

			CreditsIncluded:   -1, // Unlimited
			CreditsMax:        -1,
			CreditOverageCost: 0,

			ExecutionsLimit:     -1, // Unlimited
			WorkflowsLimit:      -1,
			MembersLimit:        -1,
			CredentialsLimit:    -1,
			SchedulesLimit:      -1,
			WebhooksLimit:       -1,
			ExecutionTimeout:    3600, // 1 hour
			MaxNodesPerWorkflow: -1,
			RetentionDays:       -1, // Custom
			LogRetentionDays:    365,

			Features:    mustJSON(EnterprisePlanFeatures()),
			Description: strPtr("Custom solutions for large organizations"),
			IsActive:    true,
			IsPublic:    true,
			SortOrder:   5,
		},
	}
}

// Feature configurations for each plan

func FreePlanFeatures() PlanFeatures {
	return PlanFeatures{
		Webhooks:          true,
		Schedules:         true,
		ManualTrigger:     true,
		BasicNodes:        true,
		AdvancedNodes:     false,
		SubWorkflows:      false,
		ErrorWorkflow:     false,
		APIAccess:         false,
		CustomFunctions:   false,
		CustomAI:          false,
		PriorityExecution: false,
		ParallelExecution: false,
		RetryOnFailure:    false,
		TeamRoles:         false,
		SharedTemplates:   false,
		WorkflowComments:  false,
		SSO:               false,
		AuditLogs:         false,
		IPWhitelist:       false,
		DataEncryption:    false,
		PrioritySupport:   false,
		DedicatedSupport:  false,
		SLAGuarantee:      false,
		CustomBranding:    false,
		WhiteLabel:        false,
	}
}

func StarterPlanFeatures() PlanFeatures {
	return PlanFeatures{
		Webhooks:          true,
		Schedules:         true,
		ManualTrigger:     true,
		BasicNodes:        true,
		AdvancedNodes:     true,
		SubWorkflows:      true,
		ErrorWorkflow:     true,
		APIAccess:         true,
		CustomFunctions:   false,
		CustomAI:          true,
		PriorityExecution: false,
		ParallelExecution: false,
		RetryOnFailure:    true,
		TeamRoles:         false,
		SharedTemplates:   false,
		WorkflowComments:  false,
		SSO:               false,
		AuditLogs:         false,
		IPWhitelist:       false,
		DataEncryption:    true,
		PrioritySupport:   false,
		DedicatedSupport:  false,
		SLAGuarantee:      false,
		CustomBranding:    false,
		WhiteLabel:        false,
	}
}

func ProPlanFeatures() PlanFeatures {
	return PlanFeatures{
		Webhooks:          true,
		Schedules:         true,
		ManualTrigger:     true,
		BasicNodes:        true,
		AdvancedNodes:     true,
		SubWorkflows:      true,
		ErrorWorkflow:     true,
		APIAccess:         true,
		CustomFunctions:   true,
		CustomAI:          true,
		PriorityExecution: true,
		ParallelExecution: true,
		RetryOnFailure:    true,
		TeamRoles:         true,
		SharedTemplates:   true,
		WorkflowComments:  true,
		SSO:               false,
		AuditLogs:         true,
		IPWhitelist:       false,
		DataEncryption:    true,
		PrioritySupport:   true,
		DedicatedSupport:  false,
		SLAGuarantee:      false,
		CustomBranding:    false,
		WhiteLabel:        false,
	}
}

func BusinessPlanFeatures() PlanFeatures {
	return PlanFeatures{
		Webhooks:          true,
		Schedules:         true,
		ManualTrigger:     true,
		BasicNodes:        true,
		AdvancedNodes:     true,
		SubWorkflows:      true,
		ErrorWorkflow:     true,
		APIAccess:         true,
		CustomFunctions:   true,
		CustomAI:          true,
		PriorityExecution: true,
		ParallelExecution: true,
		RetryOnFailure:    true,
		TeamRoles:         true,
		SharedTemplates:   true,
		WorkflowComments:  true,
		SSO:               true,
		AuditLogs:         true,
		IPWhitelist:       true,
		DataEncryption:    true,
		PrioritySupport:   true,
		DedicatedSupport:  false,
		SLAGuarantee:      true,
		CustomBranding:    true,
		WhiteLabel:        false,
	}
}

func EnterprisePlanFeatures() PlanFeatures {
	return PlanFeatures{
		Webhooks:          true,
		Schedules:         true,
		ManualTrigger:     true,
		BasicNodes:        true,
		AdvancedNodes:     true,
		SubWorkflows:      true,
		ErrorWorkflow:     true,
		APIAccess:         true,
		CustomFunctions:   true,
		CustomAI:          true,
		PriorityExecution: true,
		ParallelExecution: true,
		RetryOnFailure:    true,
		TeamRoles:         true,
		SharedTemplates:   true,
		WorkflowComments:  true,
		SSO:               true,
		AuditLogs:         true,
		IPWhitelist:       true,
		DataEncryption:    true,
		PrioritySupport:   true,
		DedicatedSupport:  true,
		SLAGuarantee:      true,
		CustomBranding:    true,
		WhiteLabel:        true,
	}
}

// Helper functions
func mustJSON(v interface{}) JSON {
	b, _ := json.Marshal(v)
	var result JSON
	_ = json.Unmarshal(b, &result)
	return result
}

func strPtr(s string) *string {
	return &s
}

// CreditCosts defines how many credits each operation type consumes
var CreditCosts = map[string]int{
	// Triggers (usually free or low cost)
	"trigger.manual":   0,
	"trigger.webhook":  1,
	"trigger.schedule": 1,

	// Actions
	"action.http":        1,
	"action.code":        2,
	"action.set":         1,
	"action.respond":     1,
	"action.subworkflow": 5,

	// Logic (usually low cost)
	"logic.if":        1,
	"logic.switch":    1,
	"logic.loop":      1, // Per iteration
	"logic.merge":     1,
	"logic.filter":    1,
	"logic.sort":      1,
	"logic.aggregate": 1,
	"logic.wait":      1,

	// Integrations (higher cost)
	"integration.slack":      2,
	"integration.email":      2,
	"integration.openai":     5,
	"integration.anthropic":  5,
	"integration.github":     2,
	"integration.postgres":   2,
	"integration.mysql":      2,
	"integration.mongodb":    2,
	"integration.redis":      1,
	"integration.aws_s3":     2,
	"integration.google":     2,
	"integration.twilio":     3,
	"integration.sendgrid":   2,
	"integration.stripe":     2,
	"integration.jira":       2,
	"integration.salesforce": 3,
	"integration.notion":     2,
	"integration.airtable":   2,
	"integration.discord":    2,
	"integration.telegram":   2,
	"integration.graphql":    2,
	"integration.ftp":        2,
	"integration.sftp":       2,
}

// GetCreditCost returns the credit cost for a node type
func GetCreditCost(nodeType string) int {
	if cost, ok := CreditCosts[nodeType]; ok {
		return cost
	}
	// Default cost for unknown node types
	return 1
}
