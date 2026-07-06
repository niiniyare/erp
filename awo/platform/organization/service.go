package organization

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// OrganizationService is the domain service for organization lifecycle, tree
// operations, and scope resolution. All persistence goes through EntityRepository
// — no raw SQL.
//
// # Scope resolution
//
// ResolveScope is the central method for application-layer org authorization.
// It converts a ViewerContext into a concrete []uuid.UUID that application
// services pass as an IN predicate to their repositories. Repositories remain
// org-unaware — they only receive filter criteria.
//
// Request evaluation order:
//
//	HTTP Request
//	  ↓ auth middleware (session validation)
//	ViewerContext (loaded with org assignments)
//	  ↓ service middleware or application service
//	OrganizationService.ResolveScope()
//	  ↓ returns []uuid.UUID (visible org IDs)
//	Application service builds filter
//	  ↓
//	Repository.List(ctx, filter)
type OrganizationService interface {
	// ── Tree operations ──────────────────────────────────────────────────────

	// Create adds a new organization node. ParentID == uuid.Nil → root node.
	Create(ctx context.Context, input CreateInput) (*OrganizationDTO, error)

	// GetByID returns an organization node by UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*OrganizationDTO, error)

	// GetByCode returns an organization node by its immutable code within a tenant.
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*OrganizationDTO, error)

	// Move re-parents id under newParentID, recomputing path+depth for the
	// entire subtree atomically. Returns error if move would create a cycle.
	// newParentID == uuid.Nil promotes id to root.
	Move(ctx context.Context, id, newParentID uuid.UUID) error

	// Disable marks id and all its descendants as active=false.
	Disable(ctx context.Context, id uuid.UUID) error

	// Enable re-activates id. Does not re-enable descendants — call Enable on
	// each descendant explicitly if needed.
	Enable(ctx context.Context, id uuid.UUID) error

	// Tree returns all organization nodes for a tenant in depth-first order.
	Tree(ctx context.Context, tenantID uuid.UUID) ([]*OrganizationDTO, error)

	// Ancestors returns the path from root to id (root first, target last).
	// Efficient: uses materialized path string, no recursive CTE.
	Ancestors(ctx context.Context, id uuid.UUID) ([]*OrganizationDTO, error)

	// Descendants returns all nodes below id. maxDepth=0 means unlimited.
	// Efficient: uses materialized path prefix query.
	Descendants(ctx context.Context, id uuid.UUID, maxDepth int) ([]*OrganizationDTO, error)

	// ── Scope resolution ─────────────────────────────────────────────────────

	// ResolveScope converts a ViewerContext into the set of org IDs visible to
	// that viewer. Application services pass the result as an explicit IN filter
	// to their repositories.
	//
	// Returns nil when VisibilityEntireTenant is in effect — caller should omit
	// the org filter entirely (all orgs in the tenant are accessible).
	//
	// Behaviour by VisibilityMode:
	//   Current       → [viewer.EffectiveOrganizationID()]
	//   Descendants   → viewer's org + all descendants via path prefix
	//   Ancestors     → viewer's org + all ancestors via path parsing
	//   EntireTenant  → nil (no filter)
	//   Explicit      → viewer.ExplicitOrganizationIDs
	//   Custom        → viewer.ScopeResolver.Resolve(ctx, viewer)
	ResolveScope(ctx context.Context, viewer ViewerContext) ([]uuid.UUID, error)

	// ── Type registry ────────────────────────────────────────────────────────

	// RegisterType adds a new org type to the tenant's type registry.
	RegisterType(ctx context.Context, tenantID uuid.UUID, input OrgTypeInput) (*OrgTypeDTO, error)

	// ListTypes returns active org types for a tenant, ordered by sort_order.
	ListTypes(ctx context.Context, tenantID uuid.UUID) ([]*OrgTypeDTO, error)

	// DeactivateType marks a type as inactive. Does not modify existing org
	// nodes that carry the type — type value on nodes is a denormalized string.
	DeactivateType(ctx context.Context, id uuid.UUID) error

	// ── Assignments ──────────────────────────────────────────────────────────

	// Assign adds userID to orgID with the given role within tenantID.
	// If setPrimary=true, any existing primary assignment for the user is cleared.
	Assign(ctx context.Context, input AssignInput) (*AssignmentDTO, error)

	// Unassign removes a user from an organization. If the removed assignment
	// was the primary, no new primary is automatically selected — caller must
	// call SetPrimary on another assignment if needed.
	Unassign(ctx context.Context, tenantID, userID, orgID uuid.UUID) error

	// SetPrimary marks the given orgID as the user's primary organization.
	// Clears any existing primary for that user in the tenant.
	SetPrimary(ctx context.Context, tenantID, userID, orgID uuid.UUID) error

	// ListAssignments returns all org memberships for a user within a tenant.
	// Includes inactive orgs — callers filter if needed.
	ListAssignments(ctx context.Context, tenantID, userID uuid.UUID) ([]*AssignmentDTO, error)

	// LoadViewerMemberships returns the OrgMembership slice needed to populate
	// ViewerContext.OrganizationAssignments. Called once per session by auth
	// middleware after session validation.
	LoadViewerMemberships(ctx context.Context, tenantID, userID uuid.UUID) ([]OrgMembership, error)
}

