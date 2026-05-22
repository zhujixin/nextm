package tag

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nextm/nextm/internal/api/dto"
	"github.com/nextm/nextm/internal/model"
)

type Repository interface {
	ListTags(ctx context.Context, spaceID string) ([]*model.Tag, error)
	GetTag(ctx context.Context, id string) (*model.Tag, error)
	CreateTag(ctx context.Context, arg interface{}) (*model.Tag, error)
	UpdateTag(ctx context.Context, arg interface{}) (*model.Tag, error)
	DeleteTag(ctx context.Context, id string) error
	GetObjectTags(ctx context.Context, objectID string) ([]*model.Tag, error)
	AssignTags(ctx context.Context, objectID string, tagIDs []string) error
	UnassignTag(ctx context.Context, objectID, tagID string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, spaceID string) ([]*model.Tag, error) {
	return s.repo.ListTags(ctx, spaceID)
}

func (s *Service) Get(ctx context.Context, id string) (*model.Tag, error) {
	tag, err := s.repo.GetTag(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}
	return tag, nil
}

func (s *Service) Create(ctx context.Context, spaceID string, req dto.CreateTagRequest) (*model.Tag, error) {
	now := model.NowMS()
	tag, err := s.repo.CreateTag(ctx, struct {
		ID        string
		SpaceID   string
		Name      string
		Color     string
		ParentID  *string
		CreatedAt int64
		UpdatedAt int64
	}{
		ID: uuid.New().String(), SpaceID: spaceID,
		Name: req.Name, Color: req.Color,
		ParentID: req.ParentID,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}
	return tag, nil
}

func (s *Service) Update(ctx context.Context, id string, req dto.UpdateTagRequest) (*model.Tag, error) {
	existing, err := s.repo.GetTag(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	now := model.NowMS()
	tag, err := s.repo.UpdateTag(ctx, struct {
		ID        string
		SpaceID   string
		Name      string
		Color     string
		ParentID  *string
		UpdatedAt int64
	}{
		ID: id, SpaceID: existing.SpaceID,
		Name:      safeStr(req.Name, existing.Name),
		Color:     safeStr(req.Color, existing.Color),
		ParentID:  safeStrPtr(req.ParentID, existing.ParentID),
		UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("update tag: %w", err)
	}
	return tag, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteTag(ctx, id)
}

func (s *Service) GetObjectTags(ctx context.Context, objectID string) ([]*model.Tag, error) {
	return s.repo.GetObjectTags(ctx, objectID)
}

func (s *Service) AssignTags(ctx context.Context, objectID string, req dto.AssignTagsRequest) error {
	return s.repo.AssignTags(ctx, objectID, req.TagIDs)
}

func (s *Service) UnassignTag(ctx context.Context, objectID, tagID string) error {
	return s.repo.UnassignTag(ctx, objectID, tagID)
}

func safeStr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

func safeStrPtr(s *string, fallback *string) *string {
	if s != nil {
		return s
	}
	return fallback
}

var _ = time.Now
