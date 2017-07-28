package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	var promep string
	if promep = os.Getenv("PROM_API"); promep == "" {
		log.Printf("Can't find a PROM_API environment variable, using default prometheus.resorcerer.svc:9090")
		promep = "prometheus.resorcerer.svc:9090"
	}
	c, err := promapi.NewClient(promapi.Config{Address: promep})
	if err != nil {
		log.Fatalf("Can't connect to Prometheus: %s", err)
	}
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigs
		log.Printf("\nDone observing, closing down …")
		done <- true
	}()
	api := promv1.NewAPI(c)
	log.Printf("Observing resource consumption using %v", api)
	go func() {
		for {
			p, err := listpods()
			if err != nil {
				log.Printf("Can't get pod list: %s", err)
				continue
			}
			log.Printf("%s", p)
			query := "container_memory_usage_bytes"
			v, err := api.Query(context.Background(), query, time.Now())
			if err != nil {
				log.Printf("Can't get data from Prometheus: %s", err)
				time.Sleep(time.Second * 2)
				continue
			}
			log.Printf("%s", v)
			time.Sleep(time.Second * 2)
		}
	}()
	<-done
}

func listpods() ([]string, error) {
	var po []string
	config, err := rest.InClusterConfig()
	if err != nil {
		return po, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return po, err
	}
	pods, err := clientset.CoreV1().Pods("resorcerer").List(metav1.ListOptions{})
	if err != nil {
		return po, err
	}
	for _, p := range pods.Items {
		po = append(po, p.GetName())
	}
	return po, nil
}
