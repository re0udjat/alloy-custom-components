package contextprocessor

import (
	"context"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type contextMetricsProcessor struct {
	contextProcessor
	nextConsumer consumer.Metrics
}

func NewContextMetricsProcessor(
	logger *zap.Logger,
	nextConsumer consumer.Metrics,
	eventOptions trace.SpanStartEventOption,
	actions []ActionConfig) (*contextMetricsProcessor, error) {
	actionRunner := NewActionsRunner()
	for _, action := range actions {
		if err := actionRunner.AddAction(action); err != nil {
			return nil, err
		}
	}

	return &contextMetricsProcessor{
		contextProcessor: contextProcessor{
			logger:        logger,
			actionsRunner: actionRunner,
			eventOptions:  eventOptions,
		},
		nextConsumer: nextConsumer,
	}, nil
}

// implements https://pkg.go.dev/go.opentelemetry.io/collector/consumer#Metrics
func (ctxt *contextMetricsProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) (err error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("Start processing", ctxt.eventOptions)
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len() && err == nil; i++ {
		rm := rms.At(i)
		attrs := rm.Resource().Attributes()
		newCtx := ctxt.actionsRunner.Apply(ctx, attrs)
		newMd := pmetric.NewMetrics()
		rm.CopyTo(newMd.ResourceMetrics().AppendEmpty())
		err = ctxt.nextConsumer.ConsumeMetrics(newCtx, newMd)
	}
	span.AddEvent("End processing", ctxt.eventOptions)
	return err
}
