package main

import (
	"net/http"
	"os"

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
	releaseVersion string
	// promep represents the Prometheus endpoint
	promep string
	// targetns represents the target namespace
	targetns string
	// consumption holds top RAM/CPU consumption, using $POD:$CONTAINER as key
	consumption map[string]rescon
)

func init() {
	// try to get the config via environment variables and
	// if that's not possible set sensible defaults:
	if envd := os.Getenv("DEBUG"); envd != "" {
		log.SetLevel(log.DebugLevel)
	}
	if promep = os.Getenv("PROM_API"); promep == "" {
		log.Printf("Can't find a PROM_API environment variable, using default prometheus.resorcerer.svc:9090")
		promep = "prometheus.resorcerer.svc:9090"
	}
	if targetns = os.Getenv("TARGET_NAMESPACE"); targetns == "" {
		log.Printf("Can't find a TARGET_NAMESPACE environment variable, using default resorcerer")
		targetns = "resorcerer"
	}
	consumption = make(map[string]rescon)
}

func main() {
	host := "0.0.0.0:"
	port := "8080"
	r := mux.NewRouter()
	r.HandleFunc("/observation/{pod}/{container}", observe).Methods("GET")
	r.HandleFunc("/recommendation/{pod}/{container}", getrec).Methods("GET")
	r.HandleFunc("/adjustment/{pod}/{container}", adjustment).Methods("POST")
	r.HandleFunc("/targets", targets).Methods("GET")
	r.HandleFunc("/version", version).Methods("GET")
	log.Printf("Serving API from: %s%s/v1", host, port)
	srv := &http.Server{Handler: r, Addr: host + port}
	log.Fatal(srv.ListenAndServe())
}
