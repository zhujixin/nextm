package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey string

const (
	traceIDKey ctxKey = "trace_id"
	spaceIDKey ctxKey = "space_id"
	accountIDKey ctxKey = "account_id"
)

func TraceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

func SpaceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(spaceIDKey).(string); ok {
		return id
	}
	return ""
}

func AccountIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(accountIDKey).(string); ok {
		return id
	}
	return ""
}

func contextWithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

// RequestID 注入请求 ID 到 context 和 response header
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(contextWithTraceID(r.Context(), id)))
	})
}

// CORS 处理跨域
func CORS(allowedOrigins, allowedHeaders []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if len(allowedOrigins) > 0 && origin != "" {
				for _, o := range allowedOrigins {
					if o == "*" || o == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", joinHeaders(allowedHeaders))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func joinHeaders(headers []string) string {
	if len(headers) == 0 {
		return "Authorization, Content-Type, X-Space-ID, X-Request-ID"
	}
	result := ""
	for i, h := range headers {
		if i > 0 {
			result += ", "
		}
		result += h
	}
	return result
}
