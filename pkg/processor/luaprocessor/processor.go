package luaprocessor

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/model/pdata"

	glua "github.com/RyouZhang/go-lua"
)

type luaProcessor struct{}

func newLuaProcessor(cfg *Config) *luaProcessor {
	return &luaProcessor{}
}

// ProcessMetrics processes metrics
func (lp *luaProcessor) ProcessMetrics(ctx context.Context, md pdata.Metrics) (pdata.Metrics, error) {
	// TODO: add processor logic here
	fmt.Println("***Hello from Lua metrics processor***")
	fmt.Println("Lua processor is processing metrics")
	res, err := glua.NewAction().WithScript(`
	function fib(n)
		if n == 0 then
			return 0
		elseif n == 1 then
			return 1
		end
		return fib(n-1) + fib(n-2)
	end
	`).WithEntrypoint("fib").AddParam(35).Execute(context.Background())

	if err != nil {
		return md, err
	}

	fmt.Println("*** lua processor Fib", res)

	return md, nil
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
