package main

import (
	"fmt"
	"strconv"
	"strings"

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
	case ispod(cs, pod): // a standalone pod (example: twocontainers)
		po, err := cs.CoreV1().Pods(namespace).Get(pod)
		if err != nil {
			return "", err
		}
		t := podwithlimits(lim)
		po.Spec = t.Spec
		_, err = cs.CoreV1().Pods(namespace).Update(po)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Pod '%s' is an unsupervised pod - replaced it with new resource limits", pod), nil
	case isdeployment(cs, pod): // an extensionsV1Beta-supervised pod; a Deployment/RS (example: nginx)
		depl, err := cs.ExtensionsV1beta1().Deployments(namespace).Get(pod)
		if err != nil {
			return "", err
		}
		// set new resource limits:
		depl.Spec.Template = podwithlimits(lim)
		_, err = cs.ExtensionsV1beta1().Deployments(namespace).Update(depl)
		if err != nil {
			return "", err
		}
		supervisor = "Deployment/RS"
	case isrc(cs, pod): // corev1-supervised pod; some sort of RC (example: simpleservice)
		rc, err := cs.CoreV1().ReplicationControllers(namespace).Get(pod)
		if err != nil {
			return "", err
		}
		// set new resource limits:
		p := podwithlimits(lim)
		rc.Spec.Template = &p
		_, err = cs.CoreV1().ReplicationControllers(namespace).Update(rc)
		if err != nil {
			return "", err
		}
		supervisor = "RC"
	default:
		return fmt.Sprintf("Dude, I don't really know that kind of supervisor, sorry"), nil
	}
	return fmt.Sprintf("Pod '%s' is supervised by '%s' - now updated it with new resource limits", pod, supervisor), nil
}

// podwithlimits creates a pod spec with new limits set as per lim.
func podwithlimits(lim rescon) v1.PodTemplateSpec {
	cpuval, _ := strconv.ParseInt(lim.CPUusagesec, 10, 64)
	memval, _ := strconv.ParseInt(lim.Meminbytes, 10, 64)
	newlim := v1.ResourceList{
		v1.ResourceCPU:    *resource.NewQuantity(cpuval, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(memval, resource.DecimalSI),
	}
	return v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits:   newlim,
						Requests: newlim,
					},
				},
			},
		},
	}
}

// supervisorinfo extracts the type, and ID of the supervisor,
// using the 'kubernetes.io/created-by' annotation, if present.
// note that since 'owner ref' is not generally available if GC is
// not enabled we're using the annotations here to figure if it's
// a supervised pod or not. note also that this is not really sustainable,
// because of: https://github.com/kubernetes/kubernetes/issues/44407
func supervisorinfo(createdby string) (stype, sid string) {
	// we would expect to have something like the following in the createdby string:
	// {"kind":"SerializedReference","apiVersion":"v1","reference":{"kind":"ReplicaSet","namespace":"resorcerer","name":"nginx-133933678", ...}}
	oreference := strings.Split(createdby, "reference\":")[1]
	switch {
	case strings.Contains(oreference, "ReplicationController"):
		stype = "ReplicationController"
	case strings.Contains(oreference, "ReplicaSet"):
		stype = "ReplicaSet"
	default:
		stype = "Unknown"
	}
	name := strings.Split(oreference, "name\":\"")[1]
	sid = strings.Split(name, "\"")[0]
	return stype, sid
}

func ispod(cs *kubernetes.Clientset, objname string) bool {
	_, err := cs.CoreV1().Pods("resorcerer").Get(objname)
	if err != nil {
		return false
	}
	return true
}

func isdeployment(cs *kubernetes.Clientset, objname string) bool {
	_, err := cs.ExtensionsV1beta1().Deployments("resorcerer").Get(objname)
	if err != nil {
		return false
	}
	return true
}

func isrc(cs *kubernetes.Clientset, objname string) bool {
	// the following is a horrible hack but found no other way around it for now.
	// apparently DC/RC don't have names so have to try out some variants, say,
	// if the pod is called foo, I'm trying foo-1, foo-2 up to foo-100 … not proud of it.
	for i := 1; i <= 100; i++ {
		_, err := cs.CoreV1().ReplicationControllers("resorcerer").Get(fmt.Sprintf("%s-%d", objname, i))
		if err == nil {
			return true
		}
	}
	return false
}
