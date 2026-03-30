# AWO API Test Suite

Modular, interactive bash test runner for the AWO ERP REST API.

## Quick start

```bash
# Seed the DB first (creates tenant + admin user)
bash scripts/seed.sh

# Interactive menu
TENANT_ID=<uuid> bash scripts/api/run.sh

# Run all suites non-interactively
TENANT_ID=<uuid> bash scripts/api/run.sh --all

# Run one suite
TENANT_ID=<uuid> bash scripts/api/run.sh --service iam

# List registered services
bash scripts/api/run.sh --list
```

## Environment variables

| Variable         | Default                   | Notes                           |
|------------------|---------------------------|---------------------------------|
| `BASE_URL`       | `http://localhost:8080`   |                                 |
| `ADMIN_EMAIL`    | `admin@platform.local`    | Created by `seed.sh`            |
| `ADMIN_PASSWORD` | `Admin1234!`              |                                 |
| `TENANT_ID`      | _(empty)_                 | Required for tenant-scoped ops  |

## Structure

```
scripts/api/
├── run.sh              ← entrypoint (interactive + --all + --service)
├── core/
│   ├── env.sh          ← config + per-session random values
│   ├── auth.sh         ← api(), apiv(), login(), logout()
│   └── utils.sh        ← ok/fail/info, assert helpers, jf()
├── services/
│   ├── iam.sh          ← auth, MFA, password reset, API keys, audit
│   ├── users.sh        ← user CRUD
│   ├── tenants.sh      ← tenant lifecycle
│   ├── finance.sh      ← accounts, transactions, reports
│   └── orders.sh       ← stub (shows how to add a module)
├── data/               ← runtime cookie jars (gitignored)
└── logs/               ← test run output (gitignored)
```

## Adding a new module

Create `scripts/api/services/mymodule.sh` — `run.sh` auto-discovers it:

```bash
#!/usr/bin/env bash
MENU+=("mymodule:MyModule — short description:run_mymodule")

run_mymodule() {
  header "MyModule"
  reset_counters
  require_auth || { warn "Skipping (not authenticated)"; return 1; }

  sep; info "List items"
  local resp
  resp=$(apiv GET /api/v1/mymodule)
  split_resp "$resp"
  assert_any "GET /api/v1/mymodule" "$RESP_STATUS" "200" "403" || true

  sep
  [[ $SUITE_FAILURES -eq 0 ]] && ok "MyModule suite passed" \
                               || fail "$SUITE_FAILURES failure(s)"
}
```

That's all — no registration required beyond the `MENU+=` line.

## Per-session random data

Every run generates a fresh `SESSION_ID` (e.g. `a3f8b2c1`) and derives:

| Variable           | Example                    |
|--------------------|----------------------------|
| `TEST_EMAIL`       | `tuser_a3f8b2@test.local`  |
| `TEST_USERNAME`    | `tuser_a3f8b2`             |
| `TEST_PASSWORD`    | `Tst_a3f8b21!`             |
| `TEST_TENANT_NAME` | `TestCo_a3f8b2`            |

No collisions between parallel or back-to-back runs against the same DB.

## Core API reference

```bash
# HTTP
api   METHOD path [curl-args]        # request with cookie jar + headers
apiv  METHOD path [curl-args]        # same + appends \n<status_code>
api_bearer TOKEN METHOD path         # Bearer-token request (no cookie)

# Session
login [email] [password]             # authenticate → AUTH_ACTIVE=true
logout                               # end session
require_auth                         # login if not already authenticated
with_anon_session <fn>               # run fn with blank cookie, then restore

# Assertions (increment $SUITE_FAILURES on mismatch)
split_resp "$resp"                   # sets $RESP_BODY and $RESP_STATUS
assert_status "label" 200 "$s"       # exact status match
assert_any "label" "$s" 200 403      # accept any listed code

# JSON
jf "$body" "data.id"                 # extract nested field (dot notation)
json_len "$body" "data"              # count items in a JSON array
```
