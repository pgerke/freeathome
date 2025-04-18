package freeathome

import (
	"context"
	"log/slog"
)

type ChannelHandler struct {
	next    slog.Handler
	records chan slog.Record
}

func (h *ChannelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *ChannelHandler) Handle(ctx context.Context, r slog.Record) error {
	h.records <- r.Clone()
	return h.next.Handle(ctx, r)
}

func (h *ChannelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ChannelHandler{
		next:    h.next.WithAttrs(attrs),
		records: h.records,
	}
}

func (h *ChannelHandler) WithGroup(name string) slog.Handler {
	return &ChannelHandler{
		next:    h.next.WithGroup(name),
		records: h.records,
	}
}
