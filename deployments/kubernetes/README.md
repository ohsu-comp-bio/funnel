> [!WARNING]
> Funnel's Kubernetes support is in active development and may involve frequent updates 🚧

# Overview

This guide will take you through the process of setting up Funnel as a kubernetes service.

Kuberenetes Resources:
- [Service](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/)
- [Roles and RoleBindings](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#default-roles-and-role-bindings)
- [Job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/)

# Deployment Steps

## 1. Create a Service:

> *[funnel-service.yml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/funnel-service.yml)*

```sh
kubectl apply -f funnel-service.yml
```

Get the clusterIP:

```sh
kubectl get services funnel --output=yaml | grep "clusterIP:"
```

Use this value to configure the server hostname of the worker config.

## 2. Create Funnel config files

> *[funnel-server-config.yml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/funnel-server-config.yml)*

We recommend setting `DisableJobCleanup` to `true` for debugging - otherwise failed jobs will be cleanup up.

> *[funnel-worker-config.yml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/funnel-worker-config.yml)*

***Remember to modify the file to have the actual server hostname.***

## 3. Create a ConfigMap

```sh
kubectl create configmap funnel-config --from-file=funnel-server-config.yml --from-file=funnel-worker-config.yml
```

## 4. Create a Service Account for Funnel

Define a Role and RoleBinding:

> *[role.yml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/role.yml)*

> *[role_binding.yml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/role_binding.yml)*

```sh
kubectl create serviceaccount funnel-sa --namespace default
kubectl apply -f role.yml
kubectl apply -f role_binding.yml
```

## 5. Create a Persistent Volume Claim

Define a PVC for storage:

> *[funnel-storage-pvc.yml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/funnel-storage-pvc.yml)*

```sh
kubectl apply -f funnel-storage-pvc.yml
```

## 6. Create a Deployment

> *[funnel-deployment.yml](https://github.com/ohsu-comp-bio/funnel/blob/develop/deployments/kubernetes/funnel-deployment.yml)*

```sh
kubectl apply -f funnel-deployment.yml
```

## 7. Proxy the Service for local testing

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
