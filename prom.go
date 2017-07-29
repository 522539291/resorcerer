package main

import (
	"context"
  "time"
  "fmt"
  "strconv"

	log "github.com/Sirupsen/logrus"
	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)


func track(c promapi.Client, namespace, pod, period string) error {
  api := promv1.NewAPI(c)
  log.Infof("Observing resource consumption of pod '%s' using %s", pod, promep)

  query := fmt.Sprintf("max_over_time(container_memory_usage_bytes{name=~\"%s\"}[%s])", pod, period)
  v, err := api.Query(context.Background(), query, time.Now())
  if err != nil {
    return fmt.Errorf("Can't get memory consumption from Prometheus: %s", err)
  }
  log.Debugf("PromQL result for mem: %v", v)
  mem, _ := strconv.Atoi(v.String())

  query = fmt.Sprintf("max_over_time(container_cpu_usage_seconds_total{name=~\"%s\"}[%s])", pod, period)
  v, err = api.Query(context.Background(), query, time.Now())
  if err != nil {
    return fmt.Errorf("Can't get CPU consumption from Prometheus: %s", err)
  }
  log.Debugf("PromQL result for CPU: %v", v)
  cpu := fmt.Sprintf("%sm", v.String() )

  k := fmt.Sprintf("%s", pod)
  consumption[k] = rescon{Meminbytes: mem, CPUinmillicores: cpu}
  return nil
}
