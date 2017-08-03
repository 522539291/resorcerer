# Demo

## Preparation

First, we make sure Prometheus, `resorcerer` and some apps are up and running:

```
deployments/promup.sh
deployments/resorcererup.sh
deployments/genworkload.sh
export $RESORCERER=http://resorcerer-resorcerer.192.168.99.100.nip.io
```

Now we can head over to the [Prometheus dashboard](http://prometheus-resorcerer.192.168.99.100.nip.io/graph) and have a look at metrics,
for example the following queries:

```
http://prometheus-resorcerer.192.168.99.100.nip.io/graph?g0.range_input=1h&g0.expr=sum(rate(container_cpu_usage_seconds_total%7Bcontainer_name%3D%22nginx%22%7D%5B10m%5D))&g0.tab=0&g1.range_input=1h&g1.expr=max_over_time(container_memory_usage_bytes%7Bcontainer_name%3D%22nginx%22%7D%5B30m%5D)&g1.tab=0

sum(rate(container_cpu_usage_seconds_total{container_name="nginx"}[10m]))

max_over_time(container_memory_usage_bytes{container_name="nginx"}[30m])
```

## Without limits

Next, we put some load on NGINX, without resource limits:

```
oc create -f deployments/loadgen.yaml
oc get jobs

# after the load test run has completed (ca. 1min), clean up:
oc delete jobs/loadgenerator
```

## Observing and getting recommendations

```
http $RESORCERER/observation/nginx/nginx?period=30m
http $RESORCERER/recommendation/nginx/nginx
```

## Observing and getting recommendations

```
http POST $RESORCERER/adjustment/nginx/nginx cpu=0.230 mem=50000000
```

## Tear down

Finally, we clean up:

```
oc delete project resorcerer
```
