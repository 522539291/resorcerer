# Background

## Architecture

![resorcerer architecture](img/resorcerer-arch.jpg)

TBD

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
