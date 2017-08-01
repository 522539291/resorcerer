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
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(recres)
}

func adjustment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pod := vars["pod"]
	container := vars["container"]

	decoder := json.NewDecoder(r.Body)
	var adjustrescon rescon
	err := decoder.Decode(&adjustrescon)
	if err != nil {
		mreq := "The resource constraints request is malformed"
		http.Error(w, mreq, http.StatusBadRequest)
		log.Error(mreq)
		return
	}
	log.Infof("Updating resources for container %s in pod %s", container, pod)
	res, err := adjust("resorcerer", pod, container, adjustrescon)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		log.Errorf("%s", err)
		return
	}
	fmt.Fprintf(w, "%s", res)
}

func targets(w http.ResponseWriter, r *http.Request) {
	pods, err := listpods("resorcerer")
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
		log.Errorf("%s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Pod []string `json:"pod"`
		// Containers []string `json:"containers"`
	}{
		pods,
	})
}

func version(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "the one and only resorcerer in version %s", releaseVersion)
}
