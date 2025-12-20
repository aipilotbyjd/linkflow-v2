package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	ws "github.com/linkflow-ai/linkflow/internal/api/websocket"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Add proper origin validation
		return true
	},
}

type WebSocketHandler struct {
	hub        *ws.Hub
	jwtManager *crypto.JWTManager
}

func NewWebSocketHandler(hub *ws.Hub, jwtManager *crypto.JWTManager) *WebSocketHandler {
	return &WebSocketHandler{
		hub:        hub,
		jwtManager: jwtManager,
	}
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

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	client := ws.NewClient(h.hub, conn, claims.UserID, claims.WorkspaceID)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
