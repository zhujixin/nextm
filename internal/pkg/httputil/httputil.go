package httputil

import (
	"encoding/json"
	"net/http"
)

// APIError 统一 API 错误响应
type APIError struct {
	Code     int         `json:"code"`
	Message  string      `json:"message"`
	Details  interface{} `json:"details,omitempty"`
	Severity string      `json:"-"` // debug | info | warn | error
	Err      error       `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// Predefined API errors
var (
	ErrUnauthorized = &APIError{Code: 2000, Message: "未认证", Severity: "info"}
	ErrTokenExpired = &APIError{Code: 2001, Message: "Token 已过期", Severity: "info"}
	ErrInvalidCred  = &APIError{Code: 2002, Message: "凭据无效", Severity: "info"}
	ErrForbidden    = &APIError{Code: 3000, Message: "权限不足", Severity: "warn"}
	ErrNotFound     = &APIError{Code: 4000, Message: "资源未找到", Severity: "info"}
	ErrConflict     = &APIError{Code: 4001, Message: "资源冲突", Severity: "warn"}
	ErrValidation   = &APIError{Code: 5000, Message: "验证错误", Severity: "info"}
	ErrInvalidInput = &APIError{Code: 5001, Message: "无效输入", Severity: "info"}
	ErrTooLarge     = &APIError{Code: 5002, Message: "请求体过大", Severity: "info"}
	ErrRateLimited  = &APIError{Code: 1003, Message: "请求限流", Severity: "warn"}
	ErrInternal     = &APIError{Code: 1001, Message: "内部错误", Severity: "error"}
	ErrUnavailable  = &APIError{Code: 1002, Message: "服务不可用", Severity: "error"}
)

// Response 成功响应
type Response struct {
	Data interface{} `json:"data"`
	Meta *Meta       `json:"meta,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// Meta 元数据
type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Total     int    `json:"total,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
	HasMore   bool   `json:"has_more,omitempty"`
}

type PaginationParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// WriteJSON 写 JSON 响应
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	resp := Response{Data: data}
	json.NewEncoder(w).Encode(resp)
}

// WriteJSONMeta 写带 Meta 的 JSON 响应
func WriteJSONMeta(w http.ResponseWriter, status int, data interface{}, meta *Meta) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	resp := Response{Data: data, Meta: meta}
	json.NewEncoder(w).Encode(resp)
}

// WriteError 写错误响应
func WriteError(w http.ResponseWriter, apiErr *APIError) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	status := http.StatusInternalServerError
	switch {
	case apiErr.Code >= 5000:
		status = http.StatusBadRequest
	case apiErr.Code >= 4000:
		status = http.StatusNotFound
	case apiErr.Code >= 3000:
		status = http.StatusForbidden
	case apiErr.Code >= 2000:
		status = http.StatusUnauthorized
	case apiErr.Code >= 1003:
		status = http.StatusTooManyRequests
	case apiErr.Code >= 1000:
		status = http.StatusInternalServerError
	}

	if apiErr.Code == 4001 {
		status = http.StatusConflict
	}
	if apiErr.Code == 5002 {
		status = http.StatusRequestEntityTooLarge
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: *apiErr})
}

// WriteInternalError 写内部错误，避免暴露详细信息
func WriteInternalError(w http.ResponseWriter) {
	WriteError(w, ErrInternal)
}
