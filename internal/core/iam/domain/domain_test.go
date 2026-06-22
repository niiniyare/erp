package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"awo.so/internal/core/iam/domain"
)

// =============================================================================
// ActorType Suite
// =============================================================================

type ActorTypeSuite struct{ suite.Suite }

func TestActorTypeSuite(t *testing.T) { suite.Run(t, new(ActorTypeSuite)) }

func (s *ActorTypeSuite) TestPlatformUserTypes() {
	s.Equal(domain.ActorPlatform, domain.ActorTypeFromUserType("SYSADMIN"))
	s.Equal(domain.ActorPlatform, domain.ActorTypeFromUserType("PLATFORM"))
}

func (s *ActorTypeSuite) TestPortalUserTypes() {
	s.Equal(domain.ActorPortal, domain.ActorTypeFromUserType("PORTAL"))
	s.Equal(domain.ActorPortal, domain.ActorTypeFromUserType("CUSTOMER"))
}

func (s *ActorTypeSuite) TestAPIUserTypes() {
	s.Equal(domain.ActorAPI, domain.ActorTypeFromUserType("API"))
	s.Equal(domain.ActorAPI, domain.ActorTypeFromUserType("SERVICE"))
}

func (s *ActorTypeSuite) TestTenantFallback() {
	for _, t := range []string{"INTERNAL", "EMPLOYEE", "UNKNOWN", ""} {
		s.Equal(domain.ActorTenant, domain.ActorTypeFromUserType(t), "type=%q", t)
	}
}

// =============================================================================
// Subject / Domain Helper Suite
// =============================================================================

type SubjectDomainSuite struct{ suite.Suite }

func TestSubjectDomainSuite(t *testing.T) { suite.Run(t, new(SubjectDomainSuite)) }

func (s *SubjectDomainSuite) TestSubjectHelpers() {
	id := "abc-123"
	s.Equal("platform:abc-123", domain.PlatformSubject(id))
	s.Equal("tenant:abc-123", domain.TenantSubject(id))
	s.Equal("portal:abc-123", domain.PortalSubject(id))
	s.Equal("api:abc-123", domain.APISubject(id))
}

func (s *SubjectDomainSuite) TestDomainHelpers() {
	tid := "tenant-uuid"
	s.Equal(domain.DomainPlatform, domain.PlatformDomain())
	s.Equal("_platform_", domain.DomainPlatform)
	s.Equal(tid, domain.TenantDomain(tid))
	s.Equal(tid+":portal", domain.PortalDomain(tid))
	s.Equal(tid+":api", domain.APIDomain(tid))
}

// =============================================================================
// RoleAssignment Suite
// =============================================================================

type RoleAssignmentSuite struct{ suite.Suite }

func TestRoleAssignmentSuite(t *testing.T) { suite.Run(t, new(RoleAssignmentSuite)) }

func (s *RoleAssignmentSuite) TestIsExpired_NoExpiry() {
	ra := &domain.RoleAssignment{IsActive: true, ExpiresAt: nil}
	s.False(ra.IsExpired())
}

func (s *RoleAssignmentSuite) TestIsExpired_Future() {
	future := time.Now().Add(time.Hour)
	ra := &domain.RoleAssignment{IsActive: true, ExpiresAt: &future}
	s.False(ra.IsExpired())
}

func (s *RoleAssignmentSuite) TestIsExpired_Past() {
	past := time.Now().Add(-time.Hour)
	ra := &domain.RoleAssignment{IsActive: true, ExpiresAt: &past}
	s.True(ra.IsExpired())
}

func (s *RoleAssignmentSuite) TestIsEffective_ActiveNotExpired() {
	future := time.Now().Add(time.Hour)
	ra := &domain.RoleAssignment{IsActive: true, ExpiresAt: &future}
	s.True(ra.IsEffective())
}

