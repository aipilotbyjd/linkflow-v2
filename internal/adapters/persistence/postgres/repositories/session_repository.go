package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *user.Session) error {
	model := sessionToModel(s)
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *SessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.Session, error) {
	var model models.UserSession
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return sessionToDomain(&model), nil
}

func (r *SessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*user.Session, error) {
	var model models.UserSession
	if err := postgres.GetTx(ctx, r.db).First(&model, "refresh_token = ?", tokenHash).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return sessionToDomain(&model), nil
}

func (r *SessionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]user.Session, error) {
	var models []models.UserSession
	if err := postgres.GetTx(ctx, r.db).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}

	sessions := make([]user.Session, len(models))
	for i, m := range models {
		sessions[i] = *sessionToDomain(&m)
	}
	return sessions, nil
}

func (r *SessionRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now()
	return postgres.GetTx(ctx, r.db).Model(&models.UserSession{}).
		Where("id = ?", sessionID).
		Update("revoked_at", now).Error
}

func (r *SessionRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return postgres.GetTx(ctx, r.db).Model(&models.UserSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&models.UserSession{}, "user_id = ?", userID).Error
}

func (r *SessionRepository) UpdateLastUsed(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now()
	return postgres.GetTx(ctx, r.db).Model(&models.UserSession{}).
		Where("id = ?", sessionID).
		Update("last_used_at", now).Error
}

func (r *SessionRepository) CleanupExpired(ctx context.Context) (int64, error) {
	result := postgres.GetTx(ctx, r.db).
		Where("expires_at < ?", time.Now()).
		Delete(&models.UserSession{})
	return result.RowsAffected, result.Error
}

// Mapper functions
func sessionToModel(s *user.Session) *models.UserSession {
	m := &models.UserSession{
		ID:           s.ID,
		UserID:       s.UserID,
		RefreshToken: s.TokenHash,
		UserAgent:    s.UserAgent,
		IPAddress:    s.IPAddress,
		ExpiresAt:    s.ExpiresAt,
		RevokedAt:    s.RevokedAt,
		CreatedAt:    s.CreatedAt,
	}
	if !s.LastUsedAt.IsZero() {
		m.LastUsedAt = &s.LastUsedAt
	}
	return m
}

func sessionToDomain(m *models.UserSession) *user.Session {
	s := &user.Session{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.RefreshToken,
		UserAgent: m.UserAgent,
		IPAddress: m.IPAddress,
		ExpiresAt: m.ExpiresAt,
		RevokedAt: m.RevokedAt,
		CreatedAt: m.CreatedAt,
	}
	if m.LastUsedAt != nil {
		s.LastUsedAt = *m.LastUsedAt
	}
	return s
}
