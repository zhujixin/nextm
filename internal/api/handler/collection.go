package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nextm/nextm/internal/api/dto"
	"github.com/nextm/nextm/internal/api/middleware"
	"github.com/nextm/nextm/internal/pkg/httputil"
	"github.com/nextm/nextm/internal/service/collection"
)

type CollectionHandler struct {
	svc *collection.Service
}

func NewCollectionHandler(svc *collection.Service) *CollectionHandler {
	return &CollectionHandler{svc: svc}
}

// ─── Collection CRUD ──────────────────────────────────────

func (h *CollectionHandler) List(w http.ResponseWriter, r *http.Request) {
	spaceID := middleware.SpaceIDFromContext(r.Context())
	if spaceID == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}
	cols, err := h.svc.List(r.Context(), spaceID)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, cols)
}

func (h *CollectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	col, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, httputil.ErrNotFound)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, col)
}

func (h *CollectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	spaceID := middleware.SpaceIDFromContext(r.Context())
	if spaceID == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	var req dto.CreateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}
	if req.Name == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	col, err := h.svc.Create(r.Context(), spaceID, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, col)
}

func (h *CollectionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	col, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrNotFound)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, col)
}

func (h *CollectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Views ────────────────────────────────────────────────

func (h *CollectionHandler) ListViews(w http.ResponseWriter, r *http.Request) {
	collectionID := chi.URLParam(r, "id")
	views, err := h.svc.ListViews(r.Context(), collectionID)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, views)
}

func (h *CollectionHandler) CreateView(w http.ResponseWriter, r *http.Request) {
	collectionID := chi.URLParam(r, "id")

	var req dto.CreateCollectionViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}
	if req.Name == "" || req.ViewType == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	view, err := h.svc.CreateView(r.Context(), collectionID, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, view)
}

func (h *CollectionHandler) UpdateView(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateCollectionViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	view, err := h.svc.UpdateView(r.Context(), id, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrNotFound)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, view)
}

func (h *CollectionHandler) DeleteView(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteView(r.Context(), id); err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Items ────────────────────────────────────────────────

func (h *CollectionHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	collectionID := chi.URLParam(r, "id")
	items, err := h.svc.ListItems(r.Context(), collectionID)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *CollectionHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	collectionID := chi.URLParam(r, "id")

	var req dto.AddCollectionItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}
	if req.ObjectID == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	item, err := h.svc.AddItem(r.Context(), collectionID, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *CollectionHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "itemId")
	if err := h.svc.RemoveItem(r.Context(), id); err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
