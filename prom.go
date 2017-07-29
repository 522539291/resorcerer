package main

import (
	"context"
  "time"
"fmt"
	log "github.com/Sirupsen/logrus"
	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)


func track(c promapi.Client, namespace,pod string) {
  api := promv1.NewAPI(c)
  log.Infof("Observing resource consumption of pod %s using %s", pod, api)
    query := "container_memory_usage_bytes"
    v, err := api.Query(context.Background(), query, time.Now())
    if err != nil {
      log.Errorf("Can't get data from Prometheus: %s", err)
    }
    log.Debugf("PromQL result: %s", v)
    // store top value for mem/cpu here
    k := fmt.Sprintf("%s", pod)
    consumption[k] = rescon{Meminbytes: 100, CPUinmillicores: "200m"}
}
