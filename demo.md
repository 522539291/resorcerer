# Demo

## Preparation

First, we make sure Prometheus, `resorcerer` and some apps are up and running:

```
deployments/promup.sh
deployments/resorcererup.sh
deployments/genworkload.sh
export $RESORCERER=http://resorcerer-resorcerer.192.168.99.100.nip.io
```

Now we can head over to the [Prometheus dashboard](http://prometheus-resorcerer.192.168.99.100.nip.io/graph) and have a look at metrics.

## Without limits

Next, we put some load on NGINX, without resource limits:

```
oc create -f deployments/loadgen.yaml
oc get jobs

# after the load test run has completed, clean up:
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
