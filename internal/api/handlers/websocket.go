package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	ws "github.com/linkflow-ai/linkflow/internal/api/websocket"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/rs/zerolog/log"
)

type WebSocketHandler struct {
	hub            *ws.Hub
	jwtManager     *crypto.JWTManager
	allowedOrigins []string
	upgrader       websocket.Upgrader
}

func NewWebSocketHandler(hub *ws.Hub, jwtManager *crypto.JWTManager) *WebSocketHandler {
	h := &WebSocketHandler{
		hub:            hub,
		jwtManager:     jwtManager,
		allowedOrigins: []string{}, // Empty = allow all (for dev)
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	return h
}

func NewWebSocketHandlerWithOrigins(hub *ws.Hub, jwtManager *crypto.JWTManager, allowedOrigins []string) *WebSocketHandler {
	h := &WebSocketHandler{
		hub:            hub,
		jwtManager:     jwtManager,
		allowedOrigins: allowedOrigins,
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	return h
}

func (h *WebSocketHandler) checkOrigin(r *http.Request) bool {
	// If no origins configured, allow all (dev mode)
	if len(h.allowedOrigins) == 0 {
		return true
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		// Allow requests without origin header (same-origin requests)
		return true
	}

	// Parse origin URL
	parsedOrigin, err := url.Parse(origin)
	if err != nil {
		log.Warn().Str("origin", origin).Msg("Invalid origin URL")
		return false
	}

	// Check against allowed origins
	originHost := parsedOrigin.Host
	for _, allowed := range h.allowedOrigins {
		// Exact match
		if allowed == origin || allowed == originHost {
			return true
		}
		// Wildcard subdomain match (e.g., "*.example.com")
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:]
			if strings.HasSuffix(originHost, domain) || originHost == domain[1:] {
				return true
			}
		}
	}

	log.Warn().Str("origin", origin).Strs("allowed", h.allowedOrigins).Msg("WebSocket origin not allowed")
	return false
}

func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Get token from query param
	token := r.URL.Query().Get("token")
	if token == "" {
		dto.ErrorResponse(w, http.StatusUnauthorized, "missing token")
		return
	}

	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "invalid token")
		return
	}

	// Get workspace ID from query param or JWT
	var workspaceID *uuid.UUID
	if claims.WorkspaceID != nil {
		workspaceID = claims.WorkspaceID
	} else if wsIDStr := r.URL.Query().Get("workspace_id"); wsIDStr != "" {
		wsID, err := uuid.Parse(wsIDStr)
		if err == nil {
			workspaceID = &wsID
		}
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	client := ws.NewClient(h.hub, conn, claims.UserID, workspaceID)
	h.hub.Register(client)

	log.Debug().
		Str("user_id", claims.UserID.String()).
		Bool("has_workspace", workspaceID != nil).
		Msg("WebSocket client registered")

	go client.WritePump()
	go client.ReadPump()
}
