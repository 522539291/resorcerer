package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// track executes a PromQL query and stores the results
// Note: should consider different strategies here as well as tracking
// image="" label as well.
func track(c promapi.Client, namespace, pod, container, period string) error {
	api := promv1.NewAPI(c)
	log.Infof("Observing resource consumption of container '%s' in pod '%s':", container, pod)

	query := fmt.Sprintf("max_over_time(container_memory_usage_bytes{pod_name=~\"%s.*\", container_name=\"%s\"}[%s])", pod, container, period)
	v, err := api.Query(context.Background(), query, time.Now())
	if err != nil {
		return fmt.Errorf("Can't get memory consumption from Prometheus: %s", err)
	}
	log.Debugf("PromQL result for mem: %s", v.String())
	mem := strings.Trim(strings.Split(v.String(), "=>")[1], " ") // NOTE: should really check here if 'v' is valid/contains something
	mem = strings.Trim(strings.Split(mem, "@")[0], " ")

	query = fmt.Sprintf("sum(rate(container_cpu_usage_seconds_total{pod_name=~\"%s.*\", container_name=\"%s\"}[%s]))", pod, container, period)
	v, err = api.Query(context.Background(), query, time.Now())
	if err != nil {
		return fmt.Errorf("Can't get CPU consumption from Prometheus: %s", err)
	}
	log.Debugf("PromQL result for CPU: %s", v.String())
	cpu := strings.Trim(strings.Split(v.String(), "=>")[1], " ")
	cpu = strings.Trim(strings.Split(cpu, "@")[0], " ")
	// fmt.Sprintf("%sm", cpu)

	k := fmt.Sprintf("%s:%s", pod, container)
	consumption[k] = rescon{Meminbytes: mem, CPUusagesec: cpu}
	log.Infof("Stored: %+v", consumption[k])
	return nil
}
