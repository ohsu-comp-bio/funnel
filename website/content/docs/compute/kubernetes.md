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

# Deploying

## 1. Deploying with Helm ⚡️

```sh
helm repo add ohsu https://ohsu-comp-bio.github.io/helm-charts
helm repo update
helm upgrade --install ohsu funnel
```

{{< details title="(Alternative) Deploying with `kubectl` ⚙️" >}}

### 1. Create a Service:

Deploy it:

```sh
kubectl apply -f funnel-service.yml
```

### 2. Create Funnel config files

> *[funnel-server.yaml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/funnel-server.yaml)*

> *[funnel-worker.yaml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/funnel-worker.yaml)*

Get the clusterIP:

```sh
export HOSTNAME=$(kubectl get services funnel --output=jsonpath='{.spec.clusterIP}')

sed -i "s|\${HOSTNAME}|${HOSTNAME}|g" funnel-worker.yaml
```

### 3. Create a ConfigMap

```sh
kubectl create configmap funnel-config --from-file=funnel-server.yaml --from-file=funnel-worker.yaml
```

### 4. Create a Service Account for Funnel

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

### 5. Create a Persistent Volume Claim

*funnel-deployment.yml*

```yaml
{{< read-file "static/funnel-config-examples/kubernetes/funnel-deployment.yml" >}}
```

### 6. Create a Deployment

> *[funnel-deployment.yml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/funnel-deployment.yml)*

```sh
kubectl apply -f funnel-deployment.yml
```

{{< /details >}}

# 2. Proxy the Service for local testing

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
