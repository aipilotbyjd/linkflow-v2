# Kubernetes Deployment

Deploy LinkFlow on Kubernetes for high availability and scalability.

## Prerequisites

- Kubernetes cluster (1.25+)
- kubectl configured
- Helm 3.x (optional)
- Ingress controller (nginx-ingress recommended)

## Quick Start

```bash
# Create namespace
kubectl create namespace linkflow

# Apply configurations
kubectl apply -f k8s/ -n linkflow
```

## Kubernetes Manifests

### Namespace and ConfigMap

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: linkflow
---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: linkflow-config
  namespace: linkflow
data:
  APP_ENVIRONMENT: "production"
  APP_DEBUG: "false"
  DATABASE_HOST: "postgres-service"
  DATABASE_PORT: "5432"
  DATABASE_NAME: "linkflow"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
```

### Secrets

```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: linkflow-secrets
  namespace: linkflow
type: Opaque
stringData:
  DATABASE_USER: linkflow
  DATABASE_PASSWORD: <base64-encoded>
  REDIS_PASSWORD: <base64-encoded>
  JWT_SECRET: <base64-encoded>
```

### API Deployment

```yaml
# k8s/api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: linkflow-api
  namespace: linkflow
spec:
  replicas: 3
  selector:
    matchLabels:
      app: linkflow-api
  template:
    metadata:
      labels:
        app: linkflow-api
    spec:
      containers:
      - name: api
        image: linkflow/api:latest
        ports:
        - containerPort: 8090
        envFrom:
        - configMapRef:
            name: linkflow-config
        - secretRef:
            name: linkflow-secrets
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "1000m"
            memory: "1Gi"
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8090
          initialDelaySeconds: 5
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8090
          initialDelaySeconds: 15
          periodSeconds: 20
---
apiVersion: v1
kind: Service
metadata:
  name: linkflow-api-service
  namespace: linkflow
spec:
  selector:
    app: linkflow-api
  ports:
  - port: 80
    targetPort: 8090
  type: ClusterIP
```

### Worker Deployment

```yaml
# k8s/worker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: linkflow-worker
  namespace: linkflow
spec:
  replicas: 5
  selector:
    matchLabels:
      app: linkflow-worker
  template:
    metadata:
      labels:
        app: linkflow-worker
    spec:
      containers:
      - name: worker
        image: linkflow/worker:latest
        envFrom:
        - configMapRef:
            name: linkflow-config
        - secretRef:
            name: linkflow-secrets
        env:
        - name: WORKER_CONCURRENCY
          value: "10"
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "2000m"
            memory: "1Gi"
        livenessProbe:
          exec:
            command: ["/bin/sh", "-c", "pgrep worker"]
          initialDelaySeconds: 10
          periodSeconds: 30
```

### Scheduler Deployment

```yaml
# k8s/scheduler-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: linkflow-scheduler
  namespace: linkflow
spec:
  replicas: 2
  selector:
    matchLabels:
      app: linkflow-scheduler
  template:
    metadata:
      labels:
        app: linkflow-scheduler
    spec:
      containers:
      - name: scheduler
        image: linkflow/scheduler:latest
        envFrom:
        - configMapRef:
            name: linkflow-config
        - secretRef:
            name: linkflow-secrets
        resources:
          requests:
            cpu: "100m"
            memory: "256Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
```

### Horizontal Pod Autoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: linkflow-api-hpa
  namespace: linkflow
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: linkflow-api
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: linkflow-worker-hpa
  namespace: linkflow
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: linkflow-worker
  minReplicas: 2
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: linkflow-ingress
  namespace: linkflow
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
spec:
  tls:
  - hosts:
    - api.linkflow.ai
    secretName: linkflow-tls
  rules:
  - host: api.linkflow.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: linkflow-api-service
            port:
              number: 80
```

### PostgreSQL (using Helm)

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami

helm install postgres bitnami/postgresql \
  --namespace linkflow \
  --set auth.postgresPassword=<password> \
  --set auth.database=linkflow \
  --set primary.persistence.size=50Gi
```

### Redis (using Helm)

```bash
helm install redis bitnami/redis \
  --namespace linkflow \
  --set auth.password=<password> \
  --set master.persistence.size=10Gi
```

## KEDA Autoscaling (Queue-based)

```yaml
# k8s/keda-scaledobject.yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: linkflow-worker-scaler
  namespace: linkflow
spec:
  scaleTargetRef:
    name: linkflow-worker
  minReplicaCount: 2
  maxReplicaCount: 50
  triggers:
  - type: redis
    metadata:
      address: redis-service:6379
      passwordFromEnv: REDIS_PASSWORD
      listName: asynq:{default}
      listLength: "100"
```

## Deployment Commands

```bash
# Apply all manifests
kubectl apply -f k8s/ -n linkflow

# Check deployment status
kubectl get pods -n linkflow
kubectl get deployments -n linkflow

# View logs
kubectl logs -f deployment/linkflow-api -n linkflow

# Scale manually
kubectl scale deployment linkflow-worker --replicas=10 -n linkflow

# Rolling update
kubectl set image deployment/linkflow-api api=linkflow/api:v1.1.0 -n linkflow

# Rollback
kubectl rollout undo deployment/linkflow-api -n linkflow

# Check rollout status
kubectl rollout status deployment/linkflow-api -n linkflow
```

## Monitoring

### Prometheus ServiceMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: linkflow-api
  namespace: linkflow
spec:
  selector:
    matchLabels:
      app: linkflow-api
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

## Resource Quotas

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: linkflow-quota
  namespace: linkflow
spec:
  hard:
    requests.cpu: "20"
    requests.memory: 40Gi
    limits.cpu: "40"
    limits.memory: 80Gi
    pods: "100"
```

## Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: linkflow-network-policy
  namespace: linkflow
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: linkflow
  - to:
    - ipBlock:
        cidr: 0.0.0.0/0
```

## Troubleshooting

```bash
# Check pod status
kubectl describe pod <pod-name> -n linkflow

# Check events
kubectl get events -n linkflow --sort-by='.lastTimestamp'

# Execute into pod
kubectl exec -it <pod-name> -n linkflow -- /bin/sh

# Port forward for debugging
kubectl port-forward svc/linkflow-api-service 8090:80 -n linkflow
```
