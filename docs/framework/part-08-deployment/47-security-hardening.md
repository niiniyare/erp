---
title: "Chapter 47: Security Hardening"
part: "Part VIII — Deployment and Operations"
chapter: 47
section: "47-security-hardening"
related:
  - "[Chapter 15: Auth and Sessions](../part-03-api/15-auth-sessions.md)"
  - "[Chapter 16: RBAC](../part-03-api/16-rbac.md)"
  - "[Chapter 45: PostgreSQL Operations](45-postgresql-operations.md)"
---

# Chapter 47: Security Hardening

Security hardening transforms a correctly-functioning application into one that is resilient against active attackers. This chapter covers TLS configuration, PostgreSQL access controls, Temporal mTLS, secret management, and the security controls checklist for production deployments.

---

## 47.1. TLS Configuration

### 47.1.1. Certificate Management with cert-manager

In Kubernetes, cert-manager automates certificate provisioning and renewal via Let's Encrypt:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: awo-tls
  namespace: awo-prod
spec:
  secretName: awo-tls-secret
  dnsNames:
    - awo.so
    - "*.awo.so"   # wildcard for tenant subdomains
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  renewBefore: 720h  # renew 30 days before expiry
```

**Wildcard certificate**: tenant subdomains (`acme.awo.so`, `beta.awo.so`) are covered by the wildcard. Let's Encrypt wildcard certificates require DNS-01 challenge, which means the cert-manager must have write access to the DNS zone (Route53, Cloudflare, etc.).

### 47.1.2. TLS Policy

Caddy (reverse proxy) handles TLS termination with a strict policy:

```json
{
  "apps": {
    "tls": {
      "automation": {
        "policies": [{
          "subjects": ["*.awo.so"],
          "key_type": "p256",
          "cipher_suites": [
            "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
            "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256"
          ],
          "protocols": ["tls1.2", "tls1.3"]
        }]
      }
    }
  }
}
```

- **TLS 1.0 and 1.1 disabled**: both are deprecated and vulnerable (BEAST, POODLE)
- **TLS 1.2 with only strong cipher suites**: no RC4, no 3DES, no export ciphers
- **TLS 1.3 preferred**: forward secrecy by design, no cipher suite negotiation

### 47.1.3. HSTS

HTTP Strict Transport Security forces browsers to always use HTTPS:

```go
app.Use(func(c *fiber.Ctx) error {
    c.Set("Strict-Transport-Security",
        "max-age=31536000; includeSubDomains; preload")
    return c.Next()
})
```

`max-age=31536000` (1 year) with `preload` allows the domain to be added to the HSTS preload list, hardcoding HTTPS in browsers before the first visit.

---

## 47.2. PostgreSQL Access Controls

### 47.2.1. Principle of Least Privilege for Database Roles

Three separate database roles:

```sql
-- Application role: CRUD on tenant schemas, read-only on public
CREATE ROLE awo_app WITH LOGIN PASSWORD '${APP_DB_PASSWORD}';
GRANT CONNECT ON DATABASE awo TO awo_app;
-- Dynamic grant when tenant schema is created:
GRANT USAGE ON SCHEMA tenant_acme TO awo_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA tenant_acme TO awo_app;
-- Platform schema (read-only)
GRANT SELECT ON public.tenants TO awo_app;
-- Audit log (insert only)
GRANT INSERT, SELECT ON public.audit_log TO awo_app;

-- Migration role: DDL permissions
CREATE ROLE awo_migrate WITH LOGIN PASSWORD '${MIGRATE_DB_PASSWORD}';
GRANT CREATE ON DATABASE awo TO awo_migrate;

-- Read replica role: read-only for reporting
CREATE ROLE awo_readonly WITH LOGIN PASSWORD '${READONLY_DB_PASSWORD}';
GRANT CONNECT ON DATABASE awo TO awo_readonly;
-- Read-only on all tenant schemas
```

**No superuser access in production**: the application role cannot CREATE/DROP tables, cannot TRUNCATE, cannot access `pg_shadow` (password hashes). A compromised application process cannot drop tables or read other tenants' passwords.

### 47.2.2. pg_hba.conf — Connection Access Control

Restrict which hosts can connect to PostgreSQL:

```
# pg_hba.conf
# TYPE  DATABASE  USER      ADDRESS           METHOD
local   all       postgres                    peer    # local superuser only
host    awo       awo_app   10.0.0.0/8        scram-sha-256
host    awo       awo_migrate 10.0.0.0/8      scram-sha-256
host    all       all       0.0.0.0/0         reject  # deny all other connections
```

`scram-sha-256` is the strongest PostgreSQL authentication method — passwords are never sent in cleartext and the server cannot be impersonated.

### 47.2.3. Row-Level Security for Platform Schema

The `public.tenants` table has RLS to prevent the app role from reading other tenants' records it doesn't need:

```sql
ALTER TABLE public.tenants ENABLE ROW LEVEL SECURITY;

-- App can only see the tenant matching its session's tenant_id
CREATE POLICY tenant_isolation ON public.tenants
    FOR ALL TO awo_app
    USING (id::text = current_setting('app.tenant_id', true));
```

`current_setting('app.tenant_id')` is set via `SET LOCAL app.tenant_id = '{tenant_uuid}'` at the start of each request transaction.

---

## 47.3. Temporal mTLS

### 47.3.1. Why mTLS for Temporal?

Temporal worker connections to the Temporal server handle workflow execution. Without mTLS:
- Any process that can reach the Temporal port can register as a worker and steal workflow tasks
- A stolen workflow task means an attacker can run arbitrary workflow code in the context of a tenant's business operation

### 47.3.2. mTLS Configuration

```go
// Worker connection with mTLS
clientCert, err := tls.LoadX509KeyPair(cfg.TLSCertPath, cfg.TLSKeyPath)
if err != nil {
    return err
}

