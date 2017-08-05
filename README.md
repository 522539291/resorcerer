# resorcerer—resource sorcerer

This is an experimental implementation, a demonstrator, for automated resource consumption management in Kubernetes. To learn more about the motivation, history and background of `resorcerer`, read the blog post [Container resource consumption—too important to ignore](https://medium.com/@mhausenblas/container-resource-consumption-too-important-to-ignore-7484609a3bb7).

- [Goals and non-goals of resorcerer](#goals-and-non-goals-of-resorcerer)
- [Setup](#setup)
	- [Launch Prometheus](#launch-prometheus)
	- [Launch resorcerer](#launch-resorcerer)
- [Usage](#usage)
	- [HTTP API](#http-api)
  - [Development](#development)
- [Architecture](background.md#architecture)

## Goals and non-goals of resorcerer

With `resorcerer` we want to contribute to the advancement of Kubernetes [Vertical Pod Autoscalers(https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler) (VPA). It's an experimental implementation, allowing users to learn more about their container resource consumption footprint. Also, `resorcerer` is an opinionated implementation, making a number of assumptions:

1. Prometheus is available in the cluster.
1. You can run `resorcerer` in privileged mode, as it needs access to all necessary metrics from the kubelet cAdvisor as well as the rights to update standalone and supervised pods.
1. As much as possible should happen automatically—that's where the magic/sorcerer comes into play ;)

For convenience, below you find a simple way to deploy `resorcerer` along with Prometheus into OpenShift (see the **Setup** section), however, `resorcerer` itself will work on any Kubernetes cluster.

## Setup

The following assumes OpenShift 1.5 or later.

### Launch Prometheus

Following the nice tutorial by [Robust Perception](https://www.robustperception.io/openshift-and-prometheus/)
we set up our Prometheus environment as follows (or you simply launch `deployments/promup.sh` which includes the following steps):

```
$ oc new-project resorcerer
$ oc create configmap prom-config-cm --from-file=deployments/prometheus.yaml
$ oc apply -f deployments/all-prometheus.yaml
$ oc expose service prometheus
$ oc get routes
NAME                HOST/PORT                                     PATH      SERVICES     PORT       TERMINATION   WILDCARD
routes/prometheus   prometheus-resorcerer.192.168.99.100.nip.io             prometheus   9090-tcp                 None
```

From the `oc routes` output above you see where your Prometheus dashboard is, `http://prometheus-resorcerer.192.168.99.100.nip.io/graph`
for me (since I'm using Minishift for development):

![Prometheus dashboard](img/prom-screen-shot.png)

If you're not familiar with the Prometheus [query language](https://prometheus.io/docs/querying/basics/), now is a good time to learn it.
You can also check out exemplary [PromQL examples](background.md#promql-examples) we're using in `resorcerer`.

Also, to verify the setup you might want to use `curl http://prometheus-resorcerer.192.168.99.100.nip.io/api/v1/targets`;
see also this example of a [targets JSON](dev/example-targets.json) result file.

### Launch resorcerer

In a nutshell, `resorcerer` is a namespace-level infrastructure daemon that you can ask to observe pods and get recommendations for the resource consumption.

Launch `resorcerer` as follows (note: in vanilla Kubernetes, replace the `oc apply` with `kubectl apply`, if on OpenShift then simple run `deployments/resorcererup.sh`):

```
$ oc apply -f deployments/all-resorcerer.yaml
```

If you're using OpenShift as the target deployment platform you can also do the following to make the `resorcerer` HTTP API
accessible from outside of the cluster:

```
$ oc expose deployment resorcerer --port=8080
$ oc expose service resorcerer
```

You might also want to deploy a couple of apps so that you can try out various pods.
If you want an on-ramp for that, simply use `deployments/genworkload.sh` to populate the cluster with some pods you can use.

When done, don't forgot to clean up with `oc delete project resorcerer`, which will remove all the resources including the project/namespace itself.

## Usage

Explains how to use the `resorcerer` HTTP API.

### HTTP API

The following is against a base URL `$RESORCERER`—something like `http://resorcerer-resorcerer.192.168.99.100.nip.io`—and operating in the default `resorcerer` namespace. You can perform the operations as described in the following.

#### Observation

To observe $CONTAINER in $POD for period $PERIOD (with valid time units "s", "m", and "h") do:

```
GET /observation/$POD/$CONTAINER?period=$PERIOD
```

For example:

```
$ http $RESORCERER/observation/nginx/nginx?period=1h
HTTP/1.1 200 OK

Successfully observed container 'nginx' in pod 'nginx'
```

#### Recommendations

To get a resource consumption recommendation for $CONTAINER in $POD do:

```
GET /recommendation/$POD/$CONTAINER
```

For example:

```
$ http $RESORCERER/recommendation/nginx/nginx
HTTP/1.1 200 OK

{
    "container": "nginx",
    "pod": "nginx",
    "resources": {
        "cpu": "0.00387342659078205",
        "mem": "9109504"
    }
}
```

#### Adjustments

To adjust the resource consumption for $CONTAINER in $POD do:

```
$ http POST /adjustment/$POD/$CONTAINER cpu=$CPUSEC mem=$MEMINBYTES
```

Note that above means effectively manipulating `spec.containers[].resources.limits/requests` and causing a new pod being launched.
There are ATM no in-place adjustments possible since the primitives are not in place yet, cf. [ISSUE-5774](https://github.com/kubernetes/kubernetes/issues/5774).

Also, note the following (tested for K8S 1.5):

- the minimum CPU seconds limit allowed is 1 millicore, that is, the minimum `$CPUSEC` you can set is `0.001`.
- the minimum memory limit allowed is 4MB, that is, the minimum `$MEMINBYTES` you can set is `4000000`.

Note that `resorcerer` enforces those minimum limits, that is, even if you try to set lower limits they will be ignored and above limits will be used in their place.

For example, to set limits to 230 millicores and 50MB:

```
$ http POST $RESORCERER/adjustment/nginx/nginx cpu=0.230 mem=50000000
HTTP/1.1 200 OK

Pod 'nginx' is supervised by 'Deployment/RS' - now updated it with new resource limits
```

### Development

The `resorcerer` daemon is shipped as a container, that is, we don't do binary releases.
If you want to extend `resorcerer`, you'll need at least Go 1.8 as well as [dep](https://github.com/golang/dep)
to build it. Note that if you don't have `dep` installed, do `go get -u github.com/golang/dep/cmd/dep` now to install it
and then `dep ensure` to install the dependencies. The result is an additional `vendor/` directory.

Now you can use the Makefile to build the binaries and container image as you see fit, for example:

```
$ go build      # build binary for your platform, for local testing
$ make release  # cut a new release (only maintainers, requires push access to quay.io)
```

Note that when you execute the `resorcerer` binary locally, for development purposes, you want to set
something like `export PROM_API=http://prometheus-resorcerer.192.168.99.100.nip.io` to let it know where
to find Prometheus.

## Kudos

We'd like to thank the following people for their support:

- Julius Volz (Prometheus)
- Stefan Schimanski (Kube API)
