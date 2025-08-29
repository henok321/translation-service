package logging

import (
	"context"
	"log/slog"

	"github.com/henok321/translation-service/api/middleware"
)

type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if requestContext, ok := ctx.Value(middleware.RequestLoggingContext).(middleware.Request); ok {
		requestGroup := slog.Group("request", "id", requestContext.ID, "method", requestContext.Method, "path", requestContext.Path)
		r.AddAttrs(requestGroup)
	}

	return h.Handler.Handle(ctx, r)
}
