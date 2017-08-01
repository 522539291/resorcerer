package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
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
// limits as specified in lim, see also:
// https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#how-pods-with-resource-limits-are-run
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
	// a supervised pod or not. Also, we need to distinguish between
	// stuff that lives in corev1 (Pod, RC) and in extensionsV1Beta
	// (Deployments, RS):
	iscore := true
	po, err := cs.CoreV1().Pods(namespace).Get(pod)
	if err != nil {
		if !strings.HasSuffix(err.Error(), "not found") { // some other error, stop right here
			return "", err
		}
		iscore = false // apparently an extensionsV1Beta supervised pod
	}
	supervisor := ""
	switch iscore {
	case true: //  either standalone or corev1-supervised pod (RC for now)
		anno := po.GetAnnotations()
		owner, ok := anno["kubernetes.io/created-by"]
		if !ok { // standalone pod (example: twocontainers)
			// TBD: replace the standalone pod
			return fmt.Sprintf("Seems like '%s' is a standalone pod, replaced it", pod), nil
		}
		// corev1-supervised pod, that is, an RC, now owner should contain something like:
		// {"kind":"SerializedReference","apiVersion":"v1","reference":{"kind":"ReplicationController", ...}
		switch {
		case strings.Contains(owner, "ReplicationController"): // this means a DC, OpenShift-specific
			supervisor = "DeploymentConfig-ReplicationController"
			// TBD (example: simpleservice)
		default:
			return fmt.Sprintf("Yeah man, so I don't really know that kind of supervisor, sorry …"), nil
		}
	case false: // an extensionsV1Beta supervised pod, that is, Deployment+RS (example: resorcerer itself)
		depl, err := cs.ExtensionsV1beta1().Deployments(namespace).Get(pod)
		if err != nil {
			return "", err
		}
		_ = depl

		cpuval, _ := strconv.ParseInt(lim.CPUusagesec, 10, 64)
		newcpu := resource.NewQuantity(cpuval, resource.DecimalSI)
		memval, _ := strconv.ParseInt(lim.Meminbytes, 10, 64)
		newmem := resource.NewQuantity(memval, resource.DecimalSI)
		// newlim := v1.ResourceList{
		// 		v1.ResourceCPU: newcpu,
		// 		v1.ResourceMemory: newmem,
		// }
		// // set new resource limits
		// depl.Spec.Template = v1.PodSpec{
		// 	Containers: []v1.Container{
		// 		{
		// 			Resources v1.ResourceRequirements{
		// 			Limits : newlim,
		// 			Requests : newlim,
		// 			}
		// 		},
		// 	},
		// }
		supervisor = "Deployment-ReplicaSet"
	}
	return fmt.Sprintf("Seems like '%s' is a pod supervised by an '%s', updated it", pod, supervisor), nil
}