func (s *RoleAssignmentSuite) TestIsEffective_InactiveNotExpired() {
	ra := &domain.RoleAssignment{IsActive: false, ExpiresAt: nil}
	s.False(ra.IsEffective())
}

func (s *RoleAssignmentSuite) TestIsEffective_ActiveExpired() {
	past := time.Now().Add(-time.Hour)
	ra := &domain.RoleAssignment{IsActive: true, ExpiresAt: &past}
	s.False(ra.IsEffective())
}

func (s *RoleAssignmentSuite) TestIsEffective_ActiveNoExpiry() {
	ra := &domain.RoleAssignment{IsActive: true, ExpiresAt: nil}
	s.True(ra.IsEffective())
}

// =============================================================================
// AssignOpts Suite
// =============================================================================

type AssignOptsSuite struct{ suite.Suite }

func TestAssignOptsSuite(t *testing.T) { suite.Run(t, new(AssignOptsSuite)) }

func (s *AssignOptsSuite) TestApplyEmpty() {
	ao := domain.ApplyAssignOpts(nil)
	s.Nil(ao.ExpiresAt)
	s.Empty(ao.AssignedBy)
	s.Empty(ao.DelegatedBy)
}

func (s *AssignOptsSuite) TestWithExpiry() {
	t := time.Now().Add(24 * time.Hour)
	ao := domain.ApplyAssignOpts([]domain.AssignOpt{domain.WithExpiry(t)})
	s.Require().NotNil(ao.ExpiresAt)
	s.True(ao.ExpiresAt.Equal(t))
}

func (s *AssignOptsSuite) TestWithAssignedBy() {
	ao := domain.ApplyAssignOpts([]domain.AssignOpt{domain.WithAssignedBy("platform:system")})
	s.Equal("platform:system", ao.AssignedBy)
}

func (s *AssignOptsSuite) TestWithDelegatedBy() {
	ao := domain.ApplyAssignOpts([]domain.AssignOpt{domain.WithDelegatedBy("tenant:mgr")})
	s.Equal("tenant:mgr", ao.DelegatedBy)
}

func (s *AssignOptsSuite) TestCombinedOpts() {
	exp := time.Now().Add(time.Hour)
	ao := domain.ApplyAssignOpts([]domain.AssignOpt{
		domain.WithAssignedBy("platform:system"),
		domain.WithDelegatedBy("tenant:manager"),
		domain.WithExpiry(exp),
	})
	s.Equal("platform:system", ao.AssignedBy)
	s.Equal("tenant:manager", ao.DelegatedBy)
	s.Require().NotNil(ao.ExpiresAt)
	s.True(ao.ExpiresAt.Equal(exp))
}

// =============================================================================
// Session Suite
// =============================================================================

type SessionSuite struct{ suite.Suite }

func TestSessionSuite(t *testing.T) { suite.Run(t, new(SessionSuite)) }

