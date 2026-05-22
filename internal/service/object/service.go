package object

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nextm/nextm/internal/api/dto"
	"github.com/nextm/nextm/internal/model"
)

type ObjectRepository interface {
	ListObjects(ctx context.Context, spaceID string, limit, offset int) ([]*model.KnowledgeObject, error)
	ListObjectsByType(ctx context.Context, spaceID, typeID string, limit, offset int) ([]*model.KnowledgeObject, error)
	GetObject(ctx context.Context, id, spaceID string) (*model.KnowledgeObject, error)
	CreateObject(ctx context.Context, arg interface{}) (*model.KnowledgeObject, error)
	UpdateObject(ctx context.Context, arg interface{}) (*model.KnowledgeObject, error)
	SoftDeleteObject(ctx context.Context, id, spaceID string, updatedAt int64) error
	ArchiveObject(ctx context.Context, id, spaceID string, updatedAt int64) error
	SearchObjects(ctx context.Context, spaceID, query string, limit, offset int) ([]*model.KnowledgeObject, error)
	CountObjectsBySpace(ctx context.Context, spaceID string) (int64, error)
	CountObjectsByType(ctx context.Context, spaceID, typeID string) (int64, error)
}

type BlockRepository interface {
	ListBlocksByObject(ctx context.Context, objectID string) ([]*model.Block, error)
	GetBlock(ctx context.Context, id string) (*model.Block, error)
	CreateBlock(ctx context.Context, arg interface{}) (*model.Block, error)
	UpdateBlock(ctx context.Context, arg interface{}) (*model.Block, error)
	DeleteBlock(ctx context.Context, id, objectID string) error
	CountBlocksByObject(ctx context.Context, objectID string) (int64, error)
}

type Service struct {
	objectRepo ObjectRepository
	blockRepo  BlockRepository
}

func NewService(objectRepo ObjectRepository, blockRepo BlockRepository) *Service {
	return &Service{
		objectRepo: objectRepo,
		blockRepo:  blockRepo,
	}
}

