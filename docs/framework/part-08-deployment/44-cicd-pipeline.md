---
title: "Chapter 44: CI/CD Pipeline"
part: "Part VIII — Deployment and Operations"
chapter: 44
section: "44-cicd-pipeline"
related:
  - "[Chapter 43: Docker Containerisation](43-docker-containerisation.md)"
  - "[Chapter 11: Database Migrations](../part-02-entity-system/11-database-migrations.md)"
---

# Chapter 44: CI/CD Pipeline

Awo uses GitHub Actions for CI/CD. The pipeline runs tests, lints migrations, builds Docker images, and deploys to staging and production using a GitOps model.

---

## 44.1. Pipeline Overview

```
Push to branch
    │
    ▼
┌─────────────┐
│   ci.yml    │ — runs on every push
│  (test +    │
│   build)    │
└──────┬──────┘
       │ (merge to main)
       ▼
┌─────────────┐
│ deploy-     │ — auto-deploy to staging
│ staging.yml │
└──────┬──────┘
       │ (manual approval)
       ▼
┌─────────────┐
│  deploy-    │ — deploy to production
│  prod.yml   │
└─────────────┘
```

---

## 44.2. CI Workflow

### 44.2.1. `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: ["**"]
  pull_request:
    branches: [main]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: awo_test
          POSTGRES_PASSWORD: awo_test
          POSTGRES_DB: awo_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
        ports:
          - 6379:6379

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version-file: go.mod
        cache: true

    - name: Run linter
      uses: golangci/golangci-lint-action@v6
      with:
        version: v1.59

    - name: Run tests
      env:
        AWO_DB_URL: postgres://awo_test:awo_test@localhost:5432/awo_test?sslmode=disable
        AWO_REDIS_URL: redis://localhost:6379
        AWO_ENV: test
        AWO_SESSION_SECRET: test-session-secret-for-ci-only
      run: go test ./... -race -coverprofile=coverage.out -timeout=120s

    - name: Upload coverage
      uses: codecov/codecov-action@v4
      with:
        file: coverage.out
        fail_ci_if_error: false

  migration-lint:
    name: Migration Lint
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Install Atlas
      run: |
        curl -sSf https://atlasgo.sh | sh

    - name: Lint migrations
      run: |
        atlas migrate lint \
          --dev-url "docker://postgres/16/dev" \
          --dir "file://db/migration" \
          --format "{{ range .Files }}{{ .Name }}: {{ range .Reports }}{{ .Text }}{{ end }}{{ end }}"

  build:
    name: Build Docker Image
    runs-on: ubuntu-latest
    needs: [test, migration-lint]
    if: github.ref == 'refs/heads/main' || startsWith(github.ref, 'refs/tags/')

    steps:
    - uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Build and push
      uses: docker/build-push-action@v5
      with:
        context: .
        push: true
        tags: |
          ghcr.io/awo-so/awo:${{ github.sha }}
          ghcr.io/awo-so/awo:latest
        build-args: |
          VERSION=${{ github.ref_name }}
          BUILD_TIME=${{ github.event.head_commit.timestamp }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

    - name: Scan image
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: ghcr.io/awo-so/awo:${{ github.sha }}
        severity: HIGH,CRITICAL
        exit-code: 1

    - name: Sign image
      run: |
        cosign sign --key env://COSIGN_PRIVATE_KEY \
          ghcr.io/awo-so/awo:${{ github.sha }}
      env:
        COSIGN_PRIVATE_KEY: ${{ secrets.COSIGN_PRIVATE_KEY }}
```

---

## 44.3. Staging Deployment

### 44.3.1. `.github/workflows/deploy-staging.yml`

```yaml
name: Deploy to Staging

on:
  workflow_run:
    workflows: [CI]
    types: [completed]
    branches: [main]

jobs:
  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    if: ${{ github.event.workflow_run.conclusion == 'success' }}
    environment: staging

    steps:
    - uses: actions/checkout@v4

    - name: Set up kubectl
      uses: azure/setup-kubectl@v4

    - name: Configure kubeconfig
      run: |
        echo "${{ secrets.STAGING_KUBECONFIG }}" | base64 -d > ~/.kube/config

    - name: Run migrations on staging
      run: |
        kubectl run migrate-${{ github.sha }} \
          --image=ghcr.io/awo-so/awo:${{ github.sha }} \
          --restart=Never \
          --env="AWO_DB_URL=${{ secrets.STAGING_DB_URL }}" \
          -- /awo-server migrate --up
        kubectl wait pod/migrate-${{ github.sha }} \
          --for=condition=Succeeded --timeout=5m
        kubectl delete pod migrate-${{ github.sha }}

    - name: Update image in staging
      run: |
        kubectl set image deployment/awo-api \
          awo=ghcr.io/awo-so/awo:${{ github.sha }} \
          -n awo-staging

    - name: Wait for rollout
      run: |
        kubectl rollout status deployment/awo-api -n awo-staging --timeout=10m

    - name: Run smoke tests
      run: |
        curl -f https://staging.awo.so/health/ready || exit 1
        curl -f https://staging.awo.so/health/deep || exit 1
```

---

## 44.4. Production Deployment

### 44.4.1. Manual Approval Gate

Production deployments require manual approval via GitHub Environments:

```yaml
name: Deploy to Production

on:
  workflow_dispatch:
    inputs:
      image_tag:
        description: Image tag to deploy (e.g. sha-abc1234)
        required: true

jobs:
  deploy-prod:
    name: Deploy to Production
    runs-on: ubuntu-latest
    environment: production  # has required reviewers configured in GitHub
```

The `production` environment in GitHub requires approval from 2 designated reviewers before the job runs. This enforces a second set of eyes on every production deployment.

### 44.4.2. Zero-Downtime Deployment Strategy

Kubernetes rolling update with maxSurge/maxUnavailable:

```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1        # one extra pod during update
    maxUnavailable: 0  # never take a pod down until replacement is ready
```

`maxUnavailable: 0` ensures at least `minReplicas` (3) pods are always running. The update sequence is:
1. Start 4th pod (surge) with new image
2. Wait for 4th pod to pass readiness probe
3. Terminate one old pod
4. Repeat until all pods are updated

### 44.4.3. Rollback

If the deployment causes increased error rates:

```bash
# Immediate rollback to previous revision
kubectl rollout undo deployment/awo-api -n awo-prod

# Verify rollback
kubectl rollout status deployment/awo-api -n awo-prod
```

The previous image is still in the container registry, so rollback is instant (no re-pull needed if the image is already cached on nodes).

---

## 44.5. Database Migration Strategy in CI/CD

### 44.5.1. Migration Gate

Migrations run before the new application code starts (init container pattern). This means:
- The database is always at least as new as the application code
- New code can depend on new columns being present
- Old code is still running on pods that haven't yet been updated

This requires the expand-contract pattern (Chapter 11): new columns must have defaults and the old code must tolerate their presence.

### 44.5.2. Migration Dry Run in CI

The migration lint job in CI checks for:
- Missing up/down migration pairs
- Destructive operations (DROP TABLE, DROP COLUMN) without a corresponding down migration
- Missing index on new foreign keys
- Lock-heavy operations that should use CONCURRENTLY

```bash
atlas migrate lint \
  --dev-url "docker://postgres/16/dev" \
  --dir "file://db/migration" \
  --git-base origin/main
```

`--git-base origin/main` lints only the migration files added since the last merge to main — not the entire migration history on every PR.
