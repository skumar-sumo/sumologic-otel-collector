package luaprocessor

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/model/pdata"

	glua "github.com/RyouZhang/go-lua"
)

type luaProcessor struct {
	function string
	script   string
}

func newLuaProcessor(cfg *Config) *luaProcessor {
	return &luaProcessor{
		function: cfg.Function,
		script:   cfg.Script,
	}
}

// ProcessMetrics processes metrics
func (lp *luaProcessor) ProcessMetrics(ctx context.Context, md pdata.Metrics) (pdata.Metrics, error) {
	// TODO: add processor logic here
	fmt.Println("***Hello from Lua metrics processor***")

	converter := NewMetricsLuaConverter(md)
	params := converter.ConvertToLua()

	fmt.Println("Lua processor, script: ", lp.script, ", function: ", lp.function)
	res, err := glua.NewAction().WithScriptPath(lp.script).WithEntrypoint(lp.function).AddParam(params).Execute(context.Background())
	if err != nil {
		return md, err
	}
	fmt.Println("*** lua script returned", res)

	return converter.ConvertFromLua(res)
}

// ProcessTraces processes traces
func (lp *luaProcessor) ProcessTraces(ctx context.Context, md pdata.Traces) (pdata.Traces, error) {
	// TODO: add processor logic here
	fmt.Println("***Hello from Lua traces processor***")

	return md, nil
}

// ProcessLogs processes logs
func (lp *luaProcessor) ProcessLogs(ctx context.Context, md pdata.Logs) (pdata.Logs, error) {
	// TODO: add processor logic here
	fmt.Println("***Hello from Lua logs processor***")

	return md, nil
}
