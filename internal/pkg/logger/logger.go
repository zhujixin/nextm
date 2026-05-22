package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type ctxKey string

const traceIDKey ctxKey = "trace_id"

func New(level string, format string) *slog.Logger {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	var h slog.Handler
	opts := &slog.HandlerOptions{
		Level: l,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					return slog.String("source", source.File+":"+itoa(int(source.Line)))
				}
			}
			return a
		},
	}

	switch format {
	case "text":
		h = slog.NewTextHandler(os.Stdout, opts)
	default:
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(h)
}

func WithTraceID(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return logger.With("trace_id", traceID)
	}
	return logger
}

func WithSource(logger *slog.Logger) *slog.Logger {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	fn := runtime.FuncForPC(pcs[0])
	return logger.With(slog.Any("source", &slog.Source{
		Function: fn.Name(),
		File:     "",
		Line:     0,
	}))
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

type TraceIDKey struct{}

func NewContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey{}, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(TraceIDKey{}).(string); ok {
		return id
	}
	return ""
}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey("logger"), logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ctxKey("logger")).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// WithDuration returns a slog attribute with human-readable duration.
func WithDuration(d time.Duration) slog.Attr {
	return slog.String("duration", d.Round(time.Millisecond).String())
}
