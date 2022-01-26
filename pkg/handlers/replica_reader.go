// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright 2020 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas-netes/pkg/k8s"
	types "github.com/openfaas/faas-provider/types"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/client-go/listers/apps/v1"
	glog "k8s.io/klog"
)

// MakeReplicaReader reads the amount of replicas for a deployment
func MakeReplicaReader(defaultNamespace string, lister v1.DeploymentLister, query *PrometheusQuery) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)

		functionName := vars["name"]
		q := r.URL.Query()
		namespace := q.Get("namespace")

		usage := q.Get("usage")
		addUsage := usage == "true" || usage == "1"

		lookupNamespace := defaultNamespace

		if len(namespace) > 0 {
			lookupNamespace = namespace
		}

		s := time.Now()

		function, err := getFunctionStatus(lookupNamespace, functionName, lister)
		if err != nil {
			log.Printf("Unable to fetch service: %s %s\n", functionName, namespace)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if function == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if addUsage {
			mixInUsage(function, query)
		}

		d := time.Since(s)
		log.Printf("Replicas: %s.%s, (%d/%d) %dms\n", functionName, lookupNamespace, function.AvailableReplicas, function.Replicas, d.Milliseconds())

		functionBytes, err := json.Marshal(function)
		if err != nil {
			glog.Errorf("Failed to marshal function: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to marshal function"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)
	}
}

// getFunctionStatus returns a function/service or nil if not found
func getFunctionStatus(functionNamespace string, functionName string, lister v1.DeploymentLister) (*types.FunctionStatus, error) {

	item, err := lister.Deployments(functionNamespace).
		Get(functionName)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	if item != nil {
		function := k8s.AsFunctionStatus(*item)
		if function != nil {
			return function, nil
		}
	}

	return nil, fmt.Errorf("function: %s not found", functionName)
}

func mixInUsage(function *types.FunctionStatus, query *PrometheusQuery) {

	cpu := queryCPU(function.Name, function.Namespace, query)
	memory := queryMemory(function.Name, function.Namespace, query)

	if cpu != nil || memory != nil {
		function.Utilisation = types.FunctionUtilisation{}

		if cpu != nil {
			function.Utilisation.CPU = *cpu
		}
		if memory != nil {
			function.Utilisation.TotalMemoryBytes = *memory
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
