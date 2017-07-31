package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

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

func adjust(namespace, pod, container string) (string, error) {
	// 1. check if standalone pod or supervised (RC, Deployment/RS)
	// 2. replace pod with new resource limits, see also:
	// https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#how-pods-with-resource-limits-are-run

	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}
	pods, err := cs.CoreV1().Pods(namespace).List(v1.ListOptions{})
	if err != nil {
		return "", err
	}
	result := ""
	for _, p := range pods.Items {
		log.Printf("%v", p.GetOwnerReferences())
		if strings.HasPrefix(p.GetName(), pod) {
			result = fmt.Sprintf("%v", p.GetOwnerReferences())
			break
		}
	}
	return result, nil
}
