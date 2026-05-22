package relation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nextm/nextm/internal/api/dto"
	"github.com/nextm/nextm/internal/model"
)

type Repository interface {
	ListRelationsBySource(ctx context.Context, sourceID string) ([]*model.Relation, error)
	ListRelationsByTarget(ctx context.Context, targetID string) ([]*model.Relation, error)
	ListRelationsByObject(ctx context.Context, objectID string) ([]*model.Relation, error)
	GetRelation(ctx context.Context, id string) (*model.Relation, error)
	CreateRelation(ctx context.Context, arg interface{}) (*model.Relation, error)
	UpdateRelation(ctx context.Context, arg interface{}) (*model.Relation, error)
	DeleteRelation(ctx context.Context, id string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListByObject(ctx context.Context, objectID string) ([]*model.Relation, error) {
	return s.repo.ListRelationsByObject(ctx, objectID)
}

func (s *Service) Create(ctx context.Context, req dto.CreateRelationRequest) (*model.Relation, error) {
	now := model.NowMS()
	rel, err := s.repo.CreateRelation(ctx, struct {
		ID                string
		SourceID          string
		TargetID          string
		Type              string
		CustomTypeID      *string
		Weight           float64
		Metadata         string
		AIGenerated      int
		SourceObjectType string
		TargetObjectType string
		CreatedAt        int64
	}{
		ID: uuid.New().String(),
		SourceID: req.SourceID, TargetID: req.TargetID,
		Type: req.Type, CustomTypeID: req.CustomTypeID,
		Weight: req.Weight, Metadata: "{}",
		CreatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create relation: %w", err)
	}
	return rel, nil
}

func (s *Service) Update(ctx context.Context, id string, req dto.UpdateRelationRequest) (*model.Relation, error) {
	existing, err := s.repo.GetRelation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("relation not found: %w", err)
	}

	now := model.NowMS()
	rel, err := s.repo.UpdateRelation(ctx, struct {
		ID           string
		Type         string
		CustomTypeID *string
		Weight      float64
		Metadata    string
		UpdatedAt   int64
	}{
		ID: id,
		Type:         safeStr(req.Type, existing.Type),
		CustomTypeID: safeStrPtr(req.CustomTypeID, existing.CustomTypeID),
		Weight:       safeFloat(req.Weight, existing.Weight),
		Metadata:     safeStr(req.Metadata, existing.Metadata),
		UpdatedAt:    now,
	})
	if err != nil {
		return nil, fmt.Errorf("update relation: %w", err)
	}
	return rel, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteRelation(ctx, id)
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

func safeFloat(f *float64, fallback float64) float64 {
	if f == nil {
		return fallback
	}
	return *f
}

var _ = time.Now
