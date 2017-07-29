package main

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	promapi "github.com/prometheus/client_golang/api"
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Errorf("Can't connect to Prometheus: %s", err)
		return
	}
	go track(c, targetns, pod)
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
	_ = json.NewEncoder(w).Encode(recres)
}

func setrec(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pod := vars["pod"]
	container := vars["container"]
	log.Infof("Updating resources for container %s in pod %s", container, pod)
}
