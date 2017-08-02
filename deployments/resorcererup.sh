#!/usr/bin/env bash

oc create serviceaccount vpa
oc policy add-role-to-user edit --serviceaccount=vpa
oc apply -f deployments/all-resorcerer.yaml
oc expose deployment resorcerer --port=8080
oc expose service resorcerer
oc describe routes/resorcerer
