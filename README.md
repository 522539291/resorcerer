# resorcerer—resource sorcerer

As the wise Tim Hockin stated in his KubeCon 2016 talk on 'Everything You Ever Wanted To Know About Resource Scheduling' (see [slides](https://speakerdeck.com/thockin/everything-you-ever-wanted-to-know-about-resource-scheduling-dot-dot-dot-almost) | [video](https://www.youtube.com/watch?v=nWGkvrIPqJ4)):

> Some 2/3 of the Borg users depend on autopilot, Google's internal resource consumption estimator and regulator.

Already in mid 2015 the Kubernetes community—informed by Google's `autopilot` experience—raised this [issue](https://github.com/kubernetes/kubernetes/issues/10782), and in early 2017 we got a proposal and now initial work on the resulting [Vertical Pod Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler) (VPA).

With `resorcerer` we want to contribute to the advancement of VPAs. It's a simple prototypical implementation that should allow users to learn more about their resource consumption footprint. It is an opinionated implementation, making a number of assumptions:

1. Prometheus is available in the cluster
1. You can run it in privileged mode (as access to all necessary metrics)
1. As much as possible should happen automatically (that's where the magic/sorcerer comes into play ;)

## Setup

The following assumes OpenShift 1.5 or later.

### Launch cAdvisor and Prometheus

Following the nice tutorial by [Robust Perception](https://www.robustperception.io/openshift-and-prometheus/)
we set up our Prometheus environment as follows (or you simply launch `./promup.sh` which includes the following steps):

```
$ oc new-project resorcerer
$ oc apply -f deployments/all-cadvisor.yaml
$ oc create configmap prom-config-cm --from-file=deployments/prometheus.yaml
$ oc apply -f deployments/all-prometheus.yaml
$ oc expose service cadvisor
$ oc expose service prometheus

$ oc get routes,svc,dc
NAME                HOST/PORT                                     PATH      SERVICES     PORT       TERMINATION   WILDCARD
routes/cadvisor     cadvisor-resorcerer.192.168.99.100.nip.io               cadvisor     8080-tcp                 None
routes/prometheus   prometheus-resorcerer.192.168.99.100.nip.io             prometheus   9090-tcp                 None

NAME             CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
svc/cadvisor     172.30.88.229    <none>        8080/TCP   1m
svc/prometheus   172.30.114.121   <none>        9090/TCP   42s

NAME            REVISION   DESIRED   CURRENT   TRIGGERED BY
dc/cadvisor     1          1         1         config,image(cadvisor:latest)
dc/prometheus   1          1         1         config,image(prometheus:latest)
```

From the `oc routes` output above you see where your Prometheus dashboard is, `http://prometheus-resorcerer.192.168.99.100.nip.io/graph` for me (since I'm using Minishift for development).

If you're not familiar with the  Prometheus [query language](https://prometheus.io/docs/querying/basics/), now is a good time to learn it.


### Launch the resorcerer infra cluster daemon

Launch resorcerer as follows (no HTTP API or API, it's a headless daemon):

```
$ oc project # make sure that you're in the resorcerer project
$ oc apply -f deployments/all-resorcerer.yaml
```

When done, don't forgot to clean up with `oc delete project resorcerer`.

### Development

We don't do binary releases, you need Go 1.8 and [dep](https://github.com/golang/dep) to build it for your platform. If you don't have `dep` installed yet, do `go get -u github.com/golang/dep/cmd/dep` now and then:

```
$ dep ensure
```

Now you can use the Makefile to build the binaries and container image as you see fit, for example, `make release`.

## Usage

TBD.
