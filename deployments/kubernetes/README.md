# Using Funnel with Kubernetes

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

*funnel-server-config.yml*

```
Database: boltdb

Compute: kubernetes

Logger:
  Level: debug

Kubernetes:
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
              image: ohsucompbio/funnel-kubernetes-worker:latest
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

I recommend setting `DisableReconciler` to `true` for debugging - otherwise failed jobs will be cleanup up. 

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
      containers:
        - name: funnel
          image: ohsucompbio/funnel:latest
          imagePullPolicy: IfNotPresent
          command: 
            - '/opt/funnel/funnel'
            - 'server'
            - 'run'
            - '--config'
            - '/etc/config/funnel-server-config.yml'
          resources: 
            requests: 
              cpu: 2 
              memory: 4Gi
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
