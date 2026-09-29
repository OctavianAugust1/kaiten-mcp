package app

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

const redactedValue = "[REDACTED]"

// NewLogger creates the process logger on the supplied stderr writer. The
// caller keeps stdout exclusively for the MCP protocol stream.
func NewLogger(level slog.Level, stderr io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: level})
	return slog.New(redactingHandler{next: handler})
}

type redactingHandler struct {
	next slog.Handler
}

func (h redactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h redactingHandler) Handle(ctx context.Context, record slog.Record) error {
	clean := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		clean.AddAttrs(redactAttr(attr))
		return true
	})
	return h.next.Handle(ctx, clean)
}

func (h redactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clean := make([]slog.Attr, len(attrs))
	for index, attr := range attrs {
		clean[index] = redactAttr(attr)
	}
	return redactingHandler{next: h.next.WithAttrs(clean)}
}

func (h redactingHandler) WithGroup(name string) slog.Handler {
	return redactingHandler{next: h.next.WithGroup(name)}
}

func redactAttr(attr slog.Attr) slog.Attr {
	attr.Value = attr.Value.Resolve()
	if sensitiveKey(attr.Key) {
		return slog.String(attr.Key, redactedValue)
	}
	if attr.Value.Kind() != slog.KindGroup {
		return attr
	}
	group := attr.Value.Group()
	clean := make([]slog.Attr, len(group))
	for index, child := range group {
		clean[index] = redactAttr(child)
	}
	return slog.Group(attr.Key, attrsToAny(clean)...)
}

func sensitiveKey(key string) bool {
	normalized := strings.NewReplacer("-", "_", " ", "_").Replace(strings.ToLower(key))
	for _, fragment := range []string{"token", "authorization", "secret", "password", "api_key"} {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}

func attrsToAny(attrs []slog.Attr) []any {
	values := make([]any, len(attrs))
	for index, attr := range attrs {
		values[index] = attr
	}
	return values
}
