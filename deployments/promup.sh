#!/usr/bin/env bash

oc new-project resorcerer
oc policy add-role-to-user admin system:serviceaccount:resorcerer:default
#oc policy who-can get node

oc apply -f deployments/all-cadvisor.yaml
oc create configmap prom-config-cm --from-file=deployments/prometheus.yml
oc apply -f deployments/all-prometheus.yaml
oc expose service cadvisor
oc expose service prometheus

sleep 2
oc get routes,svc,dc
