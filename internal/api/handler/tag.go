package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nextm/nextm/internal/api/dto"
	"github.com/nextm/nextm/internal/api/middleware"
	"github.com/nextm/nextm/internal/pkg/httputil"
	"github.com/nextm/nextm/internal/service/tag"
)

type TagHandler struct {
	svc *tag.Service
}

func NewTagHandler(svc *tag.Service) *TagHandler {
	return &TagHandler{svc: svc}
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	spaceID := middleware.SpaceIDFromContext(r.Context())
	if spaceID == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}
	tags, err := h.svc.List(r.Context(), spaceID)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, tags)
}

func (h *TagHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, httputil.ErrNotFound)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, tag)
}

func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	spaceID := middleware.SpaceIDFromContext(r.Context())
	if spaceID == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	var req dto.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}
	if req.Name == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	tag, err := h.svc.Create(r.Context(), spaceID, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, tag)
}

func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	tag, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrNotFound)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, tag)
}

func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TagHandler) GetObjectTags(w http.ResponseWriter, r *http.Request) {
	objectID := chi.URLParam(r, "id")
	tags, err := h.svc.GetObjectTags(r.Context(), objectID)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, tags)
}

func (h *TagHandler) AssignTags(w http.ResponseWriter, r *http.Request) {
	objectID := chi.URLParam(r, "id")

	var req dto.AssignTagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	if err := h.svc.AssignTags(r.Context(), objectID, req); err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "assigned"})
}

func (h *TagHandler) UnassignTag(w http.ResponseWriter, r *http.Request) {
	objectID := chi.URLParam(r, "id")
	tagID := chi.URLParam(r, "tagId")

	if err := h.svc.UnassignTag(r.Context(), objectID, tagID); err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
