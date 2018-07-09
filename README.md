## Introduction

ahabd (Docker Health Daemon) is a Kubernetes daemonset that performs docker
 restarts when the need to do so is indicated by basic docker health checks.

* Pulls alpine:latest
* Creates a container with `alpine:latest` that will `echo "hello world"`
* Starts the container
* Checks the log output of the container
* Removes the container

## Installation
To obtain a default installation with a period of 1h:
``` sh
kubectl apply -f ahabd-ds.yaml
```
