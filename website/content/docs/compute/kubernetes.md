---
title: Kubernetes
menu:
  main:
    parent: Compute
    weight: 20
---

> Funnel on Kubernetes is in active development and may involve frequent updates 🚧

# Quick Start

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

> *[role.yml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/role.yml)*

> *[role_binding.yml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/role_binding.yml)*

```sh
kubectl create serviceaccount funnel-sa --namespace default
kubectl apply -f role.yml
kubectl apply -f role_binding.yml
```

### 5. Create a Persistent Volume Claim

> *[funnel-storage-pvc.yml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/funnel-storage-pvc.yml)*

```sh
kubectl apply -f funnel-storage-pvc.yml
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

Now the funnel server can be accessed as if it were running locally. This can be verified by listing all tasks, which will return an empty JSON list:

```sh
funnel task list
# {}
```

A task can then be submitted following the [standard workflow](../../tasks):

```sh
funnel examples hello-world > hello-world.json

funnel task create hello-world.json
# <Task ID>
```

# Storage Architecture

<a href="https://www.figma.com/board/bzgv8kVL2QKESU3Sqn7S1a/Funnel-%2B-Gen3?node-id=2-1059&t=9bcuG0bMAcxBLcRD-1">
  <img title="K8s Storage" src="/img/k8s-pvc.png" />
</a>

# Additional Resources 📚

- [Helm Repo](https://ohsu-comp-bio.github.io/helm-charts)

- [Helm Repo Source](https://github.com/ohsu-comp-bio/helm-charts)

- [Helm Charts](https://github.com/ohsu-comp-bio/funnel/tree/develop/deployments/kubernetes/helm)

- [The Chart Best Practices Guide](https://helm.sh/docs/chart_best_practices/)
