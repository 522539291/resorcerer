package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	promapi "github.com/prometheus/client_golang/api"
)

func observe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pod := vars["pod"]
	period := r.URL.Query().Get("period")
	c, err := promapi.NewClient(promapi.Config{Address: promep})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Errorf("Can't connect to Prometheus: %s", err)
		return
	}
	err = track(c, targetns, pod, period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
		return
	}
	fmt.Fprintf(w, "Successfully observed pod %s", pod)
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
