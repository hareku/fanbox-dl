package ctxval

import (
	"context"
	"log/slog"
)

type slogAttrsKey struct{}

func AddSlogAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	s, ok := GetSlogAttrs(ctx)
	if !ok {
		s = []slog.Attr{}
	}

	s = append(s, attrs...)
	return context.WithValue(ctx, slogAttrsKey{}, s)
}

func GetSlogAttrs(ctx context.Context) ([]slog.Attr, bool) {
	v, ok := ctx.Value(slogAttrsKey{}).([]slog.Attr)
	return v, ok
}
