---
title: Security Incident Response
portal: 7 — Security
section: 07-security
audience: [security, sre, devops]
related:
  - "[Security Overview](01-security-overview.md)"
  - "[Authentication Flows](02-authentication-flows.md)"
  - "[Service Down Runbook](../09-operations/02-service-down.md)"
---

# Security Incident Response

## Severity Classification

| Severity | Examples | Response SLA |
|----------|---------|-------------|
| P0 — Critical | Data breach, credential leak, RLS bypass confirmed | Immediate — escalate in < 15 min |
| P1 — High | Unauthorized access, session token exposure, DoS | < 1 hour |
| P2 — Medium | Repeated brute force, anomalous API access pattern | < 4 hours |
| P3 — Low | Outdated dependency with CVE, minor misconfiguration | < 48 hours |

## Immediate Actions by Scenario

### Suspected Session Token Compromise

If a session token is believed stolen or exposed:

```bash
# Revoke the specific user's sessions
awoctl sessions revoke-all --user-id {uuid}

# If tenant-wide compromise suspected
awoctl sessions revoke-tenant --tenant-id {uuid}
```

Notify the user to change their password after revocation.

### Credential Leak (DB Password, Session Secret)

1. **Rotate immediately** — generate new credential, update Kubernetes Secret
2. **Rolling restart** — `kubectl rollout restart deployment/awoerp-server -n production`
3. **Invalidate sessions** — if `SESSION_SECRET` leaked, all tokens are potentially forged

```bash
# Force all sessions to re-login by rotating SESSION_SECRET
kubectl set env deployment/awoerp-server SESSION_SECRET="$(openssl rand -hex 64)" -n production
kubectl rollout restart deployment/awoerp-server -n production
```

Rotating `SESSION_SECRET` invalidates all current sessions — all users must re-login. Communicate via status page / email.

### Suspected RLS Bypass

If data from one tenant appears accessible by another:

1. **Immediately disable the affected endpoint** — feature flag off, or route-level block
2. **Preserve evidence** — do not restart or modify DB until logs are captured
3. **Check PostgreSQL logs** for queries without `app.tenant_id` set
4. **Escalate to DBA** — check `pg_stat_activity` and recent query patterns

```sql
-- Check if any rows have incorrect tenant_id (cross-check against known tenant list)
SELECT DISTINCT tenant_id FROM contracts
WHERE tenant_id NOT IN (SELECT id FROM tenants);

-- Check if RLS is still enabled on affected tables
SELECT tablename, rowsecurity FROM pg_tables
WHERE schemaname = 'public' AND rowsecurity = false;
```

### Brute Force on Login

The login endpoint is rate-limited to 10/min/IP. If that's being bypassed:

```bash
# Check IP patterns
kubectl logs -l app=awoerp-server -n production --since=30m \
  | grep '/auth/login' \
  | jq -r '.ip' | sort | uniq -c | sort -rn | head -20
```

Block attacking IPs at the ingress/load balancer level if rate limiting is insufficient.

### Vulnerability Disclosure

If a CVE is reported affecting a dependency:

1. Run `govulncheck ./...` to assess actual exposure
2. If exploitable path exists: patch within 24 hours for P0/P1
3. If no exploitable path: patch within next release cycle
4. Update `go.sum` after patching

## Post-Incident

### Required for P0/P1

1. **Incident log** — timeline of events, who was involved, what actions were taken
2. **Root cause analysis** — what failed, why
3. **Remediation** — what was fixed
4. **Prevention** — what process/code change prevents recurrence
5. **Customer notification** — if PII or business data was exposed, legal review required

### Audit Log Review

For any suspected unauthorized access:

```sql
-- All actions by a user in a time window
SELECT * FROM audit_log
WHERE user_id = $1
  AND created_at BETWEEN $2 AND $3
ORDER BY created_at;

-- Unusual actions (approvals, terminations) in time window
SELECT * FROM audit_log
WHERE action IN ('contracts.contract.approve', 'contracts.contract.terminate', 'iam.user.create')
  AND created_at BETWEEN $1 AND $2
ORDER BY created_at;
```

## Escalation Path

| Who | Contact |
|-----|---------|
| On-call SRE | PagerDuty rotation |
| Backend lead | Direct message |
| Security team | `security@` email alias |
| Legal/compliance | For PII breach notification |
| CTO | P0 only |

For P0 incidents: do not wait to escalate. Page immediately and investigate in parallel.

## Do Not

- Do not attempt to cover up or minimize a security incident
- Do not delete logs or evidence before incident is resolved
- Do not reuse compromised credentials — always rotate fully
- Do not patch security vulnerabilities without incident documentation
- Do not assume an incident is resolved until root cause is confirmed
