package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// StackTraceHandler is a handler that adds stack trace to error records
// and extracts request_id from context
type StackTraceHandler struct {
	slog.Handler
}

func (h *StackTraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
			r.AddAttrs(slog.String("request_id", reqID))
		}
	}

	if r.Level >= slog.LevelError {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		r.AddAttrs(slog.String("stack_trace", string(buf[:n])))
	}
	return h.Handler.Handle(ctx, r)
}

// InitStructuredLogger initialize structured logger
func InitStructuredLogger(level slog.Leveler) {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	if level.Level() == slog.LevelDebug {
		opts.AddSource = true
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
	handler := &StackTraceHandler{Handler: jsonHandler}

	slog.SetDefault(slog.New(handler))
}
