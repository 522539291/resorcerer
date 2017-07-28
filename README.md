# resorcerer—resource sorcerer

As the wise Tim Hockin stated in his KubeCon 2016 talk on 'Everything You Ever Wanted To Know About Resource Scheduling' (see [slides](https://speakerdeck.com/thockin/everything-you-ever-wanted-to-know-about-resource-scheduling-dot-dot-dot-almost) | [video](https://www.youtube.com/watch?v=nWGkvrIPqJ4)):

> Some 2/3 of the Borg users depend on autopilot, Google's internal resource consumption estimator and regulator.

Already in mid 2015 the Kubernetes community—informed by Google's `autopilot` experience—raised this [issue](https://github.com/kubernetes/kubernetes/issues/10782), and in early 2017 we got a proposal and now initial work on the resulting [Vertical Pod Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler) (VPA).

With `resorcerer` we want to contribute to the advancement of VPAs. It's a simple prototypical implementation that should allow users to learn more about their resource consumption footprint. It is an opinionated implementation, making a number of assumptions:

1. Prometheus is available in the cluster
1. You can run it in privileged mode (as access to all necessary metrics)
1. As much as possible should happen automatically (that's where the magic/sorcerer comes into play ;)

## Setup

We don't do binary releases, you need Go 1.8 and [dep](https://github.com/golang/dep) to build it for your platform. If you don't have `dep` installed yet, do `go get -u github.com/golang/dep/cmd/dep` now and then:

```
$ dep ensure
```

Now you can use the Makefile to build the binaries and containers as you see fit.

## Usage

 
