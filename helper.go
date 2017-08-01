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

// adjust updates the resource limits of container in pod,
// setting spec.containers[].resources.limits/request to the
// limits as specified in lim.
func adjust(namespace, pod, container string, lim rescon) (string, error) {
	// set up the Kube API access:
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}
	// check if we're dealing with a standalone pod or supervised pod.
	// note that since 'owner ref' is not generally available if GC is
	// not enabled we're using the annotations here to figure if it's
	// a supervised pod or not:
	po, err := cs.CoreV1().Pods(namespace).Get(pod)
	if err != nil {
		return "", err
	}
	anno := po.GetAnnotations()
	owner, ok := anno["kubernetes.io/created-by"]
	if !ok { // standalone pod
		// TBD: replace the standalone pod
		return fmt.Sprintf("Seems like '%s' is a standalone pod, replaced it", pod), nil
	}
	// supervised pod (RC or Deployment/RS), now owner should contain something like:
	// {"kind":"SerializedReference","apiVersion":"v1","reference":{"kind":"ReplicationController", ...}
	supervisor := ""
	switch {
	case strings.Contains(owner, "ReplicationController"): // this means a DC, OpenShift-specific
		supervisor = "DeploymentConfig"
	case strings.Contains(owner, "ReplicaSet"): // this means a Deployment; not yet implemented
		supervisor = "Deployment"
	default:
		return fmt.Sprintf("Yeah man, so I don't really know that kind of supervisor, sorry …"), nil
	}
	// now replace pod with one that has the new resource limits, see also:
	// https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#how-pods-with-resource-limits-are-run
	// depl, err := cs.CoreV1()
	// depl.Spec.Template.Annotations = map[string]string{ // add annotations
	// 	"foo": "bar",
	// }
	// _, err = deploymentsClient.Update(depl)
	// if errors.IsConflict(err) {
	// 	ret = fmt.Sprintf("Can't update, encountered conflict: %s", err)
	// }
	return fmt.Sprintf("Seems like '%s' is a pod supervised by a %s, updated it", pod, supervisor), nil
}
