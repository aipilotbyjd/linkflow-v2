package websocket

import (
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	clients    map[*Client]bool
	workspaces map[uuid.UUID]map[*Client]bool
	users      map[uuid.UUID]map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		workspaces: make(map[uuid.UUID]map[*Client]bool),
		users:      make(map[uuid.UUID]map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	// Register to workspace
	if client.WorkspaceID != uuid.Nil {
		if h.workspaces[client.WorkspaceID] == nil {
			h.workspaces[client.WorkspaceID] = make(map[*Client]bool)
		}
		h.workspaces[client.WorkspaceID][client] = true
	}

	// Register to user
	if client.UserID != uuid.Nil {
		if h.users[client.UserID] == nil {
			h.users[client.UserID] = make(map[*Client]bool)
		}
		h.users[client.UserID][client] = true
	}
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}

	// Unregister from workspace
	if client.WorkspaceID != uuid.Nil {
		if clients, ok := h.workspaces[client.WorkspaceID]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.workspaces, client.WorkspaceID)
			}
		}
	}

	// Unregister from user
	if client.UserID != uuid.Nil {
		if clients, ok := h.users[client.UserID]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.users, client.UserID)
			}
		}
	}
}

func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var targets map[*Client]bool

	switch {
	case message.WorkspaceID != uuid.Nil:
		targets = h.workspaces[message.WorkspaceID]
	case message.UserID != uuid.Nil:
		targets = h.users[message.UserID]
	default:
		targets = h.clients
	}

	for client := range targets {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(message *Message) {
	h.broadcast <- message
}

func (h *Hub) BroadcastToWorkspace(workspaceID uuid.UUID, event string, data interface{}) {
	h.broadcast <- &Message{
		WorkspaceID: workspaceID,
		Event:       event,
		Data:        data,
	}
}

func (h *Hub) BroadcastToUser(userID uuid.UUID, event string, data interface{}) {
	h.broadcast <- &Message{
		UserID: userID,
		Event:  event,
		Data:   data,
	}
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) WorkspaceClientCount(workspaceID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.workspaces[workspaceID])
}
