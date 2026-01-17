package credential

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	credentialQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/credential"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// ListHandler handles listing credentials
type ListHandler struct {
	handler *credentialQuery.ListCredentialsHandler
}

// NewListHandler creates a new handler
func NewListHandler(handler *credentialQuery.ListCredentialsHandler) *ListHandler {
	return &ListHandler{handler: handler}
}

// Handle handles the list credentials request
func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	query := r.URL.Query()

	page := 1
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := types.DefaultPageSize
	if ps := query.Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	var credType *credential.Type
	if t := query.Get("type"); t != "" {
		parsed := credential.Type(t)
		credType = &parsed
	}

	search := query.Get("search")

	result, err := h.handler.Handle(r.Context(), credentialQuery.ListCredentialsQuery{
		WorkspaceID: wsCtx.WorkspaceID,
		Type:        credType,
		Search:      search,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	credentials := make([]CredentialResponse, len(result.Credentials))
	for i, c := range result.Credentials {
		credentials[i] = ToCredentialResponse(&c)
	}

	common.List(w, credentials, types.PageResponse{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalItems: result.Total,
		TotalPages: result.TotalPages,
		HasMore:    result.Page < result.TotalPages,
	})
}
