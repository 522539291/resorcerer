package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func observe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pod := vars["pod"]
	periodparam := r.URL.Query().Get("period")
	period, err := time.ParseDuration(periodparam)
	if err != nil {
		period = 5 * time.Second
	}
	c, err := promapi.NewClient(promapi.Config{Address: promep})
	if err != nil {
		log.Errorf("Can't connect to Prometheus: %s", err)
	}
	api := promv1.NewAPI(c)
	log.Infof("Observing resource consumption using %s", api)
	go func() {
		// p, err := listpods(targetns)
		// if err != nil {
		// 	log.Errorf("Can't list pods in %s: %s", targetns, err)
		// 	return
		// }
		// log.Debugf("%s", p)
		container := "1"
		query := "container_memory_usage_bytes"
		v, err := api.Query(context.Background(), query, time.Now())
		if err != nil {
			log.Errorf("Can't get data from Prometheus: %s", err)
			return
		}
		log.Debugf("%s", v)
		// store top value for mem/cpu here
		k := fmt.Sprintf("%s-%s", pod, container)
		consumption[k] = rescon{Meminbytes: 100, CPUinmillicores: "200m"}
	}()
	timeout := time.After(period)
	pollinterval := 500 * time.Millisecond
	for {
		select {
		case <-timeout:
			log.Info("Observation period over")
			return
		default:
			log.Infof(".")
		}
		time.Sleep(pollinterval)
	}
}

func getrec(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pod := vars["pod"]
	log.Infof("Serving recommendation for pod %s", pod)
	cc := []concon{concon{Container: "c1", Resources: rescon{Meminbytes: 1234, CPUinmillicores: "200m"}}}
	recres := recresponse{
		Pod:             pod,
		Recommendations: cc,
	}
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	log.Error(err)
	// 	return
	// }
	_ = json.NewEncoder(w).Encode(recres)
}

func setrec(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pod := vars["pod"]
	container := vars["container"]
	log.Infof("Updating resources for container %s in pod %s", container, pod)
}
