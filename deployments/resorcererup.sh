#!/usr/bin/env bash

oc apply -f deployments/all-resorcerer.yaml
oc expose deployment resorcerer --port=8080
oc expose service resorcerer
oc describe routes/resorcerer
