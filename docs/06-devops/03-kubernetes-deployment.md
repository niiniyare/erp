---
title: Kubernetes Deployment Guide
portal: 6 — DevOps
section: 06-devops
audience: [devops, sre]
related:
  - "[DevOps Overview](01-devops-overview.md)"
  - "[Environment Variables](02-environment-variables.md)"
  - "[Health Checks](../04-backend-engineering/00-module-development-guide/21-server-startup/04-health-checks.md)"
---

# Kubernetes Deployment Guide

## Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: awoerp-server
  namespace: production
  labels:
    app: awoerp-server
    version: "1.2.0"
spec:
  replicas: 3
  selector:
    matchLabels:
      app: awoerp-server
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0      # never take pods offline during rollout
      maxSurge: 1            # one extra pod during rollout
  template:
    metadata:
      labels:
        app: awoerp-server
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: awoerp-server
      terminationGracePeriodSeconds: 60
      containers:
      - name: awoerp-server
        image: ghcr.io/org/awoerp:1.2.0
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 8081
          name: health
        - containerPort: 9090
          name: metrics
        envFrom:
        - secretRef:
            name: awoerp-secrets
        - configMapRef:
            name: awoerp-config
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8081
          initialDelaySeconds: 10
          periodSeconds: 10
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
          failureThreshold: 3
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 5"]
```

The `preStop` sleep ensures the pod is removed from the load balancer before the server begins shutting down — avoids in-flight request drops during rolling updates.

## Migration Job

Run before rolling out new app pods:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: awoerp-migrate-v1-2-0
  namespace: production
spec:
  backoffLimit: 3
  template:
    spec:
      restartPolicy: OnFailure
      containers:
      - name: migrate
        image: ghcr.io/org/awoerp:1.2.0
        command: ["/app/awoerp", "migrate", "up"]
        envFrom:
        - secretRef:
            name: awoerp-secrets
      initContainers:
      - name: wait-for-db
        image: postgres:15
        command:
        - sh
        - -c
        - |
          until pg_isready -h $DB_HOST -p $DB_PORT; do
            echo "Waiting for database..."
            sleep 2
          done
```

## Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: awoerp-server
  namespace: production
spec:
  selector:
    app: awoerp-server
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: health
    port: 8081
    targetPort: 8081
  - name: metrics
    port: 9090
    targetPort: 9090
```

## Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: awoerp-ingress
  namespace: production
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "20m"   # allow CSV import
    nginx.ingress.kubernetes.io/proxy-read-timeout: "120"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - "*.awoerp.com"
    secretName: awoerp-tls
  rules:
  - host: "*.awoerp.com"
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: awoerp-server
            port:
              number: 80
```

## HorizontalPodAutoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: awoerp-server
  namespace: production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: awoerp-server
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## PodDisruptionBudget

Ensures at least 2 pods are always running during node maintenance:

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: awoerp-server
  namespace: production
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: awoerp-server
```

## ConfigMap

Non-secret configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: awoerp-config
  namespace: production
data:
  HTTP_PORT: "8080"
  HEALTH_PORT: "8081"
  METRICS_PORT: "9090"
  LOG_LEVEL: "info"
  LOG_FORMAT: "json"
  ENV: "production"
  DATABASE_MAX_CONNS: "20"
  TEMPORAL_NAMESPACE: "awoerp-production"
  OTEL_SERVICE_NAME: "awoerp"
```

## Deploy Checklist

1. Run migration job and verify it completed successfully
2. Update image tag in Deployment spec
3. Apply manifest: `kubectl apply -f k8s/deployment.yaml -n production`
4. Monitor rollout: `kubectl rollout status deployment/awoerp-server -n production`
5. Check readiness: `kubectl get pods -n production -l app=awoerp-server`
6. Smoke test: `curl https://{tenant}.awoerp.com/api/v1/health`
7. Check error rate in Grafana for 10 minutes post-deploy
