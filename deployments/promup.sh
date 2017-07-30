#!/usr/bin/env bash

oc new-project resorcerer
oc adm policy add-cluster-role-to-user cluster-reader system:serviceaccount:resorcerer:default
oc create configmap prom-config-cm --from-file=deployments/prometheus.yml
oc apply -f deployments/all-prometheus.yaml
oc expose service prometheus

sleep 2
oc get routes,svc,dc
