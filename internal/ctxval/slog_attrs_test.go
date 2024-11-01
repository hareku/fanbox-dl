package ctxval

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddSlogAttrs(t *testing.T) {
	ctx := context.Background()

	_, ok := GetSlogAttrs(ctx)
	require.False(t, ok)

	ctx = AddSlogAttrs(ctx, slog.String("k1", "v1"))

	v, ok := GetSlogAttrs(ctx)
	require.True(t, ok)
	require.Equal(t, []slog.Attr{slog.String("k1", "v1")}, v)

	ctx2 := AddSlogAttrs(ctx, slog.String("k2", "v2"))
	v, ok = GetSlogAttrs(ctx2)
	require.True(t, ok)
	require.Equal(t, []slog.Attr{slog.String("k1", "v1"), slog.String("k2", "v2")}, v)

	v, ok = GetSlogAttrs(ctx)
	require.True(t, ok)
	require.Equal(t, []slog.Attr{slog.String("k1", "v1")}, v, "original context should not be modified")
}