func (s *Service) List(ctx context.Context, spaceID string, filter model.ObjectFilter) ([]*model.KnowledgeObject, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	var objects []*model.KnowledgeObject
	var err error

	if filter.TypeID != "" {
		objects, err = s.objectRepo.ListObjectsByType(ctx, spaceID, filter.TypeID, filter.Limit, filter.Offset)
	} else {
		objects, err = s.objectRepo.ListObjects(ctx, spaceID, filter.Limit, filter.Offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("list objects: %w", err)
	}

	total, err := s.objectRepo.CountObjectsBySpace(ctx, spaceID)
	if err != nil {
		return nil, 0, fmt.Errorf("count objects: %w", err)
	}

	return objects, total, nil
}

func (s *Service) Get(ctx context.Context, id, spaceID string) (*model.KnowledgeObject, error) {
	obj, err := s.objectRepo.GetObject(ctx, id, spaceID)
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	return obj, nil
}

func (s *Service) Create(ctx context.Context, spaceID string, req dto.CreateObjectRequest) (*model.KnowledgeObject, error) {
	now := model.NowMS()
	id := uuid.New().String()

	source := req.Source
	if source == "" {
		source = "manual"
	}

	obj, err := s.objectRepo.CreateObject(ctx, struct {
		ID         string
		SpaceID    string
		TypeID     string
		Title      string
		Properties string
		Tags       string
		CoverImage *string
		Source     string
		SourceMeta string
		WordCount  int
		CreatedAt  int64
		UpdatedAt  int64
	}{
		ID: id, SpaceID: spaceID, TypeID: req.TypeID,
		Title: req.Title, Tags: req.Tags,
		CoverImage: strPtr(req.CoverImage),
		Source: source, SourceMeta: "{}",
		WordCount: len(req.Title), CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create object: %w", err)
	}

	return obj, nil
}

func (s *Service) Update(ctx context.Context, id, spaceID string, req dto.UpdateObjectRequest) (*model.KnowledgeObject, error) {
	// 先检查是否存在
	_, err := s.objectRepo.GetObject(ctx, id, spaceID)
	if err != nil {
		return nil, fmt.Errorf("object not found: %w", err)
	}

	now := model.NowMS()
	obj, err := s.objectRepo.UpdateObject(ctx, struct {
		ID        string
		SpaceID   string
		Title     string
		Tags      string
		CoverImage *string
		Properties string
		WordCount int
		UpdatedAt int64
	}{
		ID: id, SpaceID: spaceID,
		Title:     safeStr(req.Title),
		Tags:      safeStr(req.Tags),
		CoverImage: req.CoverImage,
		Properties: safeStr(req.Properties),
		WordCount: len(safeStr(req.Title)),
		UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("update object: %w", err)
	}

	return obj, nil
}

func (s *Service) Delete(ctx context.Context, id, spaceID string) error {
	now := model.NowMS()
	if err := s.objectRepo.SoftDeleteObject(ctx, id, spaceID, now); err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}

func (s *Service) Archive(ctx context.Context, id, spaceID string) error {
	now := model.NowMS()
	if err := s.objectRepo.ArchiveObject(ctx, id, spaceID, now); err != nil {
		return fmt.Errorf("archive object: %w", err)
	}
	return nil
}

func (s *Service) Search(ctx context.Context, spaceID, query string, limit, offset int) ([]*model.KnowledgeObject, error) {
	if limit <= 0 {
		limit = 20
	}
	objects, err := s.objectRepo.SearchObjects(ctx, spaceID, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search objects: %w", err)
	}
	return objects, nil
}

// ─── Block 方法 ────────────────────────────────────────

func (s *Service) ListBlocks(ctx context.Context, objectID string) ([]*model.Block, error) {
	return s.blockRepo.ListBlocksByObject(ctx, objectID)
}

func (s *Service) CreateBlock(ctx context.Context, objectID string, req dto.CreateBlockRequest) (*model.Block, error) {
	now := model.NowMS()
	id := uuid.New().String()

	block, err := s.blockRepo.CreateBlock(ctx, struct {
		ID         string
		ObjectID   string
		ParentID   *string
		Type       string
		Content    string
		Properties string
		Position   float64
		Depth      int
		Color      string
		CreatedAt  int64
		UpdatedAt  int64
	}{
		ID: id, ObjectID: objectID,
		ParentID: req.ParentID, Type: req.Type,
		Content: req.Content, Properties: req.Properties,
		Position: req.Position, Depth: req.Depth,
		Color: req.Color, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create block: %w", err)
	}

	return block, nil
}

func (s *Service) UpdateBlock(ctx context.Context, id string, req dto.UpdateBlockRequest) (*model.Block, error) {
	block, err := s.blockRepo.GetBlock(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("block not found: %w", err)
	}

	now := model.NowMS()
	updated, err := s.blockRepo.UpdateBlock(ctx, struct {
		ID         string
		ObjectID   string
		Content    string
		Properties string
		Type       string
		Position   float64
		Depth      int
		Color      string
		UpdatedAt  int64
	}{
		ID: id, ObjectID: block.ObjectID,
		Content:    safeStrVal(req.Content, block.Content),
		Properties: block.Properties,
		Type:       safeStrVal(req.Type, block.Type),
		Position:   safeFloat(req.Position, block.Position),
		Depth:      safeInt(req.Depth, block.Depth),
		Color:      safeStrVal(req.Color, block.Color),
		UpdatedAt:  now,
	})
	if err != nil {
		return nil, fmt.Errorf("update block: %w", err)
	}

	return updated, nil
}

func (s *Service) DeleteBlock(ctx context.Context, id, objectID string) error {
	return s.blockRepo.DeleteBlock(ctx, id, objectID)
}

// ─── 辅助函数 ──────────────────────────────────────────

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeStrVal(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

func safeFloat(f *float64, fallback float64) float64 {
	if f == nil {
		return fallback
	}
	return *f
}

func safeInt(i *int, fallback int) int {
	if i == nil {
		return fallback
	}
	return *i
}

// suppress unused import
var _ = time.Now
