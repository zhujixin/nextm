package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nextm/nextm/internal/api/dto"
	"github.com/nextm/nextm/internal/model"
)

type Repository interface {
	ListCollections(ctx context.Context, spaceID string) ([]*model.Collection, error)
	GetCollection(ctx context.Context, id string) (*model.Collection, error)
	CreateCollection(ctx context.Context, arg interface{}) (*model.Collection, error)
	UpdateCollection(ctx context.Context, arg interface{}) (*model.Collection, error)
	DeleteCollection(ctx context.Context, id string) error

	ListViews(ctx context.Context, collectionID string) ([]*model.CollectionView, error)
	CreateView(ctx context.Context, arg interface{}) (*model.CollectionView, error)
	UpdateView(ctx context.Context, arg interface{}) (*model.CollectionView, error)
	DeleteView(ctx context.Context, id string) error

	ListItems(ctx context.Context, collectionID string) ([]*model.CollectionItem, error)
	AddItem(ctx context.Context, arg interface{}) (*model.CollectionItem, error)
	RemoveItem(ctx context.Context, id string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ─── Collection ───────────────────────────────────────────

func (s *Service) List(ctx context.Context, spaceID string) ([]*model.Collection, error) {
	return s.repo.ListCollections(ctx, spaceID)
}

func (s *Service) Get(ctx context.Context, id string) (*model.Collection, error) {
	c, err := s.repo.GetCollection(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}
	return c, nil
}

func (s *Service) Create(ctx context.Context, spaceID string, req dto.CreateCollectionRequest) (*model.Collection, error) {
	now := model.NowMS()
	layout := req.Layout
	if layout == "" {
		layout = "table"
	}
	sourceType := req.SourceType
	if sourceType == "" {
		sourceType = "manual"
	}
	c, err := s.repo.CreateCollection(ctx, struct {
		ID           string
		SpaceID      string
		Name         string
		SourceType   string
		SourceConfig string
		Layout       string
		CreatedAt    int64
		UpdatedAt    int64
	}{
		ID: uuid.New().String(), SpaceID: spaceID,
		Name: req.Name, SourceType: sourceType,
		SourceConfig: req.SourceConfig, Layout: layout,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create collection: %w", err)
	}
	return c, nil
}

func (s *Service) Update(ctx context.Context, id string, req dto.UpdateCollectionRequest) (*model.Collection, error) {
	existing, err := s.repo.GetCollection(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	now := model.NowMS()
	c, err := s.repo.UpdateCollection(ctx, struct {
		ID           string
		SpaceID      string
		Name         string
		SourceType   string
		SourceConfig string
		Layout       string
		UpdatedAt    int64
	}{
		ID: id, SpaceID: existing.SpaceID,
		Name:         safeStr(req.Name, existing.Name),
		SourceType:   safeStr(req.SourceType, existing.SourceType),
		SourceConfig: safeStr(req.SourceConfig, existing.SourceConfig),
		Layout:       safeStr(req.Layout, existing.Layout),
		UpdatedAt:    now,
	})
	if err != nil {
		return nil, fmt.Errorf("update collection: %w", err)
	}
	return c, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteCollection(ctx, id)
}

// ─── View ─────────────────────────────────────────────────

func (s *Service) ListViews(ctx context.Context, collectionID string) ([]*model.CollectionView, error) {
	return s.repo.ListViews(ctx, collectionID)
}

func (s *Service) CreateView(ctx context.Context, collectionID string, req dto.CreateCollectionViewRequest) (*model.CollectionView, error) {
	now := model.NowMS()
	v, err := s.repo.CreateView(ctx, struct {
		ID            string
		CollectionID  string
		Name          string
		ViewType      string
		Filters       string
		Sorts         string
		VisibleFields string
		GroupBy       string
		CalendarField string
		KanbanField   string
		CreatedAt     int64
		UpdatedAt     int64
	}{
		ID: uuid.New().String(), CollectionID: collectionID,
		Name: req.Name, ViewType: req.ViewType,
		Filters: req.Filters, Sorts: req.Sorts,
		VisibleFields: req.VisibleFields, GroupBy: req.GroupBy,
		CalendarField: req.CalendarField, KanbanField: req.KanbanField,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create view: %w", err)
	}
	return v, nil
}

func (s *Service) UpdateView(ctx context.Context, id string, req dto.UpdateCollectionViewRequest) (*model.CollectionView, error) {
	v, err := s.repo.CreateView(ctx, struct {
		ID            string
		CollectionID  string
		Name          string
		ViewType      string
		Filters       string
		Sorts         string
		VisibleFields string
		GroupBy       string
		CalendarField string
		KanbanField   string
		CreatedAt     int64
		UpdatedAt     int64
	}{
		ID:            id,
		CollectionID:  "",
		Name:          safeStr(req.Name, ""),
		ViewType:      safeStr(req.ViewType, ""),
		Filters:       safeStr(req.Filters, ""),
		Sorts:         safeStr(req.Sorts, ""),
		VisibleFields: safeStr(req.VisibleFields, ""),
		GroupBy:       safeStr(req.GroupBy, ""),
		CalendarField: safeStr(req.CalendarField, ""),
		KanbanField:   safeStr(req.KanbanField, ""),
		CreatedAt:     0,
		UpdatedAt:     model.NowMS(),
	})
	if err != nil {
		return nil, fmt.Errorf("update view: %w", err)
	}
	return v, nil
}

func (s *Service) DeleteView(ctx context.Context, id string) error {
	return s.repo.DeleteView(ctx, id)
}

// ─── Item ─────────────────────────────────────────────────

func (s *Service) ListItems(ctx context.Context, collectionID string) ([]*model.CollectionItem, error) {
	return s.repo.ListItems(ctx, collectionID)
}

func (s *Service) AddItem(ctx context.Context, collectionID string, req dto.AddCollectionItemRequest) (*model.CollectionItem, error) {
	item, err := s.repo.AddItem(ctx, struct {
		ID           string
		CollectionID string
		ObjectID     string
		Position    float64
		Note        string
	}{
		ID: uuid.New().String(), CollectionID: collectionID,
		ObjectID: req.ObjectID, Position: req.Position,
		Note: req.Note,
	})
	if err != nil {
		return nil, fmt.Errorf("add item: %w", err)
	}
	return item, nil
}

func (s *Service) RemoveItem(ctx context.Context, id string) error {
	return s.repo.RemoveItem(ctx, id)
}

// ─── helpers ──────────────────────────────────────────────

func safeStr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

var _ = time.Now
