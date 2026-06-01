---
title: DevOps Overview
portal: 6 — DevOps
section: 06-devops
audience: [devops, sre, tech-lead]
related:
  - "[Environment Variables](02-environment-variables.md)"
  - "[Kubernetes Deployment](03-kubernetes-deployment.md)"
  - "[CI/CD Pipeline](04-cicd-pipeline.md)"
  - "[Local Development](05-local-development.md)"
  - "[Monitoring and Alerting](06-monitoring.md)"
---

# DevOps Overview

## Infrastructure

| Component | Technology | Managed by |
|-----------|-----------|-----------|
| Container orchestration | Kubernetes | Cloud provider (GKE/EKS/AKS) |
| Database | PostgreSQL 15 | Managed RDS / Cloud SQL |
| Cache | Redis 7 | Managed (ElastiCache / Memorystore) |
| Workflows | Temporal | Self-hosted or Temporal Cloud |
| Object storage | S3-compatible | Cloud provider |
| Container registry | Docker Hub / GCR | Cloud provider |
| CI/CD | GitHub Actions | GitHub |

## Build

```dockerfile
# Multi-stage Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/server /server
COPY --from=builder /app/db/migration /db/migration
EXPOSE 8080 8081 9090
ENTRYPOINT ["/server"]
```

## Kubernetes Resources

```yaml
# Deployment (HTTP server)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: awoerp-server
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: awoerp
        image: awoerp:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8081
          initialDelaySeconds: 10
          periodSeconds: 5
        envFrom:
        - secretRef:
            name: awoerp-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## CI/CD Pipeline

```yaml
# .github/workflows/ci.yml (simplified)
jobs:
  test:
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: make lint
      - run: make test
      - run: make test-coverage

  build:
    needs: test
    steps:
      - run: docker build -t awoerp:${{ github.sha }} .
      - run: docker push awoerp:${{ github.sha }}

  deploy-staging:
    needs: build
    if: github.ref == 'refs/heads/main'
    steps:
      - run: kubectl set image deployment/awoerp-server awoerp=awoerp:${{ github.sha }}
```

## Migrations in CI/CD

Migrations run as a Kubernetes Job before the deployment rollout:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migrate-${{ github.sha }}
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: awoerp:${{ github.sha }}
        command: ["/server", "migrate"]
        envFrom:
        - secretRef:
            name: awoerp-secrets
      restartPolicy: Never
```

The server binary accepts a `migrate` subcommand that runs migrations and exits.
