package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Session represents a user session
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	IP           string    `json:"ip"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// SessionStore handles session storage in Redis
type SessionStore struct {
	client *Client
	prefix string
	ttl    time.Duration
}

// NewSessionStore creates a new session store
func NewSessionStore(client *Client, ttl time.Duration) *SessionStore {
	return &SessionStore{
		client: client,
		prefix: "session:",
		ttl:    ttl,
	}
}

// Create creates a new session
func (s *SessionStore) Create(ctx context.Context, userID uuid.UUID, refreshToken, userAgent, ip string) (*Session, error) {
	session := &Session{
		ID:           uuid.New().String(),
		UserID:       userID.String(),
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IP:           ip,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(s.ttl),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	// Store by session ID
	if err := s.client.Set(ctx, s.prefix+session.ID, string(data), s.ttl); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Store mapping from refresh token to session ID
	if err := s.client.Set(ctx, s.prefix+"token:"+refreshToken, session.ID, s.ttl); err != nil {
		return nil, fmt.Errorf("failed to store token mapping: %w", err)
	}

	// Add to user's session list
	userSessionsKey := s.prefix + "user:" + userID.String()
	if err := s.client.client.SAdd(ctx, userSessionsKey, session.ID).Err(); err != nil {
		return nil, fmt.Errorf("failed to add to user sessions: %w", err)
	}
	s.client.client.Expire(ctx, userSessionsKey, s.ttl)

	return session, nil
}

// Get retrieves a session by ID
func (s *SessionStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	data, err := s.client.Get(ctx, s.prefix+sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// GetByRefreshToken retrieves a session by refresh token
func (s *SessionStore) GetByRefreshToken(ctx context.Context, refreshToken string) (*Session, error) {
	sessionID, err := s.client.Get(ctx, s.prefix+"token:"+refreshToken)
	if err != nil {
		return nil, fmt.Errorf("session not found for token: %w", err)
	}

	return s.Get(ctx, sessionID)
}

// Delete deletes a session
func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	// Delete session
	if err := s.client.Del(ctx, s.prefix+sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Delete token mapping
	if err := s.client.Del(ctx, s.prefix+"token:"+session.RefreshToken); err != nil {
		return fmt.Errorf("failed to delete token mapping: %w", err)
	}

	// Remove from user's session list
	userSessionsKey := s.prefix + "user:" + session.UserID
	s.client.client.SRem(ctx, userSessionsKey, sessionID)

	return nil
}

// DeleteByRefreshToken deletes a session by refresh token
func (s *SessionStore) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	sessionID, err := s.client.Get(ctx, s.prefix+"token:"+refreshToken)
	if err != nil {
		return nil // Session already deleted or expired
	}

	return s.Delete(ctx, sessionID)
}

// DeleteAllForUser deletes all sessions for a user
func (s *SessionStore) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	userSessionsKey := s.prefix + "user:" + userID.String()

	sessionIDs, err := s.client.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	for _, sessionID := range sessionIDs {
		s.Delete(ctx, sessionID)
	}

	return s.client.Del(ctx, userSessionsKey)
}

// Refresh refreshes a session with a new token
func (s *SessionStore) Refresh(ctx context.Context, oldToken, newToken string) (*Session, error) {
	session, err := s.GetByRefreshToken(ctx, oldToken)
	if err != nil {
		return nil, err
	}

	// Delete old token mapping
	s.client.Del(ctx, s.prefix+"token:"+oldToken)

	// Update session
	session.RefreshToken = newToken
	session.ExpiresAt = time.Now().Add(s.ttl)

	data, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	// Store updated session
	if err := s.client.Set(ctx, s.prefix+session.ID, string(data), s.ttl); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Store new token mapping
	if err := s.client.Set(ctx, s.prefix+"token:"+newToken, session.ID, s.ttl); err != nil {
		return nil, fmt.Errorf("failed to store token mapping: %w", err)
	}

	return session, nil
}

// GetUserSessions returns all sessions for a user
func (s *SessionStore) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*Session, error) {
	userSessionsKey := s.prefix + "user:" + userID.String()

	sessionIDs, err := s.client.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	sessions := make([]*Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session, err := s.Get(ctx, sessionID)
		if err != nil {
			continue // Skip expired/deleted sessions
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}
