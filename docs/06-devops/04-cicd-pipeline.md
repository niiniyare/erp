---
title: CI/CD Pipeline
portal: 6 — DevOps
section: 06-devops
audience: [devops, backend-engineer]
related:
  - "[DevOps Overview](01-devops-overview.md)"
  - "[Kubernetes Deployment](03-kubernetes-deployment.md)"
  - "[Environment Variables](02-environment-variables.md)"
---

# CI/CD Pipeline

## Pipeline Overview

```
push to branch
  └── CI: lint + test + build
        └── merge to main
              └── CD: build image → push → deploy staging
                    └── manual promote
                          └── deploy production
```

## GitHub Actions: CI Workflow

`.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: ["*"]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: awoerp_test
        ports: ["5432:5432"]
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7
        ports: ["6379:6379"]

    steps:
    - uses: actions/checkout@v4

    - uses: actions/setup-go@v5
      with:
        go-version-file: go.mod
        cache: true

    - name: Install dependencies
      run: go mod download

    - name: Run migrations
      run: |
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        migrate -database "postgres://postgres:postgres@localhost:5432/awoerp_test?sslmode=disable" \
                -path db/migration up
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/awoerp_test?sslmode=disable

    - name: Generate SQLC
      run: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && sqlc generate

    - name: Lint
      uses: golangci/golangci-lint-action@v4
      with:
        version: latest

    - name: Test
      run: go test ./... -race -count=1 -timeout=120s
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/awoerp_test?sslmode=disable
        REDIS_URL: redis://localhost:6379/0
        SESSION_SECRET: ci-test-secret-not-for-production
        ENV: test

    - name: Security scan
      run: |
        go install golang.org/x/vuln/cmd/govulncheck@latest
        govulncheck ./...
```

## GitHub Actions: CD Workflow

`.github/workflows/cd.yml`:

```yaml
name: CD

on:
  push:
    branches: [main]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
      id-token: write

    steps:
    - uses: actions/checkout@v4

    - name: Set image tag
      id: meta
      run: echo "tag=$(git rev-parse --short HEAD)" >> $GITHUB_OUTPUT

    - name: Log in to GitHub Container Registry
      uses: docker/login-action@v3
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Build and push image
      uses: docker/build-push-action@v5
      with:
        context: .
        push: true
        tags: |
          ghcr.io/${{ github.repository }}:${{ steps.meta.outputs.tag }}
          ghcr.io/${{ github.repository }}:latest

    - name: Deploy to staging
      run: |
        kubectl set image deployment/awoerp-server \
          awoerp-server=ghcr.io/${{ github.repository }}:${{ steps.meta.outputs.tag }} \
          -n staging
        kubectl rollout status deployment/awoerp-server -n staging
      env:
        KUBECONFIG: ${{ secrets.STAGING_KUBECONFIG }}

    - name: Run smoke tests against staging
      run: |
        curl -sf https://staging.awoerp.com/api/v1/health
```

## Production Deploy (Manual Promotion)

Production deploys are manual to ensure human review:

```bash
# 1. Confirm staging is healthy
curl -sf https://staging.awoerp.com/api/v1/health

# 2. Run migration job
kubectl apply -f k8s/jobs/migrate-{version}.yaml -n production
kubectl wait --for=condition=complete job/awoerp-migrate-{version} -n production --timeout=300s

# 3. Deploy new image
kubectl set image deployment/awoerp-server \
  awoerp-server=ghcr.io/org/awoerp:{tag} \
  -n production

# 4. Monitor rollout
kubectl rollout status deployment/awoerp-server -n production

# 5. Smoke test
curl -sf https://demo.awoerp.com/api/v1/health
```

## Dockerfile

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o awoerp ./cmd/server

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/awoerp /app/awoerp
COPY --from=builder /app/db/migration /app/db/migration

USER nonroot
EXPOSE 8080 8081 9090

ENTRYPOINT ["/app/awoerp"]
```

Distroless base — no shell, no package manager, minimal attack surface. Binary is statically compiled.

## Branch Strategy

| Branch | Purpose | Auto-deploy |
|--------|---------|-------------|
| `main` | Production-ready code | Staging auto |
| `dev` | Active development | Dev env (if configured) |
| `feature/*` | Feature branches | None |
| `fix/*` | Bug fix branches | None |

PRs must pass CI before merge to `main`. No direct pushes to `main`.

## Versioning

Images are tagged with the short Git SHA (`a1b2c3d`). No semver for internal deployments. Semver used only for public API version (`/api/v1`).

Rollback = deploy the previous image tag:

```bash
kubectl set image deployment/awoerp-server \
  awoerp-server=ghcr.io/org/awoerp:{previous-sha} \
  -n production
```