func makeSession(userType string, isActive bool, expiresIn time.Duration) *domain.Session {
	return &domain.Session{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TenantID:  uuid.New(),
		UserType:  userType,
		IsActive:  isActive,
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func (s *SessionSuite) TestActorType_InternalIsTenant() {
	sess := makeSession("INTERNAL", true, time.Hour)
	s.Equal(domain.ActorTenant, sess.ActorType())
}

func (s *SessionSuite) TestActorType_SysadminIsPlatform() {
	sess := makeSession("SYSADMIN", true, time.Hour)
	s.Equal(domain.ActorPlatform, sess.ActorType())
}

func (s *SessionSuite) TestIsExpired_Future() {
	sess := makeSession("INTERNAL", true, time.Hour)
	s.False(sess.IsExpired())
}

func (s *SessionSuite) TestIsExpired_Past() {
	sess := makeSession("INTERNAL", true, -time.Hour)
	s.True(sess.IsExpired())
}

func (s *SessionSuite) TestIsValid_ActiveNotExpired() {
	sess := makeSession("INTERNAL", true, time.Hour)
	s.True(sess.IsValid())
}

func (s *SessionSuite) TestIsValid_InactiveNotExpired() {
	sess := makeSession("INTERNAL", false, time.Hour)
	s.False(sess.IsValid())
}

func (s *SessionSuite) TestIsValid_ActiveExpired() {
	sess := makeSession("INTERNAL", true, -time.Hour)
	s.False(sess.IsValid())
}

// =============================================================================
// ResolvedSession Suite
// =============================================================================

type ResolvedSessionSuite struct{ suite.Suite }

func TestResolvedSessionSuite(t *testing.T) { suite.Run(t, new(ResolvedSessionSuite)) }

func makeResolved(userType string) *domain.ResolvedSession {
	return &domain.ResolvedSession{
		UserID:   uuid.New(),
		UserType: userType,
		TenantID: uuid.New(),
		Configuration: domain.Configuration{
			Flags:    map[string]bool{"hr.payroll_v2": true, "disabled_flag": false},
			Settings: map[string]string{"session_ttl": "8", "bool_setting": "true", "float_setting": "3.14"},
			Prefs:    map[string]string{},
		},
	}
}

func (s *ResolvedSessionSuite) TestToPrincipal_Tenant() {
	rs := makeResolved("INTERNAL")
	p := rs.ToPrincipal()
	s.Equal("tenant:"+rs.UserID.String(), p.Subject)
	s.Equal(rs.TenantID.String(), p.Domain)
}

func (s *ResolvedSessionSuite) TestToPrincipal_Platform() {
	rs := makeResolved("SYSADMIN")
	p := rs.ToPrincipal()
	s.Equal("platform:"+rs.UserID.String(), p.Subject)
	s.Equal(domain.DomainPlatform, p.Domain)
}

func (s *ResolvedSessionSuite) TestToPrincipal_Portal() {
	rs := makeResolved("PORTAL")
	p := rs.ToPrincipal()
	s.Equal("portal:"+rs.UserID.String(), p.Subject)
	s.Equal(rs.TenantID.String()+":portal", p.Domain)
}

func (s *ResolvedSessionSuite) TestToPrincipal_API() {
	rs := makeResolved("API")
	p := rs.ToPrincipal()
	s.Equal("api:"+rs.UserID.String(), p.Subject)
	s.Equal(rs.TenantID.String()+":api", p.Domain)
}

func (s *ResolvedSessionSuite) TestIsPortal_True() {
	s.True(makeResolved("PORTAL").IsPortal())
}

func (s *ResolvedSessionSuite) TestIsPortal_False() {
	s.False(makeResolved("INTERNAL").IsPortal())
}

func (s *ResolvedSessionSuite) TestIsPlatform_True() {
	s.True(makeResolved("SYSADMIN").IsPlatform())
}

func (s *ResolvedSessionSuite) TestIsPlatform_False() {
	s.False(makeResolved("INTERNAL").IsPlatform())
}

func (s *ResolvedSessionSuite) TestFeatureEnabled_True() {
	s.True(makeResolved("INTERNAL").FeatureEnabled("hr.payroll_v2"))
}

func (s *ResolvedSessionSuite) TestFeatureEnabled_FalseFlag() {
	s.False(makeResolved("INTERNAL").FeatureEnabled("disabled_flag"))
}

func (s *ResolvedSessionSuite) TestFeatureEnabled_Missing() {
	s.False(makeResolved("INTERNAL").FeatureEnabled("nonexistent"))
}

func (s *ResolvedSessionSuite) TestFeatureEnabled_NilReceiver() {
	var rs *domain.ResolvedSession
	s.False(rs.FeatureEnabled("anything"))
}

func (s *ResolvedSessionSuite) TestSettingString_Present() {
	s.Equal("8", makeResolved("INTERNAL").SettingString("session_ttl", "0"))
}

func (s *ResolvedSessionSuite) TestSettingString_Default() {
	s.Equal("fallback", makeResolved("INTERNAL").SettingString("missing", "fallback"))
}

func (s *ResolvedSessionSuite) TestSettingBool_True() {
	s.True(makeResolved("INTERNAL").SettingBool("bool_setting", false))
}

func (s *ResolvedSessionSuite) TestSettingBool_Default() {
	s.True(makeResolved("INTERNAL").SettingBool("missing", true))
}

func (s *ResolvedSessionSuite) TestSettingInt_Valid() {
	s.Equal(8, makeResolved("INTERNAL").SettingInt("session_ttl", 0))
}

func (s *ResolvedSessionSuite) TestSettingInt_Default() {
	s.Equal(99, makeResolved("INTERNAL").SettingInt("missing", 99))
}

func (s *ResolvedSessionSuite) TestSettingDecimal_Valid() {
	s.InDelta(3.14, makeResolved("INTERNAL").SettingDecimal("float_setting", 0), 0.001)
}

func (s *ResolvedSessionSuite) TestSettingDecimal_Default() {
	s.InDelta(1.5, makeResolved("INTERNAL").SettingDecimal("missing", 1.5), 0.001)
}

// =============================================================================
// User Suite
// =============================================================================

type UserSuite struct{ suite.Suite }

func TestUserSuite(t *testing.T) { suite.Run(t, new(UserSuite)) }

func (s *UserSuite) TestIsLocked_NoLockout() {
	u := &domain.User{LockoutUntil: nil}
	s.False(u.IsLocked())
}

func (s *UserSuite) TestIsLocked_PastLockout() {
	past := time.Now().Add(-time.Minute)
	u := &domain.User{LockoutUntil: &past}
	s.False(u.IsLocked())
}

func (s *UserSuite) TestIsLocked_FutureLockout() {
	future := time.Now().Add(time.Hour)
	u := &domain.User{LockoutUntil: &future}
	s.True(u.IsLocked())
}

func (s *UserSuite) TestCanAuthenticate_Happy() {
	u := &domain.User{
		IsActive:      true,
		AccountStatus: domain.AccountStatusActive,
		DeletedAt:     nil,
		LockoutUntil:  nil,
	}
	s.True(u.CanAuthenticate())
}

func (s *UserSuite) TestCanAuthenticate_Inactive() {
	u := &domain.User{IsActive: false, AccountStatus: domain.AccountStatusActive}
	s.False(u.CanAuthenticate())
}

func (s *UserSuite) TestCanAuthenticate_WrongStatus() {
	u := &domain.User{IsActive: true, AccountStatus: domain.AccountStatusSuspended}
	s.False(u.CanAuthenticate())
}

func (s *UserSuite) TestCanAuthenticate_SoftDeleted() {
	now := time.Now()
	u := &domain.User{IsActive: true, AccountStatus: domain.AccountStatusActive, DeletedAt: &now}
	s.False(u.CanAuthenticate())
}

func (s *UserSuite) TestCanAuthenticate_Locked() {
	future := time.Now().Add(time.Hour)
	u := &domain.User{IsActive: true, AccountStatus: domain.AccountStatusActive, LockoutUntil: &future}
	s.False(u.CanAuthenticate())
}

// =============================================================================
// AccountStatus Suite
// =============================================================================

type AccountStatusSuite struct{ suite.Suite }

func TestAccountStatusSuite(t *testing.T) { suite.Run(t, new(AccountStatusSuite)) }

func (s *AccountStatusSuite) TestAllValid() {
	for _, st := range domain.AllAccountStatuses() {
		s.True(st.IsValid(), "status=%q", st)
	}
}

func (s *AccountStatusSuite) TestInvalid() {
	s.False(domain.AccountStatus("BANNED").IsValid())
	s.False(domain.AccountStatus("").IsValid())
}

// =============================================================================
// EmploymentStatus Suite
// =============================================================================

type EmploymentStatusSuite struct{ suite.Suite }

func TestEmploymentStatusSuite(t *testing.T) { suite.Run(t, new(EmploymentStatusSuite)) }

func (s *EmploymentStatusSuite) TestAllValid() {
	for _, st := range domain.AllEmploymentStatuses() {
		s.True(st.IsValid(), "status=%q", st)
	}
}

func (s *EmploymentStatusSuite) TestInvalid() {
	s.False(domain.EmploymentStatus("FIRED").IsValid())
	s.False(domain.EmploymentStatus("").IsValid())
}

// =============================================================================
// Person Suite
// =============================================================================

type PersonSuite struct{ suite.Suite }

func TestPersonSuite(t *testing.T) { suite.Run(t, new(PersonSuite)) }

func (s *PersonSuite) TestGetFullName_NoMiddle() {
	p := &domain.Person{FirstName: "Alice", LastName: "Smith"}
	s.Equal("Alice Smith", p.GetFullName())
}

func (s *PersonSuite) TestGetFullName_WithMiddle() {
	mid := "Marie"
	p := &domain.Person{FirstName: "Alice", MiddleName: &mid, LastName: "Smith"}
	s.Equal("Alice Marie Smith", p.GetFullName())
}

func (s *PersonSuite) TestGetFullName_EmptyMiddle() {
	empty := ""
	p := &domain.Person{FirstName: "Bob", MiddleName: &empty, LastName: "Jones"}
	s.Equal("Bob Jones", p.GetFullName())
}

// =============================================================================
// Employee Suite
// =============================================================================

type EmployeeSuite struct{ suite.Suite }

func TestEmployeeSuite(t *testing.T) { suite.Run(t, new(EmployeeSuite)) }

func (s *EmployeeSuite) TestIsCurrentlyEmployed_Active() {
	e := &domain.Employee{Status: domain.EmploymentStatusActive}
	s.True(e.IsCurrentlyEmployed())
}

func (s *EmployeeSuite) TestIsCurrentlyEmployed_OnLeave() {
	e := &domain.Employee{Status: domain.EmploymentStatusOnLeave}
	s.True(e.IsCurrentlyEmployed())
}

func (s *EmployeeSuite) TestIsCurrentlyEmployed_Terminated() {
	e := &domain.Employee{Status: domain.EmploymentStatusTerminated}
	s.False(e.IsCurrentlyEmployed())
}

func (s *EmployeeSuite) TestIsCurrentlyEmployed_SoftDeleted() {
	now := time.Now()
	e := &domain.Employee{Status: domain.EmploymentStatusActive, DeletedAt: &now}
	s.False(e.IsCurrentlyEmployed())
}

// =============================================================================
// CreateUserRequest Validate Suite
// =============================================================================

type CreateUserRequestSuite struct{ suite.Suite }

func TestCreateUserRequestSuite(t *testing.T) { suite.Run(t, new(CreateUserRequestSuite)) }

func validCreateUserReq() *domain.CreateUserRequest {
	return &domain.CreateUserRequest{
		EntityID: uuid.New(),
		Username: "alice",
		Email:    "alice@example.com",
		Password: "securepassword",
		UserType: "INTERNAL",
	}
}

func (s *CreateUserRequestSuite) TestValidate_Happy() {
	s.Require().NoError(validCreateUserReq().Validate())
}

func (s *CreateUserRequestSuite) TestValidate_MissingEntityID() {
	req := validCreateUserReq()
	req.EntityID = uuid.Nil
	s.Require().Error(req.Validate())
}

func (s *CreateUserRequestSuite) TestValidate_MissingUsername() {
	req := validCreateUserReq()
	req.Username = ""
	s.Require().Error(req.Validate())
}

func (s *CreateUserRequestSuite) TestValidate_InvalidEmail() {
	req := validCreateUserReq()
	req.Email = "not-an-email"
	s.Require().Error(req.Validate())
}

func (s *CreateUserRequestSuite) TestValidate_ShortPassword() {
	req := validCreateUserReq()
	req.Password = "short"
	s.Require().Error(req.Validate())
}

func (s *CreateUserRequestSuite) TestValidate_MissingUserType() {
	req := validCreateUserReq()
	req.UserType = ""
	s.Require().Error(req.Validate())
}

// =============================================================================
// PasswordResetToken Suite
// =============================================================================

type PasswordResetTokenSuite struct{ suite.Suite }

func TestPasswordResetTokenSuite(t *testing.T) { suite.Run(t, new(PasswordResetTokenSuite)) }

func (s *PasswordResetTokenSuite) TestIsExpired_Future() {
	t := &domain.PasswordResetToken{ExpiresAt: time.Now().Add(time.Hour)}
	s.False(t.IsExpired())
}

func (s *PasswordResetTokenSuite) TestIsExpired_Past() {
	t := &domain.PasswordResetToken{ExpiresAt: time.Now().Add(-time.Minute)}
	s.True(t.IsExpired())
}

func (s *PasswordResetTokenSuite) TestIsUsed_Nil() {
	t := &domain.PasswordResetToken{UsedAt: nil}
	s.False(t.IsUsed())
}

func (s *PasswordResetTokenSuite) TestIsUsed_Set() {
	now := time.Now()
	t := &domain.PasswordResetToken{UsedAt: &now}
	s.True(t.IsUsed())
}

// =============================================================================
// Errors Suite
// =============================================================================

type ErrorsSuite struct{ suite.Suite }

func TestErrorsSuite(t *testing.T) { suite.Run(t, new(ErrorsSuite)) }

func (s *ErrorsSuite) TestSentinelErrorMessages() {
	s.Contains(domain.ErrForbidden.Error(), "AUTHZ_FORBIDDEN")
	s.Contains(domain.ErrUnauthorized.Error(), "AUTHZ_UNAUTHORIZED")
	s.Contains(domain.ErrInvalidRequest.Error(), "AUTHZ_INVALID")
	s.Contains(domain.ErrPolicyConflict.Error(), "AUTHZ_DUPLICATE")
	s.Contains(domain.ErrPolicyLimitExceeded.Error(), "AUTHZ_POLICY_LIMIT")
}

func (s *ErrorsSuite) TestErrInvalidIdentity() {
	err := domain.ErrInvalidIdentity("username is required")
	s.Require().NotNil(err)
	s.Equal(400, err.HTTPStatus)
	s.Contains(err.Error(), "IDENTITY_INVALID")
	s.Contains(err.Error(), "username is required")
}

// =============================================================================
// MarshalSessionJSON Suite
// =============================================================================

type MarshalSessionJSONSuite struct{ suite.Suite }

func TestMarshalSessionJSONSuite(t *testing.T) { suite.Run(t, new(MarshalSessionJSONSuite)) }

func (s *MarshalSessionJSONSuite) TestMarshalValid() {
	scope := domain.EntityScope{Type: domain.EntityScopeAll}
	b := domain.MarshalSessionJSON(scope)
	s.Contains(string(b), "all")
}

func (s *MarshalSessionJSONSuite) TestMarshalFallback() {
	// channels are not JSON-serialisable; should return "{}"
	b := domain.MarshalSessionJSON(make(chan int))
	s.Equal("{}", string(b))
}

// =============================================================================
// Configuration Suite
// =============================================================================

type ConfigurationSuite struct{ suite.Suite }

func TestConfigurationSuite(t *testing.T) { suite.Run(t, new(ConfigurationSuite)) }

func (s *ConfigurationSuite) TestDefaultConfiguration() {
	cfg := domain.DefaultConfiguration()
	s.NotNil(cfg.Flags)
	s.NotNil(cfg.Settings)
	s.NotNil(cfg.Prefs)
}

func (s *ConfigurationSuite) TestDefaultSessionConfig() {
	cfg := domain.DefaultSessionConfig()
	s.Equal(8*time.Hour, cfg.SessionTTL)
	s.Equal("session", cfg.CookieName)
}
