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
	container := vars["container"]
	period := r.URL.Query().Get("period")
	c, err := promapi.NewClient(promapi.Config{Address: promep})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Errorf("Can't connect to Prometheus: %s", err)
		return
	}
	err = track(c, targetns, pod, container, period)
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
	container := vars["container"]
	log.Infof("Serving recommendation for container '%s' in pod '%s'", container, pod)
	k := fmt.Sprintf("%s:%s", pod, container)
	rec, ok := consumption[k]
	if !ok {
		e := fmt.Errorf("Can't retrieve recommendation for container '%s' in pod '%s': no such entry exists!", container, pod)
		http.Error(w, fmt.Sprintf("%s", e), http.StatusBadRequest)
		log.Errorf("%s", e)
		return
	}
	recres := recresponse{
		Pod:       pod,
		Container: container,
		Resources: rec,
	}
	_ = json.NewEncoder(w).Encode(recres)
}

func setrec(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pod := vars["pod"]
	container := vars["container"]
	log.Infof("Updating resources for container %s in pod %s", container, pod)

	res, err := adjust("resorcerer", pod, container)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		log.Errorf("%s", err)
		return
	}
	fmt.Fprintf(w, "%s", res)
}

func version(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "the one and only resorcerer in version %s", releaseVersion)
}
