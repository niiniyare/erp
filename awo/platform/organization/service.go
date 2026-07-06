package organization

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// OrganizationService is the domain service for organization lifecycle and
// tree operations. All persistence goes through EntityRepository — no raw SQL.
type OrganizationService interface {
	// Create adds a new organization node. If input.ParentID is uuid.Nil the
	// node becomes a root (top-level) node for the tenant.
	Create(ctx context.Context, input CreateInput) (*OrganizationDTO, error)

	// GetByID returns an organization node by UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*OrganizationDTO, error)

	// GetByCode returns an organization node by its immutable code.
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*OrganizationDTO, error)

	// Move re-parents a node to a new parent, recomputing path and depth for the
	// entire subtree atomically. Rejects moves that would create cycles.
	Move(ctx context.Context, id, newParentID uuid.UUID) error

	// Disable marks a node and all its descendants as active=false.
	Disable(ctx context.Context, id uuid.UUID) error

	// Enable re-activates a node (ancestors must be active for this to be meaningful).
	Enable(ctx context.Context, id uuid.UUID) error

	// Tree returns all organization nodes for a tenant in depth-first order.
	// Use ScopeFromContext to restrict the visible subtree.
	Tree(ctx context.Context, tenantID uuid.UUID) ([]*OrganizationDTO, error)

	// Ancestors returns the ordered path from the root to the given node
	// (root first, target last). Efficient: uses materialized path.
	Ancestors(ctx context.Context, id uuid.UUID) ([]*OrganizationDTO, error)

	// Descendants returns all nodes below the given node in the tree.
	// Efficient: uses materialized path prefix query.
	Descendants(ctx context.Context, id uuid.UUID, depth int) ([]*OrganizationDTO, error)

	// Resolve applies an OrganizationScope to return the visible org IDs for a
	// given caller. Used by middleware to inject ScopeContext.
	Resolve(ctx context.Context, tenantID, callerOrgID uuid.UUID, scope OrganizationScope) ([]uuid.UUID, error)
}

// CreateInput holds the fields needed to create a new organization node.
type CreateInput struct {
	TenantID    uuid.UUID
	Name        string
	Code        string
	Type        string // metadata-driven; not validated against a fixed enum
	Description string
	ParentID    uuid.UUID // uuid.Nil = root node
	Active      bool
}

// OrganizationDTO is the read model returned by OrganizationService methods.
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
}

// stubOrganizationService is the stub implementation.
// Replace with a real implementation injecting EntityRepository[platform_organization].
type stubOrganizationService struct{}

// NewOrganizationService returns a stub OrganizationService.
// TODO: inject EntityRepository[platform_organization] when implementing for real.
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

func (s *stubOrganizationService) Resolve(_ context.Context, _, _ uuid.UUID, _ OrganizationScope) ([]uuid.UUID, error) {
	return nil, fmt.Errorf("not implemented: OrganizationService.Resolve")
}
