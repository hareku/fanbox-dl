package applog

import (
	"context"
	"log/slog"

	"github.com/hareku/fanbox-dl/internal/ctxval"
)

type ContextValueLogHandler struct {
	slog.Handler
}

func NewContextValueLogHandler(h slog.Handler) *ContextValueLogHandler {
	return &ContextValueLogHandler{Handler: h}
}

func (h *ContextValueLogHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctxval.GetSlogAttrs(ctx); ok {
		r.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, r)
}
