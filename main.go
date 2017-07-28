package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Need Prometheus endpoint address")
	}
	promep := os.Args[1]
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
	query := "container_memory_usage_bytes{image=~\"CONTAINER:.*\"}"
	log.Printf("Observing resource consumption with %s", query)
	go func() {
		for {
			v, err := api.Query(context.Background(), query, time.Now())
			if err != nil {
				log.Printf("Can't get data from Prometheus: %s", err)
				time.Sleep(time.Second * 2)
				continue
			}
			fmt.Printf("%s", v)
			time.Sleep(time.Second * 2)
		}
	}()
	<-done
}
