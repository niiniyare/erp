package iam_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"awo.so/awo/auth"
	contribredis "awo.so/awo/contrib/redis"
	"awo.so/awo/platform/iam"
	"awo.so/awo/runtime"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *goredis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })
	return mr, client
}

// apiTokenCacheKey replicates the cache key formula used in service.go.
// Format: "iam:api_token:{sha256hex(rawToken)}"
func apiTokenCacheKey(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return "iam:api_token:" + hex.EncodeToString(sum[:])
}

func storeSessionInRedis(t *testing.T, mr *miniredis.Miniredis, session *auth.Session) {
	t.Helper()
	data, err := json.Marshal(session)
	require.NoError(t, err)
	require.NoError(t, mr.Set("session:"+session.Token, string(data)))
	mr.SetTTL("session:"+session.Token, time.Hour)
}

func storeAPITokenCache(t *testing.T, mr *miniredis.Miniredis, rawToken string, session *auth.Session) {
	t.Helper()
	data, err := json.Marshal(session)
	require.NoError(t, err)
	require.NoError(t, mr.Set(apiTokenCacheKey(rawToken), string(data)))
	mr.SetTTL(apiTokenCacheKey(rawToken), 60*time.Second)
}

// ── ValidateToken ─────────────────────────────────────────────────────────────

func TestValidateToken_ValidSession_Succeeds(t *testing.T) {
	mr, client := newTestRedis(t)
	svc := &iam.AuthService{Sessions: contribredis.NewSessionStore(client), Cache: contribredis.New(client)}

	session := &auth.Session{
		Token:     "good-token",
		UserID:    uuid.New(),
		TenantID:  uuid.New(),
		Roles:     []string{"role:finance.viewer"},
		IssuedAt:  time.Now().Add(-time.Minute),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	storeSessionInRedis(t, mr, session)

	got, err := svc.ValidateToken(context.Background(), "good-token")
	require.NoError(t, err)
	assert.Equal(t, session.UserID, got.UserID)
	assert.Equal(t, session.TenantID, got.TenantID)
}

func TestValidateToken_MissingKey_Returns401(t *testing.T) {
	_, client := newTestRedis(t)
	svc := &iam.AuthService{Sessions: contribredis.NewSessionStore(client), Cache: contribredis.New(client)}

	_, err := svc.ValidateToken(context.Background(), "nonexistent-token")
	require.Error(t, err)

	var be *runtime.BusinessError
	require.ErrorAs(t, err, &be)
	assert.Equal(t, 401, be.Status)
	assert.Equal(t, "iam.session.not_found", be.Code)
}

func TestValidateToken_WallClockExpired_Returns401(t *testing.T) {
	mr, client := newTestRedis(t)
	svc := &iam.AuthService{Sessions: contribredis.NewSessionStore(client), Cache: contribredis.New(client)}

	// Key present in Redis but session timestamp is already expired.
	session := &auth.Session{
		Token:     "stale-token",
		UserID:    uuid.New(),
		TenantID:  uuid.New(),
		Roles:     []string{"role:finance.viewer"},
		IssuedAt:  time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-30 * time.Minute), // expired
	}
	storeSessionInRedis(t, mr, session)

	_, err := svc.ValidateToken(context.Background(), "stale-token")
	require.Error(t, err)

	var be *runtime.BusinessError
	require.ErrorAs(t, err, &be)
	assert.Equal(t, 401, be.Status)
	assert.Equal(t, "iam.session.expired", be.Code)
}

func TestValidateToken_RedisDown_Returns503(t *testing.T) {
	mr, client := newTestRedis(t)
	svc := &iam.AuthService{Sessions: contribredis.NewSessionStore(client), Cache: contribredis.New(client)}

	mr.Close() // simulate Redis outage

	_, err := svc.ValidateToken(context.Background(), "any-token")
	require.Error(t, err)

	var be *runtime.BusinessError
	require.ErrorAs(t, err, &be)
	assert.Equal(t, 503, be.Status)
	assert.Equal(t, "iam.service_unavailable", be.Code)
}

// ── ValidateAPIToken ──────────────────────────────────────────────────────────

func TestValidateAPIToken_CacheHit_ReturnsCachedSession(t *testing.T) {
	mr, client := newTestRedis(t)
	// No DB wired — DB path must not be reached on a cache hit.
	svc := &iam.AuthService{Sessions: contribredis.NewSessionStore(client), Cache: contribredis.New(client)}

	rawToken := "svc-api-key-abc"
	tenantID := uuid.New()
	saID := uuid.New()

	cached := &auth.Session{
		Token:            rawToken,
		ServiceAccountID: saID,
		TenantID:         tenantID,
		Roles:            []string{"role:api-client"},
		IssuedAt:         time.Now().Add(-time.Minute),
		ExpiresAt:        time.Now().Add(time.Hour),
	}
	storeAPITokenCache(t, mr, rawToken, cached)

	got, err := svc.ValidateAPIToken(context.Background(), rawToken, tenantID)
	require.NoError(t, err)
	assert.Equal(t, saID, got.ServiceAccountID)
	assert.Equal(t, tenantID, got.TenantID)
	assert.Equal(t, []string{"role:api-client"}, got.Roles)
}

func TestValidateAPIToken_ExpiredCacheEntry_Returns401(t *testing.T) {
	mr, client := newTestRedis(t)
	svc := &iam.AuthService{Sessions: contribredis.NewSessionStore(client), Cache: contribredis.New(client)}

	rawToken := "expired-api-key"
	tenantID := uuid.New()

	expired := &auth.Session{
		Token:            rawToken,
		ServiceAccountID: uuid.New(),
		TenantID:         tenantID,
		Roles:            []string{"role:api-client"},
		IssuedAt:         time.Now().Add(-2 * time.Hour),
		ExpiresAt:        time.Now().Add(-time.Minute), // expired
	}
	storeAPITokenCache(t, mr, rawToken, expired)

	_, err := svc.ValidateAPIToken(context.Background(), rawToken, tenantID)
	require.Error(t, err)

	var be *runtime.BusinessError
	require.ErrorAs(t, err, &be)
	assert.Equal(t, 401, be.Status)
	assert.Equal(t, "iam.api_token.expired", be.Code)
}

func TestValidateAPIToken_RedisDown_Returns503(t *testing.T) {
	mr, client := newTestRedis(t)
	svc := &iam.AuthService{Sessions: contribredis.NewSessionStore(client), Cache: contribredis.New(client)}

	mr.Close()

	_, err := svc.ValidateAPIToken(context.Background(), "any-key", uuid.New())
	require.Error(t, err)

	var be *runtime.BusinessError
	require.ErrorAs(t, err, &be)
	assert.Equal(t, 503, be.Status)
	assert.Equal(t, "iam.service_unavailable", be.Code)
}
