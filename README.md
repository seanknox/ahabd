## Introduction

ahabd (Docker Health Daemon) is a Kubernetes daemonset that performs docker
 restarts when the need to do so is indicated by basic docker health checks.

* Pulls alpine:latest
* Creates a container from alpine:latest that echo's "I'm alive"
* Starts the container
* Removes the container

## Installation
``` sh
kubectl apply -f ahabd-ds.yaml
```
