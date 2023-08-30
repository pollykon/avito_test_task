package logger

import (
	"context"
	"log/slog"
)

// Slogger which does nothing (for tests)

type NoopHandler struct{}

func NewNoopHandler() slog.Handler {
	return NoopHandler{}
}

func (h NoopHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (h NoopHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (h NoopHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h NoopHandler) WithGroup(name string) slog.Handler {
	return h
}
