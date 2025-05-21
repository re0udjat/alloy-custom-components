package contextprocessor

import (
	"context"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type contextTracesProcessor struct {
	contextProcessor
	nextConsumer consumer.Traces
}

func NewContextTracesProcessor(
	logger *zap.Logger,
	nextConsumer consumer.Traces,
	eventOptions trace.SpanStartEventOption,
	actions []ActionConfig) (*contextTracesProcessor, error) {
	actionRunner := NewActionsRunner()

	for _, action := range actions {
		if err := actionRunner.AddAction(action); err != nil {
			return nil, err
		}
	}

	return &contextTracesProcessor{
		contextProcessor: contextProcessor{
			logger:        logger,
			actionsRunner: actionRunner,
			eventOptions:  eventOptions,
		},
		nextConsumer: nextConsumer,
	}, nil
}

// Implements https://pkg.go.dev/go.opentelemetry.io/collector/consumer#Traces
func (ctxt *contextTracesProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) (err error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("Start processing", ctxt.eventOptions)

	rss := td.ResourceSpans()
	for i := 0; i < rss.Len() && err == nil; i++ {
		rt := rss.At(i)
		attrs := rt.Resource().Attributes()
		newCtx := ctxt.actionsRunner.Apply(ctx, attrs)
		newTd := ptrace.NewTraces()
		rt.CopyTo(newTd.ResourceSpans().AppendEmpty())
		err = ctxt.nextConsumer.ConsumeTraces(newCtx, newTd)
	}

	span.AddEvent("End processing", ctxt.eventOptions)
	return err
}
