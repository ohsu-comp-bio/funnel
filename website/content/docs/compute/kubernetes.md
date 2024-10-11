---
title: Kubernetes
menu:
  main:
    parent: Compute
    weight: 20
---

# Kubernetes

This guide will take you through the process of setting up Funnel as a kubernetes service.

#### Create a Service:

*funnel-service.yml*

```yaml
{{< read-file "static/funnel-config-examples/kubernetes/funnel-service.yml" >}}
```

Deploy it:

```sh
kubectl apply -f funnel-service.yml
```

Get the clusterIP:

```sh
kubectl get services funnel --output=yaml | grep "clusterIP:"
```

Use this value to configure the server hostname of the worker config.

#### Create Funnel config files

*funnel-server-config.yml*

```yaml
{{< read-file "static/funnel-config-examples/kubernetes/funnel-server-config.yml" >}}
```

We recommend setting `DisableJobCleanup` to `true` for debugging - otherwise failed jobs will be cleanup up.

*funnel-worker-config.yml*

***Remember to modify the template below to have the actual server hostname.***

```yaml
{{< read-file "static/funnel-config-examples/kubernetes/funnel-worker-config.yml" >}}
```

#### Create a ConfigMap

```sh
kubectl create configmap funnel-config --from-file=funnel-server-config.yml --from-file=funnel-worker-config.yml
```

#### Create a Service Account for Funnel

Define a Role and RoleBinding:

*role.yml*

```yaml
{{< read-file "static/funnel-config-examples/kubernetes/role.yml" >}}
```

*role_binding.yml*

```yaml
{{< read-file "static/funnel-config-examples/kubernetes/role_binding.yml" >}}
```

Create the service account, role and role binding:

```sh
kubectl create serviceaccount funnel-sa --namespace default
kubectl create -f role.yml
kubectl create -f role_binding.yml
```

#### Create a Deployment

*funnel-deployment.yml*

```yaml
{{< read-file "static/funnel-config-examples/kubernetes/funnel-deployment.yml" >}}
```

Deploy it:

```sh
kubectl apply -f funnel-deployment.yml
```

#### Proxy the Service for local testing

```sh
kubectl port-forward service/funnel 8000:8000
```

Now you can access the funnel server locally. Verify by running:

```sh
funnel task list
```

Now try running a task:

```sh
funnel examples hello-world > hello.json
funnel task create hello.json
```
