package main

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type rescon struct {
	Meminbytes  string `json:"mem"`
	CPUusagesec string `json:"cpu"`
}

type recresponse struct {
	Pod       string `json:"pod"`
	Container string `json:"container"`
	Resources rescon `json:"resources"`
}

var (
	// promep represents the Prometheus endpoint
	promep string
	// targetns represents the target namespace
	targetns string
	// consumption holds top RAM/CPU consumption, using $POD:$CONTAINER as key
	consumption map[string]rescon
)

func init() {
	loadenvs()
	consumption = make(map[string]rescon)
}

func main() {
	host := "0.0.0.0:"
	port := "8080"
	r := mux.NewRouter()
	r.HandleFunc("/observation/{pod}/{container}", observe).Methods("GET")
	r.HandleFunc("/recommendation/{pod}/{container}", getrec).Methods("GET")
	r.HandleFunc("/recommendation/{pod}/{container}", setrec).Methods("POST")
	log.Printf("Serving API from: %s%s/v1", host, port)
	srv := &http.Server{Handler: r, Addr: host + port}
	log.Fatal(srv.ListenAndServe())
}
