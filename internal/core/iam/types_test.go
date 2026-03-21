package iam

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ---------------------------------------------------------------------------
// TypesSuite — pure unit tests, no database required.
// ---------------------------------------------------------------------------

type TypesSuite struct{ suite.Suite }

func TestTypesSuite(t *testing.T) { suite.Run(t, new(TypesSuite)) }

// ---- Subject helpers -------------------------------------------------------

func (s *TypesSuite) TestPlatformSubject() {
	s.Equal("platform:usr_ops_001", PlatformSubject("usr_ops_001"))
}

func (s *TypesSuite) TestTenantSubject() {
	s.Equal("tenant:usr_amina", TenantSubject("usr_amina"))
}

func (s *TypesSuite) TestPortalSubject() {
	s.Equal("portal:cust_nairobi_001", PortalSubject("cust_nairobi_001"))
}

func (s *TypesSuite) TestAPISubject() {
	s.Equal("api:cli_quickbooks_sync", APISubject("cli_quickbooks_sync"))
}

func (s *TypesSuite) TestSubjectPrefixesAreDistinct() {
	id := "same_id"
	subjects := []string{
		PlatformSubject(id),
		TenantSubject(id),
		PortalSubject(id),
		APISubject(id),
	}
	seen := make(map[string]bool)
	for _, sub := range subjects {
		s.False(seen[sub], "duplicate subject: %s", sub)
		seen[sub] = true
	}
}

// ---- Domain helpers -------------------------------------------------------

func (s *TypesSuite) TestTenantDomain_ReturnsRawID() {
	id := "a1b2c3d4-5678-90ab-cdef-0123456789ab"
	s.Equal(id, TenantDomain(id))
}

func (s *TypesSuite) TestPortalDomain_AppendsPortalSuffix() {
	id := "a1b2c3d4-5678-90ab-cdef-0123456789ab"
	s.Equal(id+":portal", PortalDomain(id))
}

func (s *TypesSuite) TestAPIDomain_AppendsAPISuffix() {
	id := "a1b2c3d4-5678-90ab-cdef-0123456789ab"
	s.Equal(id+":api", APIDomain(id))
}

func (s *TypesSuite) TestDomainBuilders_ProduceDistinctValues() {
	id := "same-tenant-uuid"
	domains := []string{
		TenantDomain(id),
		PortalDomain(id),
		APIDomain(id),
	}
	seen := make(map[string]bool)
	for _, d := range domains {
		s.False(seen[d], "duplicate domain: %s", d)
		seen[d] = true
	}
}

// ---- Constants -------------------------------------------------------------

func (s *TypesSuite) TestDomainPlatformConstant() {
	s.Equal("_platform_", DomainPlatform)
}

func (s *TypesSuite) TestLocalsKeyPrincipalConstant() {
	s.Equal("authz_principal", LocalsKeyPrincipal)
}

// ---- AssignOpt functional options -----------------------------------------

func (s *TypesSuite) TestWithExpiry_SetsExpiresAt() {
	expiry := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)
	o := &assignOpts{}
	WithExpiry(expiry)(o)

	s.Require().NotNil(o.expiresAt)
	s.Equal(expiry, *o.expiresAt)
}

func (s *TypesSuite) TestWithAssignedBy_SetsField() {
	o := &assignOpts{}
	WithAssignedBy("tenant:usr_ceo_001")(o)
	s.Equal("tenant:usr_ceo_001", o.assignedBy)
}

func (s *TypesSuite) TestWithDelegatedBy_SetsField() {
	o := &assignOpts{}
	WithDelegatedBy("tenant:usr_cfo")(o)
	s.Equal("tenant:usr_cfo", o.delegatedBy)
}

func (s *TypesSuite) TestMultipleOpts_AllApplied() {
	expiry := time.Now().Add(24 * time.Hour)
	o := &assignOpts{}
	opts := []AssignOpt{
		WithExpiry(expiry),
		WithAssignedBy("assigner"),
		WithDelegatedBy("delegator"),
	}
	for _, opt := range opts {
		opt(o)
	}

	s.Require().NotNil(o.expiresAt)
	s.Equal(expiry.Unix(), o.expiresAt.Unix())
	s.Equal("assigner", o.assignedBy)
	s.Equal("delegator", o.delegatedBy)
}

func (s *TypesSuite) TestLastOptWins_WhenSameFieldSetTwice() {
	o := &assignOpts{}
	WithAssignedBy("first")(o)
	WithAssignedBy("second")(o)
	s.Equal("second", o.assignedBy)
}

func (s *TypesSuite) TestWithExpiry_NilByDefault() {
	o := &assignOpts{}
	// No opts applied
	s.Nil(o.expiresAt)
}

// ---- nullableString helper ------------------------------------------------

func (s *TypesSuite) TestNullableString_EmptyReturnsNil() {
	s.Nil(nullableString(""))
}

func (s *TypesSuite) TestNullableString_NonEmptyReturnsPointer() {
	p := nullableString("hello")
	s.Require().NotNil(p)
	s.Equal("hello", *p)
}
