---
title: "Chapter 43: Docker Containerisation"
part: "Part VIII — Deployment and Operations"
chapter: 43
section: "43-docker-containerisation"
related:
  - "[Chapter 42: Environment Architecture](42-environment-architecture.md)"
  - "[Chapter 44: CI/CD Pipeline](44-cicd-pipeline.md)"
---

# Chapter 43: Docker Containerisation

Awo ships as a single Docker image containing the Go binary and the compiled frontend assets. This chapter covers the multi-stage Dockerfile, image security practices, and the Docker Compose configuration for development.

---

## 43.1. Multi-Stage Dockerfile

### 43.1.1. The Dockerfile

```dockerfile
# syntax=docker/dockerfile:1.6

# ────────────────────────────────────────────────────────────
# Stage 1: Build Go binary
# ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS go-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Download Go modules (cached layer — only re-runs when go.mod/go.sum changes)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o /awo-server \
    ./cmd/server

# ────────────────────────────────────────────────────────────
# Stage 2: Build frontend assets
# ────────────────────────────────────────────────────────────
FROM node:20-alpine AS frontend-builder

WORKDIR /web

COPY web/package.json web/package-lock.json ./
RUN npm ci --production=false

COPY web/ .
RUN npm run build  # outputs to /web/dist

# ────────────────────────────────────────────────────────────
# Stage 3: Run Atlas migrations check (schema drift verification)
# ────────────────────────────────────────────────────────────
FROM arigaio/atlas:latest AS atlas-verifier
COPY db/migration /migration
# Used in CI for migration linting — not included in final image

# ────────────────────────────────────────────────────────────
# Stage 4: Final minimal runtime image
# ────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

# Copy binary
COPY --from=go-builder /awo-server /awo-server

# Copy frontend assets (served as embedded or static files)
COPY --from=frontend-builder /web/dist /web/dist

# Copy database migrations (applied at startup)
COPY --from=go-builder /app/db/migration /db/migration

# Copy CA certificates (needed for HTTPS calls to M-PESA, Africa's Talking, CBK)
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER nonroot:nonroot
EXPOSE 3000

ENTRYPOINT ["/awo-server"]
```

### 43.1.2. Why Distroless?

`gcr.io/distroless/static:nonroot` contains only the static binary runtime — no shell, no package manager, no libc (since the binary is statically linked with `CGO_ENABLED=0`).

**Security benefits**:
- No shell means no shell injection attack surface
- No package manager means no `apt install malware` from a compromised process
- `nonroot` user means the process cannot escalate to root even if compromised
- Container image vulnerabilities scanners (Trivy, Snyk) report far fewer CVEs against distroless vs Ubuntu/Alpine base images

**Trade-off**: no shell makes debugging harder. For production debugging, use `kubectl exec -it pod -- /bin/sh` on a debug sidecar rather than the main container.

### 43.1.3. Build Arguments for Versioning

```bash
docker build \
  --build-arg VERSION=$(git describe --tags --always) \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t awo-server:$(git rev-parse --short HEAD) \
  .
```

The version and build time are baked into the binary via `-ldflags`. The `/health/live` endpoint returns them:

```json
{
  "status": "ok",
  "version": "v1.4.2-3-gabcd1234",
  "built_at": "2025-07-15T10:23:45Z"
}
```

This makes it trivial to confirm which version is running in each environment during an incident.

---

## 43.2. Image Security

### 43.2.1. Vulnerability Scanning in CI

Every image build runs Trivy:

```yaml
- name: Scan image for vulnerabilities
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: awo-server:${{ github.sha }}
    format: sarif
    output: trivy-results.sarif
    severity: HIGH,CRITICAL
    exit-code: 1  # fail build on HIGH or CRITICAL CVEs
```

### 43.2.2. Non-Root User

The `USER nonroot:nonroot` directive in the Dockerfile ensures the process runs as UID 65532 (the distroless nonroot user). Kubernetes should also enforce this:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65532
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
```

`readOnlyRootFilesystem: true` prevents the process from writing to the container filesystem. Temporary files must go to a mounted `emptyDir` volume.

### 43.2.3. Image Signing

Images are signed with Cosign after build:

```bash
cosign sign --key cosign.key awo-server:${DIGEST}
```

Kubernetes admission webhook (Kyverno or OPA Gatekeeper) verifies the signature before allowing the pod to start. This prevents running unsigned or tampered images in production.

---

## 43.3. Docker Compose for Development

### 43.3.1. Full Development Stack

```yaml
# docker-compose.dev.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: awo
      POSTGRES_PASSWORD: awo_dev
      POSTGRES_DB: awo
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U awo"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  temporal:
    image: temporalio/auto-setup:1.24
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=awo
      - POSTGRES_PWD=awo_dev
      - POSTGRES_SEEDS=postgres
      - DYNAMIC_CONFIG_FILE_PATH=/etc/temporal/config.yaml
    ports:
      - "7233:7233"
    volumes:
      - ./config/temporal-dev.yaml:/etc/temporal/config.yaml

  temporal-ui:
    image: temporalio/ui:2.26
    depends_on:
      - temporal
    environment:
      - TEMPORAL_ADDRESS=temporal:7233
    ports:
      - "8080:8080"

  # Optional: Jaeger for distributed tracing
  jaeger:
    image: jaegertracing/all-in-one:1.57
    ports:
      - "16686:16686"  # UI
      - "4317:4317"    # OTLP gRPC
    environment:
      - COLLECTOR_OTLP_ENABLED=true

volumes:
  postgres_data:
```

### 43.3.2. Starting and Stopping

```bash
# Start infrastructure only (no app server — run the Go server directly)
docker compose -f docker-compose.dev.yml up -d

# Check all services are healthy
docker compose -f docker-compose.dev.yml ps

# Stop and preserve data
docker compose -f docker-compose.dev.yml down

# Stop and destroy data (clean slate)
docker compose -f docker-compose.dev.yml down -v
```

### 43.3.3. Init Container Pattern for Migrations

In Kubernetes, run migrations as an init container before the main container starts:

```yaml
initContainers:
- name: migrate
  image: awo-server:${IMAGE_TAG}
  command: ["/awo-server", "migrate", "--up"]
  env:
  - name: AWO_DB_URL
    valueFrom:
      secretKeyRef:
        name: awo-secrets
        key: AWO_DB_URL
```

The main container only starts after the init container exits successfully (exit code 0). If migrations fail, the deployment is halted — no traffic is sent to a pod with an outdated schema.
