package handler

import (
	"encoding/json"
	"net/http"

	"github.com/nextm/nextm/internal/api/dto"
	"github.com/nextm/nextm/internal/api/middleware"
	"github.com/nextm/nextm/internal/pkg/httputil"
	"github.com/nextm/nextm/internal/service/auth"
)

type AuthHandler struct {
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		httputil.WriteError(w, httputil.ErrValidation)
		return
	}

	if len(req.Password) < 8 {
		httputil.WriteError(w, &httputil.APIError{
			Code:    5000,
			Message: "密码长度至少 8 位",
		})
		return
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, &httputil.APIError{
			Code:    409,
			Message: err.Error(),
		})
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, &httputil.APIError{
			Code:    2002,
			Message: err.Error(),
		})
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, httputil.ErrInvalidInput)
		return
	}

	resp, err := h.svc.RefreshToken(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, &httputil.APIError{
			Code:    2001,
			Message: "Token 已过期或无效",
		})
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

// GetAccounts 获取已登录账号列表
func (h *AuthHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	// TODO: 实现多账号列表
	httputil.WriteJSON(w, http.StatusOK, []interface{}{})
}

// SwitchAccount 切换账号
func (h *AuthHandler) SwitchAccount(w http.ResponseWriter, r *http.Request) {
	httputil.WriteError(w, &httputil.APIError{
		Code:    1001,
		Message: "暂未实现",
	})
}

// Logout 登出
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	accountID := middleware.AccountIDFromContext(r.Context())
	if accountID == "" {
		httputil.WriteError(w, httputil.ErrUnauthorized)
		return
	}

	// TODO: 调用 service 登出逻辑
	if err := h.svc.RevokeAllTokens(r.Context(), accountID); err != nil {
		httputil.WriteError(w, httputil.ErrInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

// RegisterRoutes 注册认证路由
func (h *AuthHandler) RegisterRoutes(r chiRouter) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.Refresh)
	r.Get("/auth/accounts", h.GetAccounts)
	r.Post("/auth/switch", h.SwitchAccount)
	r.Delete("/auth/accounts/{id}", h.Logout)
}

type chiRouter interface {
	Post(string, http.HandlerFunc)
	Get(string, http.HandlerFunc)
	Put(string, http.HandlerFunc)
	Patch(string, http.HandlerFunc)
	Delete(string, http.HandlerFunc)
}
