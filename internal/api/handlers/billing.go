package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type BillingHandler struct {
	billingSvc          *services.BillingService
	stripeWebhookSecret string
}

func NewBillingHandler(billingSvc *services.BillingService) *BillingHandler {
	return &BillingHandler{billingSvc: billingSvc}
}

func NewBillingHandlerWithWebhookSecret(billingSvc *services.BillingService, webhookSecret string) *BillingHandler {
	return &BillingHandler{billingSvc: billingSvc, stripeWebhookSecret: webhookSecret}
}

func (h *BillingHandler) GetPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.billingSvc.GetPlans(r.Context())
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get plans")
		return
	}

	response := []dto.PlanResponse{}
	for _, plan := range plans {
		response = append(response, dto.PlanResponse{
			ID:               plan.ID,
			Name:             plan.Name,
			Description:      plan.Description,
			PriceMonthly:     plan.PriceMonthly,
			PriceYearly:      plan.PriceYearly,
			ExecutionsLimit:  plan.ExecutionsLimit,
			WorkflowsLimit:   plan.WorkflowsLimit,
			MembersLimit:     plan.MembersLimit,
			CredentialsLimit: plan.CredentialsLimit,
			Features:         plan.Features,
		})
	}

	dto.JSON(w, http.StatusOK, response)
}

func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	subscription, err := h.billingSvc.GetSubscription(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "subscription not found")
		return
	}

	var cancelAt *int64
	if subscription.CancelAt != nil {
		ts := subscription.CancelAt.Unix()
		cancelAt = &ts
	}

	dto.JSON(w, http.StatusOK, dto.SubscriptionResponse{
		ID:                 subscription.ID.String(),
		PlanID:             subscription.PlanID,
		Status:             subscription.Status,
		BillingCycle:       subscription.BillingCycle,
		CurrentPeriodStart: subscription.CurrentPeriodStart.Unix(),
		CurrentPeriodEnd:   subscription.CurrentPeriodEnd.Unix(),
		CancelAt:           cancelAt,
	})
}

func (h *BillingHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	var req dto.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	subscription, err := h.billingSvc.CreateSubscription(r.Context(), services.CreateSubscriptionInput{
		WorkspaceID:  wsCtx.WorkspaceID,
		PlanID:       req.PlanID,
		BillingCycle: req.BillingCycle,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create subscription")
		return
	}

	dto.Created(w, dto.SubscriptionResponse{
		ID:                 subscription.ID.String(),
		PlanID:             subscription.PlanID,
		Status:             subscription.Status,
		BillingCycle:       subscription.BillingCycle,
		CurrentPeriodStart: subscription.CurrentPeriodStart.Unix(),
		CurrentPeriodEnd:   subscription.CurrentPeriodEnd.Unix(),
	})
}

func (h *BillingHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	cancelAtPeriodEnd := r.URL.Query().Get("at_period_end") == "true"

	if err := h.billingSvc.CancelSubscription(r.Context(), wsCtx.WorkspaceID, cancelAtPeriodEnd); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to cancel subscription")
		return
	}

	dto.JSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (h *BillingHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	usage, err := h.billingSvc.GetUsage(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get usage")
		return
	}

	dto.JSON(w, http.StatusOK, dto.UsageResponse{
		Executions:   usage.Executions,
		Workflows:    usage.Workflows,
		Members:      usage.Members,
		Credentials:  usage.Credentials,
		StorageBytes: usage.StorageBytes,
		PeriodStart:  usage.PeriodStart.Unix(),
		PeriodEnd:    usage.PeriodEnd.Unix(),
	})
}

func (h *BillingHandler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	opts := repositories.NewListOptions(page, perPage)

	invoices, total, err := h.billingSvc.GetInvoices(r.Context(), wsCtx.WorkspaceID, opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get invoices")
		return
	}

	response := []map[string]interface{}{}
	for _, inv := range invoices {
		response = append(response, map[string]interface{}{
			"id":          inv.ID.String(),
			"number":      inv.Number,
			"status":      inv.Status,
			"amount_due":  inv.AmountDue,
			"amount_paid": inv.AmountPaid,
			"currency":    inv.Currency,
			"invoice_url": inv.InvoiceURL,
			"created_at":  inv.CreatedAt.Unix(),
		})
	}

	totalPages := int(total) / opts.Limit
	if int(total)%opts.Limit > 0 {
		totalPages++
	}

	dto.JSONWithMeta(w, http.StatusOK, response, &dto.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *BillingHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "failed to read body")
		return
	}

	// Verify Stripe signature if webhook secret is configured
	if h.stripeWebhookSecret != "" {
		sigHeader := r.Header.Get("Stripe-Signature")
		if !h.verifyStripeSignature(body, sigHeader) {
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid signature")
			return
		}
	}

	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	eventType, _ := event["type"].(string)
	data, _ := event["data"].(map[string]interface{})

	if err := h.billingSvc.HandleStripeWebhook(r.Context(), eventType, data); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to handle webhook")
		return
	}

	dto.JSON(w, http.StatusOK, map[string]string{"received": "true"})
}

func (h *BillingHandler) verifyStripeSignature(payload []byte, sigHeader string) bool {
	if sigHeader == "" {
		return false
	}

	// Parse signature header
	// Format: t=timestamp,v1=signature
	var timestamp string
	var signature string
	
	parts := strings.Split(sigHeader, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			signature = kv[1]
		}
	}

	if timestamp == "" || signature == "" {
		return false
	}

	// Verify timestamp is not too old (5 minute tolerance)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix()-ts > 300 {
		return false
	}

	// Compute expected signature
	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(h.stripeWebhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}