// ── Input / DTO types ─────────────────────────────────────────────────────────

// CreateInput holds the fields needed to create a new organization node.
type CreateInput struct {
	TenantID    uuid.UUID
	Name        string
	Code        string
	Type        string    // references platform_org_type.name; not validated by framework
	Description string
	ParentID    uuid.UUID // uuid.Nil = root node
	Active      bool
}

// OrganizationDTO is the read model returned by tree operation methods.
type OrganizationDTO struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	Name     string
	Code     string
	Type     string
	Path     string
	Depth    int
	ParentID uuid.UUID // uuid.Nil for root nodes
	Active   bool
	Description string
}

// OrgTypeInput holds the fields needed to register a new org type.
type OrgTypeInput struct {
	Name        string
	Label       string
	Description string
	SortOrder   int
}

// OrgTypeDTO is the read model for a registered org type.
type OrgTypeDTO struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Label       string
	Description string
	SortOrder   int
	Active      bool
}

// AssignInput holds the fields needed to add a user to an organization.
type AssignInput struct {
	TenantID       uuid.UUID
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Role           string // org-level role: manager, member, viewer, …
	SetPrimary     bool
}

// AssignmentDTO is the read model for a single org membership record.
type AssignmentDTO struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Role           string
	IsPrimary      bool
}

// ── Stub implementation ───────────────────────────────────────────────────────

type stubOrganizationService struct{}

// NewOrganizationService returns a stub OrganizationService.
// TODO: inject EntityRepository[platform_organization] and related repos.
func NewOrganizationService() OrganizationService {
	return &stubOrganizationService{}
}

func (s *stubOrganizationService) Create(_ context.Context, _ CreateInput) (*OrganizationDTO, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.Create")
}

func (s *stubOrganizationService) GetByID(_ context.Context, _ uuid.UUID) (*OrganizationDTO, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.GetByID")
}

func (s *stubOrganizationService) GetByCode(_ context.Context, _ uuid.UUID, _ string) (*OrganizationDTO, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.GetByCode")
}

func (s *stubOrganizationService) Move(_ context.Context, _, _ uuid.UUID) error {
	return fmt.Errorf("not implemented: OrganizationService.Move")
}

func (s *stubOrganizationService) Disable(_ context.Context, _ uuid.UUID) error {
	return fmt.Errorf("not implemented: OrganizationService.Disable")
}

func (s *stubOrganizationService) Enable(_ context.Context, _ uuid.UUID) error {
	return fmt.Errorf("not implemented: OrganizationService.Enable")
}

func (s *stubOrganizationService) Tree(_ context.Context, _ uuid.UUID) ([]*OrganizationDTO, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.Tree")
}

func (s *stubOrganizationService) Ancestors(_ context.Context, _ uuid.UUID) ([]*OrganizationDTO, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.Ancestors")
}

func (s *stubOrganizationService) Descendants(_ context.Context, _ uuid.UUID, _ int) ([]*OrganizationDTO, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.Descendants")
}

func (s *stubOrganizationService) ResolveScope(_ context.Context, _ ViewerContext) ([]uuid.UUID, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.ResolveScope")
}

func (s *stubOrganizationService) RegisterType(_ context.Context, _ uuid.UUID, _ OrgTypeInput) (*OrgTypeDTO, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.RegisterType")
}

func (s *stubOrganizationService) ListTypes(_ context.Context, _ uuid.UUID) ([]*OrgTypeDTO, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.ListTypes")
}

func (s *stubOrganizationService) DeactivateType(_ context.Context, _ uuid.UUID) error {
	return fmt.Errorf("not implemented: OrganizationService.DeactivateType")
}

func (s *stubOrganizationService) Assign(_ context.Context, _ AssignInput) (*AssignmentDTO, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.Assign")
}

func (s *stubOrganizationService) Unassign(_ context.Context, _, _, _ uuid.UUID) error {
	return fmt.Errorf("not implemented: OrganizationService.Unassign")
}

func (s *stubOrganizationService) SetPrimary(_ context.Context, _, _, _ uuid.UUID) error {
	return fmt.Errorf("not implemented: OrganizationService.SetPrimary")
}

func (s *stubOrganizationService) ListAssignments(_ context.Context, _, _ uuid.UUID) ([]*AssignmentDTO, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.ListAssignments")
}

func (s *stubOrganizationService) LoadViewerMemberships(_ context.Context, _, _ uuid.UUID) ([]OrgMembership, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.LoadViewerMemberships")
}
