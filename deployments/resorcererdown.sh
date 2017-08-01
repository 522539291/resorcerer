#!/usr/bin/env bash

oc delete deployment resorcerer
oc delete svc resorcerer
oc delete route resorcerer
