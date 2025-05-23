package contextprocessor

import "fmt"

var (
	errMissingActionConfig       = fmt.Errorf("Missing actions configuration")
	errMissingActionConfigKey    = fmt.Errorf("Missing action key")
	errMissingActionConfigSource = fmt.Errorf("Missing action source, must be 'from_attribute' or 'value'")
	errMissingActionDeleteParams = fmt.Errorf("Action delete does not support 'from_attribute' and/or 'value'")
)

// ActionValue is the enum to capture the four types of actions to perform on the context
type ActionType string

const (
	INSERT ActionType = "insert"
	UPDATE ActionType = "update"
	UPSERT ActionType = "upsert"
	DELETE ActionType = "delete"
)

type ActionConfig struct {
	Key           *string    `mapstructure:"key"`
	Action        ActionType `mapstructure:"action"`
	ValueDefault  *string    `mapstructure:"value"`
	FromAttribute *string    `mapstructure:"from_attribute"`
}

// Represents the receiver config settings within the collector's config.yaml
type Config struct {
	ActionsConfig []ActionConfig `mapstructure:"actions"`
}

func (cfg *Config) Validate() error {
	if cfg.ActionsConfig == nil || len(cfg.ActionsConfig) == 0 {
		return errMissingActionConfig
	}

	for _, action := range cfg.ActionsConfig {
		if action.Key == nil || *action.Key == "" {
			return errMissingActionConfigKey
		}
		if action.Action != DELETE {
			if action.FromAttribute == nil && action.ValueDefault == nil {
				return errMissingActionConfigSource
			}
		} else {
			if action.FromAttribute != nil || action.ValueDefault != nil {
				return errMissingActionDeleteParams
			}
		}
	}
	return nil

}
