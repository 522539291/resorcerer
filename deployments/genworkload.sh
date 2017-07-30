#!/usr/bin/env bash

oc new-app mhausenblas/simpleservice:0.5.0
oc expose dc simpleservice --port=9876
oc create -f https://raw.githubusercontent.com/mhausenblas/kbe/master/specs/pods/pod.yaml
