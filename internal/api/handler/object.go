package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nextm/nextm/internal/api/dto"
	"github.com/nextm/nextm/internal/api/middleware"
	"github.com/nextm/nextm/internal/model"
	"github.com/nextm/nextm/internal/pkg/httputil"
	"github.com/nextm/nextm/internal/service/object"
)

type ObjectHandler struct {
	svc *object.Service
}

func NewObjectHandler(svc *object.Service) *ObjectHandler {
	return &ObjectHandler{svc: svc}
}

func (h *ObjectHandler) List(w http.ResponseWriter, r *http.Request) {
	spaceID := middleware.SpaceIDFromContext(r.Context())
	if spaceID == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	typeID := r.URL.Query().Get("typeId")

	filter := model.ObjectFilter{
		SpaceID: spaceID,
		TypeID:  typeID,
		Limit:   limit,
		Offset:  offset,
	}

	objects, total, err := h.svc.List(r.Context(), spaceID, filter)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}

	httputil.WriteJSONMeta(w, http.StatusOK, objects, &httputil.Meta{
		Total:   int(total),
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: filter.Offset+filter.Limit < int(total),
	})
}

func (h *ObjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	spaceID := middleware.SpaceIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if spaceID == "" || id == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	obj, err := h.svc.Get(r.Context(), id, spaceID)
	if err != nil {
		httputil.WriteError(w, httputil.ErrNotFound)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, obj)
}

func (h *ObjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	spaceID := middleware.SpaceIDFromContext(r.Context())
	if spaceID == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	var req dto.CreateObjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	if req.Title == "" || req.TypeID == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	obj, err := h.svc.Create(r.Context(), spaceID, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, obj)
}

func (h *ObjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	spaceID := middleware.SpaceIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if spaceID == "" || id == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	var req dto.UpdateObjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	obj, err := h.svc.Update(r.Context(), id, spaceID, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrNotFound)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, obj)
}

func (h *ObjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	spaceID := middleware.SpaceIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if spaceID == "" || id == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	if err := h.svc.Delete(r.Context(), id, spaceID); err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ObjectHandler) Archive(w http.ResponseWriter, r *http.Request) {
	spaceID := middleware.SpaceIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if spaceID == "" || id == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	if err := h.svc.Archive(r.Context(), id, spaceID); err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

func (h *ObjectHandler) Search(w http.ResponseWriter, r *http.Request) {
	spaceID := middleware.SpaceIDFromContext(r.Context())
	if spaceID == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	objects, err := h.svc.Search(r.Context(), spaceID, query, limit, offset)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, objects)
}

// ─── 块 Handler ────────────────────────────────────────

func (h *ObjectHandler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	objectID := chi.URLParam(r, "id")

	blocks, err := h.svc.ListBlocks(r.Context(), objectID)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, blocks)
}

func (h *ObjectHandler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	objectID := chi.URLParam(r, "id")

	var req dto.CreateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	if req.Type == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	block, err := h.svc.CreateBlock(r.Context(), objectID, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, block)
}

func (h *ObjectHandler) UpdateBlock(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	block, err := h.svc.UpdateBlock(r.Context(), id, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrNotFound)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, block)
}

func (h *ObjectHandler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	objectID := chi.URLParam(r, "objectId")

	if err := h.svc.DeleteBlock(r.Context(), id, objectID); err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes 注册对象路由
func (h *ObjectHandler) RegisterRoutes(r chiRouter) {
	// 搜索
	r.Get("/objects/search", h.Search)

	// 对象 CRUD
	r.Get("/objects", h.List)
	r.Post("/objects", h.Create)
	r.Get("/objects/{id}", h.Get)
	r.Put("/objects/{id}", h.Update)
	r.Delete("/objects/{id}", h.Delete)
	r.Patch("/objects/{id}/archive", h.Archive)

	// 块管理
	r.Get("/objects/{id}/blocks", h.ListBlocks)
	r.Post("/objects/{id}/blocks", h.CreateBlock)
	r.Put("/blocks/{id}", h.UpdateBlock)
	r.Delete("/blocks/{id}", h.DeleteBlock)
}
