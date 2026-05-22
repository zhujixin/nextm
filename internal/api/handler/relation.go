package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nextm/nextm/internal/api/dto"
	"github.com/nextm/nextm/internal/pkg/httputil"
	"github.com/nextm/nextm/internal/service/relation"
)

type RelationHandler struct {
	svc *relation.Service
}

func NewRelationHandler(svc *relation.Service) *RelationHandler {
	return &RelationHandler{svc: svc}
}

func (h *RelationHandler) ListByObject(w http.ResponseWriter, r *http.Request) {
	objectID := chi.URLParam(r, "id")
	rels, err := h.svc.ListByObject(r.Context(), objectID)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, rels)
}

func (h *RelationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRelationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}
	if req.SourceID == "" || req.TargetID == "" || req.Type == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	rel, err := h.svc.Create(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, rel)
}

func (h *RelationHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateRelationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	rel, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		httputil.WriteError(w, httputil.ErrNotFound)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, rel)
}

func (h *RelationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
