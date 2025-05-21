package contextprocessor

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

// The current context and the attributes
type evContext struct {
	ctx           context.Context
	cliInfo       client.Info
	resourceAttrs pcommon.Map
	newMetadata   map[string][]string
}

// Actions
type Action interface {
	execute(*evContext)
}

type ActionsRunner struct {
	actions []Action
}

// Implement actions
func generateAction(action ActionConfig) (Action, error) {
	valueDefault := ""
	if action.ValueDefault != nil {
		valueDefault = *action.ValueDefault
	}

	fromAttribute := ""
	if action.FromAttribute != nil {
		fromAttribute = *action.FromAttribute
	}

	switch action.Action {
	case INSERT:
		return &actionInsert{
			key:      *action.Key,
			value:    valueDefault,
			fromAttr: fromAttribute,
		}, nil
	case UPSERT:
		return &actionUpsert{
			key:      *action.Key,
			value:    valueDefault,
			fromAttr: fromAttribute,
		}, nil
	case UPDATE:
		return &actionUpdate{
			key:      *action.Key,
			value:    valueDefault,
			fromAttr: fromAttribute,
		}, nil
	case DELETE:
		return &actionDelete{
			key:      *action.Key,
			value:    valueDefault,
			fromAttr: fromAttribute,
		}, nil
	default:
		return nil, fmt.Errorf("Unknown action type")
	}
}

type actionInsert struct {
	key      string
	value    string
	fromAttr string
}

func (a *actionInsert) execute(evContext *evContext) {
	value := []string{a.value}
	if len(a.fromAttr) > 0 {
		value[0], _ = evContext.getAttrKey(a.fromAttr, a.value)
	}
	if currentValue, exists := evContext.getContextKey(a.key); !exists {
		evContext.setContextKey(a.key, value)
	} else {
		evContext.setContextKey(a.key, currentValue)
	}
}

type actionUpsert struct {
	key      string
	value    string
	fromAttr string
}

func (a *actionUpsert) execute(evContext *evContext) {
	value := []string{a.value}
	if len(a.fromAttr) > 0 {
		value[0], _ = evContext.getAttrKey(a.fromAttr, a.value)
	}
	evContext.setContextKey(a.key, value)
}

type actionUpdate struct {
	key      string
	value    string
	fromAttr string
}

func (a *actionUpdate) execute(evContext *evContext) {
	value := []string{a.value}
	if len(a.fromAttr) > 0 {
		value[0], _ = evContext.getAttrKey(a.fromAttr, a.value)
	}

	if v, exists := evContext.getContextKey(a.key); exists {
		// Add the new value to the current list of strings, or
		evContext.setContextKey(a.key, append(v, value[0]))
		// Overwriting the current value
		// evContext.setContextKey(a.key, value)
	}
}

type actionDelete struct {
	key      string
	value    string
	fromAttr string
}

func (a *actionDelete) execute(evContext *evContext) {
	evContext.delContextKey(a.key)
}

// Implement event context
func createEvenContext(ctx context.Context, attrs pcommon.Map) *evContext {
	return &evContext{
		ctx:           ctx,
		cliInfo:       client.FromContext(ctx),
		resourceAttrs: attrs,
		newMetadata:   make(map[string][]string),
	}
}

func (evContext *evContext) getContext() context.Context {
	return client.NewContext(evContext.ctx, client.Info{
		Metadata: client.NewMetadata(evContext.newMetadata),
	})
}

func (evContext *evContext) getContextKey(key string) ([]string, bool) {
	if v, exists := evContext.newMetadata[key]; exists {
		return v, exists
	} else {
		value := evContext.cliInfo.Metadata.Get(key)
		return value, (len(value) != 0)
	}
}

func (evContext *evContext) setContextKey(key string, value []string) {
	evContext.newMetadata[key] = value
}

func (evContext *evContext) delContextKey(key string) {
	// Delete a key is only deleted from newMetadata -> it's available again from actual metadata
	delete(evContext.newMetadata, key)
}

func (evContext *evContext) getAttrKey(key, def string) (string, bool) {
	value := def
	v, exists := evContext.resourceAttrs.Get(key)
	if exists {
		switch v.Type() {
		case pcommon.ValueTypeStr:
			value = v.Str()
		default:
			value = v.AsString()
		}
	}
	return value, exists
}

// Implement actions runner
func NewActionsRunner() *ActionsRunner {
	return &ActionsRunner{
		actions: make([]Action, 0),
	}
}

func (ar *ActionsRunner) AddAction(actionCfg ActionConfig) error {
	action, err := generateAction(actionCfg)
	if err == nil {
		ar.actions = append(ar.actions, action)
	}
	return err
}

// The executeCommands method executes all the commands one by one
func (ar *ActionsRunner) Apply(ctx context.Context, attrs pcommon.Map) context.Context {
	evContext := createEvenContext(ctx, attrs)
	for _, a := range ar.actions {
		a.execute(evContext)
	}
	return evContext.getContext()
}
