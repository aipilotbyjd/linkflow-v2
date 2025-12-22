package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

type Client struct {
	Hub         *Hub
	Conn        *websocket.Conn
	Send        chan []byte
	UserID      uuid.UUID
	WorkspaceID *uuid.UUID
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, workspaceID *uuid.UUID) *Client {
	return &Client{
		Hub:         hub,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		UserID:      userID,
		WorkspaceID: workspaceID,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("WebSocket read error")
			}
			break
		}

		// Handle incoming messages (e.g., subscription to specific execution)
		c.handleMessage(message)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Batch queued messages
			n := len(c.Send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type SubscribePayload struct {
	ExecutionID string `json:"execution_id"`
	WorkflowID  string `json:"workflow_id"`
}

type UnsubscribePayload struct {
	ExecutionID string `json:"execution_id"`
	WorkflowID  string `json:"workflow_id"`
}

func (c *Client) handleMessage(message []byte) {
	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Warn().Err(err).Msg("Failed to parse WebSocket message")
		c.sendError("Invalid message format")
		return
	}

	switch msg.Type {
	case "subscribe":
		var payload SubscribePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			c.sendError("Invalid subscribe payload")
			return
		}
		c.handleSubscribe(payload)

	case "unsubscribe":
		var payload UnsubscribePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			c.sendError("Invalid unsubscribe payload")
			return
		}
		c.handleUnsubscribe(payload)

	case "ping":
		c.sendPong()

	default:
		c.sendError("Unknown message type: " + msg.Type)
	}
}

func (c *Client) handleSubscribe(payload SubscribePayload) {
	if payload.ExecutionID != "" {
		execID, err := uuid.Parse(payload.ExecutionID)
		if err != nil {
			c.sendError("Invalid execution ID")
			return
		}
		c.Hub.SubscribeToExecution(c, execID)
		c.sendAck("subscribed", map[string]string{"execution_id": payload.ExecutionID})
	}
	if payload.WorkflowID != "" {
		workflowID, err := uuid.Parse(payload.WorkflowID)
		if err != nil {
			c.sendError("Invalid workflow ID")
			return
		}
		c.Hub.SubscribeToWorkflow(c, workflowID)
		c.sendAck("subscribed", map[string]string{"workflow_id": payload.WorkflowID})
	}
}

func (c *Client) handleUnsubscribe(payload UnsubscribePayload) {
	if payload.ExecutionID != "" {
		execID, _ := uuid.Parse(payload.ExecutionID)
		c.Hub.UnsubscribeFromExecution(c, execID)
		c.sendAck("unsubscribed", map[string]string{"execution_id": payload.ExecutionID})
	}
	if payload.WorkflowID != "" {
		workflowID, _ := uuid.Parse(payload.WorkflowID)
		c.Hub.UnsubscribeFromWorkflow(c, workflowID)
		c.sendAck("unsubscribed", map[string]string{"workflow_id": payload.WorkflowID})
	}
}

func (c *Client) sendAck(msgType string, data interface{}) {
	response := map[string]interface{}{
		"type": msgType,
		"data": data,
	}
	jsonData, _ := json.Marshal(response)
	select {
	case c.Send <- jsonData:
	default:
		log.Warn().Msg("WebSocket send buffer full")
	}
}

func (c *Client) sendError(message string) {
	response := map[string]interface{}{
		"type":  "error",
		"error": message,
	}
	jsonData, _ := json.Marshal(response)
	select {
	case c.Send <- jsonData:
	default:
	}
}

func (c *Client) sendPong() {
	response := map[string]string{"type": "pong"}
	jsonData, _ := json.Marshal(response)
	select {
	case c.Send <- jsonData:
	default:
	}
}
