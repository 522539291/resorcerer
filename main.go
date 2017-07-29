package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

type rescon struct {
	Meminbytes      int    `json:"mem"`
	CPUinmillicores string `json:"cpu"`
}

var (
	promep, targetns string
	// consumption captures the top RAM/CPU consumption, using POD-CONTAINER as key
	consumption map[string]rescon
)

func init() {
	loadenvs()
	consumption = make(map[string]rescon)
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigs
		log.Printf("\nDone observing, closing down …")
		done <- true
	}()
	host := "0.0.0.0:"
	port := "8080"
	r := mux.NewRouter()
	r.HandleFunc("/observation", observe).Methods("GET")
	r.HandleFunc("/recommendation", getrec).Methods("GET")
	r.HandleFunc("/recommendation", setrec).Methods("POST")
	log.Printf("Serving API from: %s%s/v1", host, port)
	srv := &http.Server{Handler: r, Addr: host + port}
	log.Fatal(srv.ListenAndServe())
	<-done
}

func observe(w http.ResponseWriter, r *http.Request) {
	c, err := promapi.NewClient(promapi.Config{Address: promep})
	if err != nil {
		log.Errorf("Can't connect to Prometheus: %s", err)
	}
	api := promv1.NewAPI(c)
	log.Infof("Observing resource consumption using %v", api)
	go func() {
		// p, err := listpods(targetns)
		// if err != nil {
		// 	log.Errorf("Can't list pods in %s: %s", targetns, err)
		// 	return
		// }
		// log.Debugf("%s", p)
		pod := "simpleservice"
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
	period := 5 * time.Second // should come via URL query parameter
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

}

func setrec(w http.ResponseWriter, r *http.Request) {

}

// loadenvs tries to get the config via environment variables and
// if that's not possible to set sensible defaults.
func loadenvs() {
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
}

// listpods returns a slice of pod names in the given namespace.
func listpods(namespace string) ([]string, error) {
	var po []string
	config, err := rest.InClusterConfig()
	if err != nil {
		return po, err
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return po, err
	}
	pods, err := cs.CoreV1().Pods(namespace).List(v1.ListOptions{})
	if err != nil {
		return po, err
	}
	for _, p := range pods.Items {
		po = append(po, p.GetName())
	}
	return po, nil
}
