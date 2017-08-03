package main

import (
	"fmt"
	"strconv"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

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
	supervisor := ""
	switch {
	case ispod(cs, namespace, pod): // a corev1-standalone pod (example: twocontainers)
		return updatepod(cs, namespace, pod, container, lim)
	case isdeployment(cs, namespace, pod): // an extensionsV1Beta-supervised pod; a Deployment/RS (example: nginx)
		s, err := updatedeployment(cs, namespace, pod, container, lim)
		if err != nil {
			return "", err
		}
		supervisor = s
	case isrc(cs, namespace, pod): // a corev1-supervised pod; some sort of RC (example: simpleservice)
		s, err := updaterc(cs, namespace, pod, container, lim)
		if err != nil {
			return "", err
		}
		supervisor = s
	default:
		return fmt.Sprintf("Dude, I don't really know that kind of supervisor, sorry"), nil
	}
	return fmt.Sprintf("Pod '%s' is supervised by '%s' - now updated it with new resource limits", pod, supervisor), nil
}

// withlimits creates resource requirements with limits set as per lim.
func withlimits(lim rescon) v1.ResourceRequirements {
	lim = ensurelow(lim)
	cpuval, _ := resource.ParseQuantity(lim.CPUusagesec)
	memval, _ := resource.ParseQuantity(lim.Meminbytes)
	newlim := v1.ResourceList{
		v1.ResourceCPU:    cpuval,
		v1.ResourceMemory: memval,
	}
	return v1.ResourceRequirements{
		Limits:   newlim,
		Requests: newlim,
	}
}

// ensurelow makes sure that the resource limitations
// don't fall below the minimal allowed ones, that is
// CPU must be at least 1 millicore (== 0.001) and
// memory must be at least 4MB (==4000000).
func ensurelow(lim rescon) rescon {
	cpu, _ := strconv.ParseFloat(lim.CPUusagesec, 64)
	mem, _ := strconv.ParseInt(lim.Meminbytes, 10, 64)
	sanitizedlim := lim
	if cpu <= 0.001 { // must be at least 1 millicore (== 0.001)
		sanitizedlim.CPUusagesec = "0.001"
	}
	if mem < 4000000 { // must be at least 4MB (==4000000)
		sanitizedlim.Meminbytes = "4000000"
	}
	return sanitizedlim
}

func ispod(cs *kubernetes.Clientset, namespace, objname string) bool {
	_, err := cs.CoreV1().Pods(namespace).Get(objname)
	if err != nil {
		return false
	}
	return true
}

func updatepod(cs *kubernetes.Clientset, namespace, pod, container string, lim rescon) (string, error) {
	po, err := cs.CoreV1().Pods(namespace).Get(pod)
	if err != nil {
		return "", err
	}
	for i, c := range po.Spec.Containers {
		if c.Name == container {
			po.Spec.Containers[i].Resources = withlimits(lim)
		}
	}
	po.SetResourceVersion("")
	var immediately int64
	err = cs.CoreV1().Pods(namespace).Delete(pod, &v1.DeleteOptions{GracePeriodSeconds: &immediately})
	if err != nil {
		return "", err
	}
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			fmt.Print(".")
		}
	}()
	time.Sleep(10 * time.Second)
	ticker.Stop()
	_, err = cs.CoreV1().Pods(namespace).Create(po)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Pod '%s' is an unsupervised pod - replaced it with new resource limits", pod), nil
}

func isdeployment(cs *kubernetes.Clientset, namespace, objname string) bool {
	_, err := cs.ExtensionsV1beta1().Deployments(namespace).Get(objname)
	if err != nil {
		return false
	}
	return true
}

func updatedeployment(cs *kubernetes.Clientset, namespace, pod, container string, lim rescon) (string, error) {
	depl, err := cs.ExtensionsV1beta1().Deployments(namespace).Get(pod)
	if err != nil {
		return "", err
	}
	for i, c := range depl.Spec.Template.Spec.Containers {
		if c.Name == container {
			depl.Spec.Template.Spec.Containers[i].Resources = withlimits(lim)
		}
	}
	_, err = cs.ExtensionsV1beta1().Deployments(namespace).Update(depl)
	if err != nil {
		return "", err
	}
	return "Deployment/RS", nil
}

func isrc(cs *kubernetes.Clientset, namespace, objname string) bool {
	// the following is a horrible hack but found no other way around it for now.
	// apparently DC/RC don't have names so have to try out some variants, say,
	// if the pod is called foo, I'm trying foo-1, foo-2 up to foo-100 … not proud of it.
	for i := 1; i <= 100; i++ {
		_, err := cs.CoreV1().ReplicationControllers(namespace).Get(fmt.Sprintf("%s-%d", objname, i))
		if err == nil {
			return true
		}
	}
	return false
}

func updaterc(cs *kubernetes.Clientset, namespace, pod, container string, lim rescon) (string, error) {
	var rc *v1.ReplicationController
	for i := 1; i <= 100; i++ {
		tmprc, err := cs.CoreV1().ReplicationControllers(namespace).Get(fmt.Sprintf("%s-%d", pod, i))
		if err == nil {
			rc = tmprc
		}
	}
	if rc == nil {
		return "", fmt.Errorf("Can't find any RC for %s", pod)
	}
	for i, c := range rc.Spec.Template.Spec.Containers {
		if c.Name == container {
			rc.Spec.Template.Spec.Containers[i].Resources = withlimits(lim)
		}
	}
	_, err := cs.CoreV1().ReplicationControllers(namespace).Update(rc)
	if err != nil {
		return "", err
	}
	return "RC", nil
}
