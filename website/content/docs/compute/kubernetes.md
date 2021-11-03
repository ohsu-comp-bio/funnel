---
title: Kubernetes
menu:
  main:
    parent: Compute
    weight: 20
---

# Kubernetes

This guide will take you through the process of setting up Funnel as a kubernetes service.

Kuberenetes Resources: 
- [Service](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/)
- [Roles and RoleBindings](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#default-roles-and-role-bindings)
- [Job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/)

Additional Funnel deployment resources can be found here: https://github.com/ohsu-comp-bio/funnel/tree/master/deployments/kubernetes

#### Create a Service:

*funnel-service.yml*

```
apiVersion: v1
kind: Service
metadata:
  name: funnel
spec:
  selector:
    app: funnel
  ports:
    - name: http
      protocol: TCP
      port: 8000
      targetPort: 8000
    - name: rpc
      protocol: TCP
      port: 9090
      targetPort: 9090

```

Deploy it:

```
kubectl apply -f funnel-service.yml
```

Get the clusterIP:

```
kubectl get services funnel --output=yaml | grep clusterIP
```

Use this value to configure the server hostname of the worker config. 

#### Create Funnel config files

*Note*: The configures job template uses the image, `ohsucompbio/funnel-dind:latest`, which is built on docker's official [docker-in-docker image (dind)](https://hub.docker.com/_/docker). You can also use the experimental [rootless dind variant](https://docs.docker.com/engine/security/rootless/) by changing the image to `ohsucompbio/funnel-dind-rootless:latest`.

*funnel-server-config.yml*

```
Database: boltdb

Compute: kubernetes

Logger:
  Level: debug

Kubernetes:
  DisableJobCleanup: false
  DisableReconciler: false
  ReconcileRate: 5m
  Namespace: default
  Template: | 
    apiVersion: batch/v1
    kind: Job
    metadata:
      ## DO NOT CHANGE NAME
      name: {{.TaskId}}
      namespace: {{.Namespace}}
    spec: 
      backoffLimit: 0
      completions: 1
      template:
        spec:
          restartPolicy: Never
          containers: 
            - name: {{printf "funnel-worker-%s" .TaskId}}
              image: ohsucompbio/funnel-dind:latest
              imagePullPolicy: IfNotPresent
              args:
                - "funnel"
                - "worker"
                - "run"
                - "--config"
                - "/etc/config/funnel-worker-config.yml"
                - "--taskID"
                - {{.TaskId}}
              resources:
                  requests:
                    cpu: {{if ne .Cpus 0 -}}{{.Cpus}}{{ else }}{{"100m"}}{{end}}
                    memory: {{if ne .RamGb 0.0 -}}{{printf "%.0fG" .RamGb}}{{else}}{{"16M"}}{{end}}
                    ephemeral-storage: {{if ne .DiskGb 0.0 -}}{{printf "%.0fG" .DiskGb}}{{else}}{{"100M"}}{{end}}
              volumeMounts:
                - name: {{printf "funnel-storage-%s" .TaskId}}
                  mountPath: {{printf "/opt/funnel/funnel-work-dir/%s" .TaskId}}
                - name: config-volume
                  mountPath: /etc/config

              securityContext:
                privileged: true
    
          volumes: 
            - name: {{printf "funnel-storage-%s" .TaskId}}
              emptyDir: {}
            - name: config-volume
              configMap:
                name: funnel-config
```

I recommend setting `DisableJobCleanup` to `true` for debugging - otherwise failed jobs will be cleanup up. 

*funnel-worker-config.yml*

***Remember to modify the template below to have the actual server hostname.***

```
Database: boltdb

BoltDB:
  Path: /opt/funnel/funnel-work-dir/funnel.bolt.db

Compute: kubernetes

Logger:
  Level: debug

RPCClient:
  MaxRetries: 3
  Timeout: 30s

EventWriters:
  - rpc
  - log

Server:
  HostName: < funnel service clusterIP >
  RPCPort: 9090
```

#### Create a ConfigMap

```
kubectl create configmap funnel-config --from-file=funnel-server-config.yml --from-file=funnel-worker-config.yml
```

#### Create a Service Account for Funnel

Define a Role and RoleBinding:

*role.yml*

```
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: default
  name: funnel-role
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["batch", "extensions"]
  resources: ["jobs"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["extensions", "apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

*role_binding.yml*

```
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: funnel-rolebinding
  namespace: default
subjects:
- kind: ServiceAccount
  name: funnel-sa
roleRef:
  kind: Role
  name: funnel-role
  apiGroup: rbac.authorization.k8s.io
```

Create the service account, role and role binding:

```
kubectl create serviceaccount funnel-sa --namespace default
kubectl create -f role.yml
kubectl create -f role_binding.yml
```

#### Create a Deployment

*funnel-deployment.yml*

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: funnel
  labels:
    app: funnel
spec:
  replicas: 1
  selector:
    matchLabels:
      app: funnel
  template:
    metadata:
      labels:
        app: funnel
    spec:
      serviceAccountName: funnel-sa
      containers:
        - name: funnel
          image: ohsucompbio/funnel:latest
          imagePullPolicy: IfNotPresent
          command: 
            - 'funnel'
            - 'server'
            - 'run'
            - '--config'
            - '/etc/config/funnel-server-config.yml'
          resources: 
            requests: 
              cpu: 2 
              memory: 4G
              ephemeral-storage: 25G # needed since we are using boltdb
          volumeMounts:
            - name: funnel-deployment-storage
              mountPath: /opt/funnel/funnel-work-dir
            - name: config-volume
              mountPath: /etc/config
          ports:
            - containerPort: 8000
            - containerPort: 9090

      volumes:
        - name: funnel-deployment-storage
          emptyDir: {}
        - name: config-volume
          configMap:
            name: funnel-config
```

Deploy it:

```
kubectl apply -f funnel-deployment.yml
```

#### Proxy the Service for local testing

```
kubectl port-forward service/funnel 8000:8000
```

Now you can access the funnel server locally. Verify by running:

```
funnel task list
```

Now try running a task:

```
funnel examples hello-world > hello.json
funnel task create hello.json
```
