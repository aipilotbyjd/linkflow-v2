package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/rs/zerolog/log"
)

// Note errors
var (
	ErrNoteNotFound        = errors.New("note not found")
	ErrNoteContentRequired = errors.New("note content is required")
	ErrNoteResourceRequired = errors.New("resource_id and resource_name are required")
	ErrNoteAccessDenied    = errors.New("access denied to note")
)

// NoteService handles note management
type NoteService struct {
	noteRepo *repositories.NoteRepository
}

// NewNoteService creates a new NoteService
func NewNoteService(noteRepo *repositories.NoteRepository) *NoteService {
	if noteRepo == nil {
		panic("note service: noteRepo is required")
	}
	return &NoteService{
		noteRepo: noteRepo,
	}
}

type CreateNoteInput struct {
	WorkspaceID  uuid.UUID
	ResourceID   uuid.UUID
	ResourceName string
	CreatedBy    uuid.UUID
	Content      string
	Position     models.JSON
	Size         models.JSON
	Color        string
}

// Create creates a new note
func (s *NoteService) Create(ctx context.Context, input CreateNoteInput) (*models.Note, error) {
	if input.Content == "" {
		return nil, ErrNoteContentRequired
	}
	if input.ResourceID == uuid.Nil || input.ResourceName == "" {
		return nil, ErrNoteResourceRequired
	}

	// Default color
	color := input.Color
	if color == "" {
		color = "yellow"
	}

	// Default position
	position := input.Position
	if position == nil {
		position = models.JSON{"x": 0, "y": 0}
	}

	// Default size
	size := input.Size
	if size == nil {
		size = models.JSON{"width": 200, "height": 150}
	}

	note := &models.Note{
		WorkspaceID:  input.WorkspaceID,
		ResourceID:   input.ResourceID,
		ResourceName: input.ResourceName,
		CreatedBy:    input.CreatedBy,
		Content:      input.Content,
		Position:     position,
		Size:         size,
		Color:        color,
	}

	if err := s.noteRepo.Create(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	log.Info().
		Str("note_id", note.ID.String()).
		Str("workspace_id", input.WorkspaceID.String()).
		Str("resource_id", input.ResourceID.String()).
		Str("resource_name", input.ResourceName).
		Msg("Note created")

	return note, nil
}

// GetByID returns a note by ID
func (s *NoteService) GetByID(ctx context.Context, id uuid.UUID) (*models.Note, error) {
	note, err := s.noteRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNoteNotFound, id)
	}
	return note, nil
}

// GetByWorkspace returns paginated notes for a workspace
func (s *NoteService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Note, int64, error) {
	notes, total, err := s.noteRepo.FindByWorkspaceID(ctx, workspaceID, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notes: %w", err)
	}
	return notes, total, nil
}

// GetByResource returns notes for a specific resource
func (s *NoteService) GetByResource(ctx context.Context, workspaceID, resourceID uuid.UUID, resourceName string, opts *repositories.ListOptions) ([]models.Note, int64, error) {
	notes, total, err := s.noteRepo.FindByResource(ctx, workspaceID, resourceID, resourceName, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notes: %w", err)
	}
	return notes, total, nil
}

// GetWithFilters returns paginated notes with filters
func (s *NoteService) GetWithFilters(ctx context.Context, workspaceID uuid.UUID, filter *repositories.NoteFilter, opts *repositories.ListOptions) ([]models.Note, int64, error) {
	notes, total, err := s.noteRepo.FindWithFilters(ctx, workspaceID, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notes: %w", err)
	}
	return notes, total, nil
}

type UpdateNoteInput struct {
	Content  *string
	Position *models.JSON
	Size     *models.JSON
	Color    *string
}

// Update updates an existing note
func (s *NoteService) Update(ctx context.Context, id uuid.UUID, input UpdateNoteInput) (*models.Note, error) {
	note, err := s.noteRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNoteNotFound, id)
	}

	if input.Content != nil {
		note.Content = *input.Content
	}
	if input.Position != nil {
		note.Position = *input.Position
	}
	if input.Size != nil {
		note.Size = *input.Size
	}
	if input.Color != nil {
		note.Color = *input.Color
	}

	if err := s.noteRepo.Update(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	log.Info().
		Str("note_id", note.ID.String()).
		Msg("Note updated")

	return note, nil
}

// Delete deletes a note
func (s *NoteService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.noteRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	log.Info().
		Str("note_id", id.String()).
		Msg("Note deleted")

	return nil
}

// DeleteByResource deletes all notes for a resource
func (s *NoteService) DeleteByResource(ctx context.Context, workspaceID, resourceID uuid.UUID, resourceName string) error {
	if err := s.noteRepo.DeleteByResource(ctx, workspaceID, resourceID, resourceName); err != nil {
		return fmt.Errorf("failed to delete notes: %w", err)
	}

	log.Info().
		Str("resource_id", resourceID.String()).
		Str("resource_name", resourceName).
		Msg("Notes deleted for resource")

	return nil
}

// CountByResource returns count of notes for a resource
func (s *NoteService) CountByResource(ctx context.Context, workspaceID, resourceID uuid.UUID, resourceName string) (int64, error) {
	return s.noteRepo.CountByResource(ctx, workspaceID, resourceID, resourceName)
}
