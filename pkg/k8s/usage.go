package k8s

import (
	"fmt"
	"log"
	"net/url"
	"strconv"

	types "github.com/openfaas/faas-provider/types"
)

func MixInUsage(function *types.FunctionStatus, query *PrometheusQuery) {

	cpu := queryCPU(function.Name, function.Namespace, query)
	memory := queryMemory(function.Name, function.Namespace, query)

	if cpu != nil || memory != nil {
		function.Usage = &types.FunctionUsage{}

		if cpu != nil {
			function.Usage.CPU = *cpu
		}
		if memory != nil {
			function.Usage.TotalMemoryBytes = *memory
		}

	}
}

func queryCPU(name, namespace string, query *PrometheusQuery) *float64 {
	var val *float64
	// Convert from nano CPU to milli CPU
	q := fmt.Sprintf(`sum( rate ( pod_cpu_usage_seconds_total {function_name="%s"}[1m] ) * 1000 ) by (function_name)`,
		name+"."+namespace)

	results, err := query.Fetch(url.QueryEscape(q))
	if err != nil {
		log.Printf("Error querying Prometheus for %s, error: %s\n", q, err.Error())
		return nil
	}

	if len(results.Data.Result) > 0 && len(results.Data.Result[0].Value) > 0 {

		v := results.Data.Result[0].Value[1].(string)
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			val = &f
		}
	}
	return val
}

func queryMemory(name, namespace string, query *PrometheusQuery) *float64 {
	var val *float64
	q := fmt.Sprintf(`ceil(sum( pod_memory_working_set_bytes {function_name="%s"}) by (function_name)) `,
		name+"."+namespace)

	results, err := query.Fetch(url.QueryEscape(q))
	if err != nil {
		log.Printf("Error querying Prometheus for %s, error: %s\n", q, err.Error())
		return nil
	}

	if len(results.Data.Result) > 0 && len(results.Data.Result[0].Value) > 0 {
		v := results.Data.Result[0].Value[1].(string)
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			val = &f
		}
	}
	return val
}
