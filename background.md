# Background

In this document we discuss the `resorcerer` architecture, some more details about the way Prometheus is used to do the heavy lifting concerning the resource consumption recommendations as well as provide further reading material relevant for this topic.

## Architecture

On a high level, the `resorcerer` architecture looks as follows:

![resorcerer architecture](img/resorcerer-arch.jpg)

Notice the distinction of 'system land' on the left-hand side and the 'user land' on the right-hand side: this means that `resorcerer` along with Prometheus is considered part of the infrastructure, not accessible directly to end-users (developers). Prometheus scrapes the metrics (CPU, memory and many many more) using the `kubelet`-internal cAdvisor and makes it available to query them for `resorcerer`.

In user land we have four apps running, with a varying number of pods and some of the pods have one, one has two containers and `app 1 `has even three containers running. Also, note the `/namespace` in system land, essentially reminding us that `resorcerer` is a namespace-level infra daemon.

Now, when a user starts interacting with `resorcerer` using the [HTTP API](../#http-api), two things need to be known: the pod name and the container name. If there's only a single container in a pod then the latter equals the former. Equipped with this, a typical sequence (for container `c1` in pod `p1`) would be:

```
http $RESORCERER/observation/p1/c1?period=1h
http $RESORCERER/recommendation/p1/c1
http POST $RESORCERER/adjustment/p1/c1 cpu=0.5 mem=200000000
```

Above you see a full control loop: observing, and then, based on the recommendation received, adjusting the container resource consumption settings, formally called [resource requests and limits](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/).

For the observation/recommendation phase the following holds true:

- Parameterized [PromQL queries](https://github.com/mhausenblas/resorcerer/blob/master/prom.go#L17) are used to come up with a recommendation.
- At least once `/observation` must be called to trigger the recommendation creation and after that one or more `/recommendation` calls can be done in an idempotent manner.
- Since `resorcerer` is stateless, users are themselves responsible to store recommendation results in a persistent way, for example for long-term studies.

For the adjustment phase the following holds true:

- Overall, the [algorithm](https://github.com/mhausenblas/resorcerer/blob/master/adjuster.go#L39) first tries to determine the ownership question (standalone, RC-supervised, RS-supervised) and then apply respective update strategies with the new limits.
- Formally, `spec.containers[].resources.limits/request` is set on a per-container basis with `limits==request`.
- The minimum allowed values, at least 1 millicore for CPU time and
 4MB for memory are enforced by `resorcerer`.
- A standalone pod will be deleted and a new pod (with the respective resource limits) will be created instead. While this is perfectly fine for stateless applications, it's the responsibility of the user to make sure the data is replicated/copied through external means (networked volume or container storage solutions such as StorageOS or REX-Ray) in case of stateful services that are not cloud-native. For example, you're good with Cassandra or Kafka since those systems know how to shard their data but you need to invest some thoughts how to deal with, say, a MySQL database.

Now that you know about the high-level workings, let's have a closer look at some PromQL queries that one can use to asses resource consumption.

## PromQL examples

Aggregate CPU usage for all containers in pod `nginx` over the last 10 minutes:

```
sum(rate(container_cpu_usage_seconds_total{container_name="nginx"}[10m]))
```

Aggregate CPU usage for pods that names start with `simple` over the last 3 minutes:

```
sum(rate(container_cpu_usage_seconds_total{pod_name=~"simple.+", container_name="POD"}[3m])) without (cpu)
```

Maximum value memory usage in bytes over the last 5 minutes for container `sise` in pod `twocontainers`:

```
max_over_time(container_memory_usage_bytes{pod_name="twocontainers", container_name="sise"}[5m])
```

The 99 percentile of the cumulative CPU time consumed for CPU30 in seconds over the last 60 seconds:

```
quantile_over_time(0.99,container_cpu_usage_seconds_total{cpu="cpu30"}[60s])
```

Average Resident Set Size (RSS), excl. swapped out memory:

```
avg(container_memory_rss)
```

The 5 largest RSS entries:

```
topk(5,container_memory_rss)
```

## Resources

- [Hands on: Monitoring Kubernetes with Prometheus](https://coreos.com/blog/monitoring-kubernetes-with-prometheus.html)
- [Monitoring Kubernetes with Prometheus (Kubernetes Ireland, 2016)](https://www.slideshare.net/brianbrazil/monitoring-kubernetes-with-prometheus-kubernetes-ireland-2016)
- [Kubernetes service discovery](https://prometheus.io/docs/operating/configuration/#<kubernetes_sd_config>) configuration parameters (Prometheus docs)
- [metrics cAdvisor](https://github.com/google/cadvisor/blob/master/metrics/prometheus.go) source_labels
- [Prometheus Ops Metrics Example](https://github.com/openshift/origin/tree/master/examples/prometheus) from `openshift/origin`
- [wkulhanek/openshift-prometheus](https://github.com/wkulhanek/openshift-prometheus)
- [waynedovey/openshift-prometheus](https://github.com/waynedovey/openshift-prometheus)
