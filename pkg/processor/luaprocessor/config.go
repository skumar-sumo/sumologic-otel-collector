package luaprocessor

import "go.opentelemetry.io/collector/config"

type Config struct {
	*config.ProcessorSettings `mapstructure:"-"`
	Function                  string `mapstructure:"function"`
	Script                    string `mapstructure:"script"`
}
