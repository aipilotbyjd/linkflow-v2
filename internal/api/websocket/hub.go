package websocket

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Hub struct {
	clients    map[*Client]bool
	userConns  map[uuid.UUID]map[*Client]bool
	wsConns    map[uuid.UUID]map[*Client]bool // workspace connections
	execConns  map[uuid.UUID]map[*Client]bool // execution subscriptions
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		userConns:  make(map[uuid.UUID]map[*Client]bool),
		wsConns:    make(map[uuid.UUID]map[*Client]bool),
		execConns:  make(map[uuid.UUID]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true

			// Add to user connections
			if _, ok := h.userConns[client.UserID]; !ok {
				h.userConns[client.UserID] = make(map[*Client]bool)
			}
			h.userConns[client.UserID][client] = true

			// Add to workspace connections
			if client.WorkspaceID != nil {
				if _, ok := h.wsConns[*client.WorkspaceID]; !ok {
					h.wsConns[*client.WorkspaceID] = make(map[*Client]bool)
				}
				h.wsConns[*client.WorkspaceID][client] = true
			}
			h.mu.Unlock()

			log.Debug().
				Str("user_id", client.UserID.String()).
				Msg("WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)

				// Remove from user connections
				if conns, ok := h.userConns[client.UserID]; ok {
					delete(conns, client)
					if len(conns) == 0 {
						delete(h.userConns, client.UserID)
					}
				}

				// Remove from workspace connections
				if client.WorkspaceID != nil {
					if conns, ok := h.wsConns[*client.WorkspaceID]; ok {
						delete(conns, client)
						if len(conns) == 0 {
							delete(h.wsConns, *client.WorkspaceID)
						}
					}
				}
			}
			h.mu.Unlock()

			log.Debug().
				Str("user_id", client.UserID.String()).
				Msg("WebSocket client disconnected")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

func (h *Hub) SendToUser(userID uuid.UUID, event *Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if conns, ok := h.userConns[userID]; ok {
		for client := range conns {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

func (h *Hub) SendToWorkspace(workspaceID uuid.UUID, event *Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if conns, ok := h.wsConns[workspaceID]; ok {
		for client := range conns {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

func (h *Hub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) BroadcastToWorkspace(workspaceID uuid.UUID, event Event) {
	h.SendToWorkspace(workspaceID, &event)
}

func (h *Hub) BroadcastToExecution(executionID uuid.UUID, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if conns, ok := h.execConns[executionID]; ok {
		for client := range conns {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

func (h *Hub) SubscribeToExecution(client *Client, executionID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.execConns[executionID]; !ok {
		h.execConns[executionID] = make(map[*Client]bool)
	}
	h.execConns[executionID][client] = true
}

func (h *Hub) UnsubscribeFromExecution(client *Client, executionID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.execConns[executionID]; ok {
		delete(conns, client)
		if len(conns) == 0 {
			delete(h.execConns, executionID)
		}
	}
}

func (h *Hub) CleanupExecutionSubscriptions(executionID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.execConns, executionID)
}