caCert, err := os.ReadFile(cfg.TLSCAPath)
if err != nil {
    return err
}
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{clientCert},
    RootCAs:      caCertPool,
    ServerName:   cfg.TemporalHost,
}

temporalClient, err := client.Dial(client.Options{
    HostPort:  cfg.TemporalHost,
    Namespace: cfg.TemporalNamespace,
    ConnectionOptions: client.ConnectionOptions{
        TLS: tlsConfig,
    },
})
```

Worker certificates are issued by an internal CA (or Temporal Cloud provides them). Only workers with valid certificates signed by the trusted CA can connect.

---

## 47.4. Secret Management

### 47.4.1. What Must Never Be in Git

```
.env files
*.pem (TLS private keys)
*_secret* (anything with "secret" in filename)
config/production.yaml (if it contains passwords)
```

`.gitignore` enforces this:
```
.env
.env.*
*.pem
*.key
!*.pub  # public keys are OK
```

### 47.4.2. Secret Rotation

Secrets that must be rotated periodically:

| Secret | Rotation frequency | How |
|---|---|---|
| DB passwords | Quarterly | Update Kubernetes Secret, rolling restart |
| Session secret | Annually (all sessions invalidated) | Update and restart |
| M-PESA credentials | When compromised | Revoke in Daraja portal, update immediately |
| TLS certificates | Auto-renewed by cert-manager | No manual action |
| Cosign signing key | Annually | Re-sign existing images with new key |

### 47.4.3. Kubernetes Secret Encryption at Rest

By default, Kubernetes stores Secrets base64-encoded (not encrypted) in etcd. Enable envelope encryption:

```yaml
# kube-apiserver --encryption-provider-config
apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
resources:
- resources:
  - secrets
  providers:
  - aescbc:
      keys:
      - name: key1
        secret: ${BASE64_32_BYTE_KEY}
  - identity: {}
```

With this configured, Secret values are AES-256-CBC encrypted in etcd. Access to etcd without the encryption key yields ciphertext, not secret values.

---

## 47.5. Application Security Controls

### 47.5.1. Security Headers

```go
app.Use(func(c *fiber.Ctx) error {
    // Prevent MIME type sniffing
    c.Set("X-Content-Type-Options", "nosniff")
    // Prevent clickjacking
    c.Set("X-Frame-Options", "DENY")
    // Referrer policy — don't leak tenant URLs to third parties
    c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
    // Permissions policy — disable unused browser APIs
    c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(self)")
    // CSP — restrict script sources
    c.Set("Content-Security-Policy",
        "default-src 'self'; "+
        "script-src 'self' 'nonce-"+c.Locals("cspNonce").(string)+"'; "+
        "img-src 'self' data: blob:; "+
        "connect-src 'self'")
    return c.Next()
})
```

### 47.5.2. SQL Injection Prevention

Awo uses parameterised queries exclusively through the ent ORM and the repository filter DSL. Raw SQL is used only in migrations (which run as the migrate role, not as user input).

The filter DSL never concatenates user input into SQL:
```go
// Safe: filter DSL builds parameterised query
filter.Eq("customer_name", userInput)
// → WHERE customer_name = $1 with args[0] = userInput

// Never: string concatenation into SQL
db.Raw("WHERE name = '" + userInput + "'")  // SQL injection
```

### 47.5.3. Input Validation

All API inputs are validated before processing:

```go
type CreateInvoiceInput struct {
    CustomerID  uuid.UUID       `json:"customer_id" validate:"required,uuid"`
    PostingDate time.Time       `json:"posting_date" validate:"required"`
    Lines       []InvoiceLine   `json:"lines" validate:"required,min=1,max=500,dive"`
    Currency    string          `json:"currency" validate:"omitempty,len=3,uppercase"`
}

func (h *InvoiceHandler) Create(c *fiber.Ctx) error {
    var input CreateInvoiceInput
    if err := c.BodyParser(&input); err != nil {
        return errs.NewValidationError("INVALID_JSON", "request body is not valid JSON")
    }
    if err := h.validator.Struct(input); err != nil {
        return mapValidationErrors(err)
    }
    // ... proceed
}
```

### 47.5.4. Production Security Checklist

Before go-live, verify:

- [ ] All API endpoints require authentication (no anonymous access to business data)
- [ ] Tenant isolation verified: test that tenant A cannot access tenant B's data via any API endpoint
- [ ] Rate limiting active on all public endpoints
- [ ] TLS 1.2+ enforced, TLS 1.0/1.1 blocked (verify with `testssl.sh`)
- [ ] Security headers present (verify with `securityheaders.com`)
- [ ] Secrets not in git or container image (scan with `trufflesecurity/trufflehog`)
- [ ] Vulnerability scan of Docker image shows no HIGH/CRITICAL CVEs
- [ ] Database role does not have superuser privileges
- [ ] Audit log table has no DELETE permission for app role
- [ ] Temporal workers use mTLS
- [ ] Session cookies have `HttpOnly`, `Secure`, `SameSite=Lax`
- [ ] Admin accounts have MFA enforced
- [ ] PostgreSQL `pg_hba.conf` restricts access to Kubernetes pod CIDR only
- [ ] Backup encryption keys are stored separately from backups
- [ ] Incident response runbook is documented and tested
