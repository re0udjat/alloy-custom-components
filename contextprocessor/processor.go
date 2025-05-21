package contextprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type contextProcessor struct {
	logger        *zap.Logger
	actionsRunner *ActionsRunner
	cancel        context.CancelFunc
	eventOptions  trace.SpanStartEventOption
}

// Implements https://pkg.go.dev/go.opentelemetry.io/collector/component#Component Start
func (ctxp *contextProcessor) Start(ctx context.Context, host component.Host) error {
	ctx = context.Background()
	ctx, ctxp.cancel = context.WithCancel(ctx)
	for extension, _ := range host.GetExtensions() {
		ctxp.logger.Info("Extension", zap.String("id", extension.String()))
	}
	return nil
}

// Implements https://pkg.go.dev/go.opentelemetry.io/collector/component#Component Shutdown
func (ctxp *contextProcessor) Shutdown(ctx context.Context) error {
	ctxp.cancel()
	return nil
}
