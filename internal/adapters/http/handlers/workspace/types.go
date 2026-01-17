package workspace

import (
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

// MemberResponse represents member in responses
type MemberResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	WorkspaceID string `json:"workspace_id"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joined_at"`
}

// ToMemberResponse converts a domain member to response
func ToMemberResponse(m *workspace.Member) MemberResponse {
	return MemberResponse{
		ID:          m.ID.String(),
		UserID:      m.UserID.String(),
		WorkspaceID: m.WorkspaceID.String(),
		Role:        string(m.Role),
		JoinedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
