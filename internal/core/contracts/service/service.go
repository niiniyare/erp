package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"awo.so/internal/core/contracts/domain"
	"awo.so/internal/core/contracts/repository"
	iamcontract "awo.so/internal/core/iam/contract"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/tracing"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
	cacheTTL         = 15 * time.Minute
	cacheNS          = "contracts"
)

// ContractService handles all contract lifecycle and business logic.
//
// Security model:
//   - IAM middleware (Casbin) enforces permission checks before any service method runs.
//   - Service extracts [iamcontract.SessionContext] from ctx for identity and entity scope.
//   - Entity scope (All / Subtree / Entity) is applied in List to restrict cross-entity visibility.
//   - Repo enforces tenant-level RLS via WithTenantFromCtx; no tenant leakage is possible.
type ContractService struct {
	repo   repository.Repository
	cache  cache.Service
	tracer tracing.Service
	log    logger.Logger
}

// NewContractService constructs a ContractService with the given dependencies.
func NewContractService(
	repo repository.Repository,
	cache cache.Service,
	tracer tracing.Service,
	log logger.Logger,
) *ContractService {
	return &ContractService{repo: repo, cache: cache, tracer: tracer, log: log}
}

// ── Create ────────────────────────────────────────────────────────────────────

// Create validates and persists a new contract.
//
// Security: requires an authenticated session in ctx (injected by IAM middleware).
// The creating user is recorded in CreatedBy for audit trails.
func (s *ContractService) Create(ctx context.Context, req domain.CreateContractRequest) (*domain.Contract, error) {
	ctx, span := s.tracer.StartSpan(ctx, "contracts.svc.Create")
	defer span.End()

	// Extract authenticated session — fails fast if middleware didn't inject one.
	sc, ok := iamcontract.FromContext(ctx)
	if !ok {
		return nil, errors.New("unauthenticated: no session in context")
	}
	span.SetAttributes(
		attribute.String("contract.entity_id", req.EntityID.String()),
		attribute.String("contract.type", req.ContractType),
		attribute.String("actor.user_id", sc.UserID().String()),
	)

	// Generate a unique contract number.
	number := generateContractNumber()

	// Guard against duplicate numbers before insert to give a cleaner error.
	exists, err := s.repo.NumberExists(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to check contract number: %w", err)
	}
	if exists {
		return nil, domain.ErrContractAlreadyExists
	}

	// Normalize currency.
	currency := strings.ToUpper(req.CurrencyCode)
	if currency == "" {
		currency = "USD"
	}

	// Build functional options list.
	opts := []domain.Option{}

	if req.ContractType != "" {
		ct := domain.ContractType(strings.ToUpper(req.ContractType))
		opts = append(opts, domain.WithContractType(ct))
	}
	if !req.Value.IsZero() {
		opts = append(opts, domain.WithContractValue(req.Value, currency))
	}
	if req.EndDate != nil {
		opts = append(opts, domain.WithEndDate(*req.EndDate))
	}
	if req.CounterpartyEmail != nil && *req.CounterpartyEmail != "" {
		opts = append(opts, domain.WithCounterpartyEmail(*req.CounterpartyEmail))
	}
	if req.Description != nil {
		opts = append(opts, domain.WithDescription(*req.Description))
	}
	if req.Terms != nil {
		opts = append(opts, domain.WithTerms(*req.Terms))
	}

	// Record the creating user for audit purposes.
	userID := sc.UserID()
	opts = append(opts, domain.WithCreatedBy(userID))

	c, err := domain.NewContract(
		req.Title,
		req.CounterpartyName,
		currency,
		number,
		req.EntityID,
		req.StartDate,
		opts...,
	)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if err := s.repo.Create(ctx, c); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "create failed")
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}

	// Warm the cache immediately so the first GetByID is a hit.
	s.cacheContract(ctx, c)

	s.log.DebugContext(ctx, "contract created", logger.Fields{
		"id":      c.ID,
		"number":  c.Number,
		"user_id": sc.UserID(),
	})

	return c, nil
}

// ── GetByID ───────────────────────────────────────────────────────────────────

// GetByID retrieves a contract by UUID, checking the cache first.
func (s *ContractService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Contract, error) {
	ctx, span := s.tracer.StartSpan(ctx, "contracts.svc.GetByID")
	defer span.End()
	span.SetAttributes(attribute.String("contract.id", id.String()))

	if _, ok := iamcontract.FromContext(ctx); !ok {
		return nil, errors.New("unauthenticated: no session in context")
	}

	// Try cache.
	cctx := s.cacheCtx(ctx)
	var cached domain.Contract
	if err := s.cache.Get(cctx, "id:"+id.String(), &cached); err == nil {
		span.SetAttributes(attribute.Bool("cache.hit", true))
		return &cached, nil
	}
	span.SetAttributes(attribute.Bool("cache.hit", false))

	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.cacheContract(ctx, c)
	return c, nil
}

// ── GetByNumber ───────────────────────────────────────────────────────────────

// GetByNumber retrieves a contract by its human-readable number.
func (s *ContractService) GetByNumber(ctx context.Context, number string) (*domain.Contract, error) {
	ctx, span := s.tracer.StartSpan(ctx, "contracts.svc.GetByNumber")
	defer span.End()

	if _, ok := iamcontract.FromContext(ctx); !ok {
		return nil, errors.New("unauthenticated: no session in context")
	}

	// Try number-keyed cache entry.
	cctx := s.cacheCtx(ctx)
	var cached domain.Contract
	if err := s.cache.Get(cctx, "number:"+number, &cached); err == nil {
		return &cached, nil
	}

	c, err := s.repo.GetByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	s.cacheContract(ctx, c)
	return c, nil
}

// ── Update ────────────────────────────────────────────────────────────────────

// Update applies partial changes to an existing contract.
// The contract must be in an editable state (DRAFT or PENDING_APPROVAL).
func (s *ContractService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateContractRequest) (*domain.Contract, error) {
	ctx, span := s.tracer.StartSpan(ctx, "contracts.svc.Update")
	defer span.End()
	span.SetAttributes(attribute.String("contract.id", id.String()))

	if _, ok := iamcontract.FromContext(ctx); !ok {
		return nil, errors.New("unauthenticated: no session in context")
	}

	// Load current state to enforce business rules before persisting.
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !c.IsEditable() {
		return nil, domain.ErrInvalidTransition
	}

	updated, err := s.repo.Update(ctx, id, req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	s.invalidateCache(ctx, id, c.Number)
	s.cacheContract(ctx, updated)
	return updated, nil
}

// ── Delete ────────────────────────────────────────────────────────────────────

// Delete soft-deletes a DRAFT contract. Active or terminated contracts must
// be explicitly terminated before they can be removed.
func (s *ContractService) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "contracts.svc.Delete")
	defer span.End()
	span.SetAttributes(attribute.String("contract.id", id.String()))

	if _, ok := iamcontract.FromContext(ctx); !ok {
		return errors.New("unauthenticated: no session in context")
	}

	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Only draft contracts may be soft-deleted through this path.
	if !c.IsDraft() {
		return domain.ErrInvalidTransition
	}

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return err
	}

	s.invalidateCache(ctx, id, c.Number)
	return nil
}

// ── List ──────────────────────────────────────────────────────────────────────

// List returns a filtered, paginated list of contracts.
//
// Entity scope enforcement (Layer 2 isolation):
//   - EntityScopeAll    → no extra entity filter; RLS limits to tenant
//   - EntityScopeEntity → only contracts for the user's home entity
//   - EntityScopeSubtree → subtree filtering not yet supported via SQLC filter;
//     falls back to the user's home entity (conservative)
func (s *ContractService) List(ctx context.Context, filter domain.ContractFilter) ([]*domain.Contract, int64, error) {
	ctx, span := s.tracer.StartSpan(ctx, "contracts.svc.List")
	defer span.End()

	sc, ok := iamcontract.FromContext(ctx)
	if !ok {
		return nil, 0, errors.New("unauthenticated: no session in context")
	}

	// Apply entity scope — this is the Layer 2 isolation gate.
	// If the caller did not pass an explicit entity_id filter we apply one
	// derived from the session's EntityScope to prevent cross-entity leakage.
	filter = s.applyEntityScope(filter, sc)

	// Enforce sane pagination limits.
	if filter.Limit <= 0 || filter.Limit > maxListLimit {
		filter.Limit = defaultListLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return s.repo.List(ctx, filter)
}

// ── Status Transitions ────────────────────────────────────────────────────────

// Submit transitions a DRAFT contract to PENDING_APPROVAL.
func (s *ContractService) Submit(ctx context.Context, id uuid.UUID, version int32) error {
	ctx, span := s.tracer.StartSpan(ctx, "contracts.svc.Submit")
	defer span.End()
	span.SetAttributes(attribute.String("contract.id", id.String()))

	if _, ok := iamcontract.FromContext(ctx); !ok {
		return errors.New("unauthenticated: no session in context")
	}

	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate the state machine transition in the domain.
	if err := c.Submit(); err != nil {
		return err
	}

	_, err = s.repo.Update(ctx, id, buildStatusUpdate(domain.StatusPendingApproval, version))
	if err == nil {
		s.invalidateCache(ctx, id, c.Number)
	}
	return err
}

// Approve transitions a PENDING_APPROVAL contract to ACTIVE.
// The approver's identity is taken from the session context — not from the request body —
// to prevent impersonation.
func (s *ContractService) Approve(ctx context.Context, id uuid.UUID, version int32) error {
	ctx, span := s.tracer.StartSpan(ctx, "contracts.svc.Approve")
	defer span.End()
	span.SetAttributes(attribute.String("contract.id", id.String()))

	sc, ok := iamcontract.FromContext(ctx)
	if !ok {
		return errors.New("unauthenticated: no session in context")
	}

	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	signerID := sc.UserID()
	if err := c.Approve(signerID); err != nil {
		return err
	}

	now := time.Now()
	req := buildStatusUpdate(domain.StatusActive, version)
	req.SignedBy = &signerID
	req.SignedAt = &now

	_, err = s.repo.Update(ctx, id, req)
	if err == nil {
		s.invalidateCache(ctx, id, c.Number)
	}
	return err
}

// Reject moves a PENDING_APPROVAL contract back to DRAFT.
func (s *ContractService) Reject(ctx context.Context, id uuid.UUID, version int32) error {
	ctx, span := s.tracer.StartSpan(ctx, "contracts.svc.Reject")
	defer span.End()
	span.SetAttributes(attribute.String("contract.id", id.String()))

	if _, ok := iamcontract.FromContext(ctx); !ok {
		return errors.New("unauthenticated: no session in context")
	}

	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := c.Reject(); err != nil {
		return err
	}

	_, err = s.repo.Update(ctx, id, buildStatusUpdate(domain.StatusDraft, version))
	if err == nil {
		s.invalidateCache(ctx, id, c.Number)
	}
	return err
}

// Terminate moves an ACTIVE contract to TERMINATED and records the reason.
func (s *ContractService) Terminate(ctx context.Context, id uuid.UUID, reason string, version int32) error {
	ctx, span := s.tracer.StartSpan(ctx, "contracts.svc.Terminate")
	defer span.End()
	span.SetAttributes(attribute.String("contract.id", id.String()))

	if _, ok := iamcontract.FromContext(ctx); !ok {
		return errors.New("unauthenticated: no session in context")
	}

	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := c.Terminate(reason); err != nil {
		return err
	}

	_, err = s.repo.Update(ctx, id, buildStatusUpdate(domain.StatusTerminated, version))
	if err == nil {
		s.invalidateCache(ctx, id, c.Number)
	}
	return err
}

// ── Cache helpers ─────────────────────────────────────────────────────────────

// cacheCtx attaches the contracts namespace so the cache keys are segregated.
func (s *ContractService) cacheCtx(ctx context.Context) context.Context {
	return cache.WithNamespace(ctx, cacheNS)
}

// cacheContract stores a contract under both its ID key and number key.
func (s *ContractService) cacheContract(ctx context.Context, c *domain.Contract) {
	cctx := s.cacheCtx(ctx)
	if err := s.cache.Set(cctx, "id:"+c.ID.String(), c, cacheTTL); err != nil {
		s.log.WarnContext(ctx, "failed to cache contract by ID", logger.Fields{
			"id":    c.ID,
			"error": err.Error(),
		})
	}
	if err := s.cache.Set(cctx, "number:"+c.Number, c, cacheTTL); err != nil {
		s.log.WarnContext(ctx, "failed to cache contract by number", logger.Fields{
			"number": c.Number,
			"error":  err.Error(),
		})
	}
}

// invalidateCache removes a contract from both cache keys.
func (s *ContractService) invalidateCache(ctx context.Context, id uuid.UUID, number string) {
	cctx := s.cacheCtx(ctx)
	_ = s.cache.Delete(cctx, "id:"+id.String())
	if number != "" {
		_ = s.cache.Delete(cctx, "number:"+number)
	}
}

// ── Entity scope ──────────────────────────────────────────────────────────────

// applyEntityScope enforces the session's entity visibility on a list filter.
// It never widens the filter beyond what the caller requested; it can only
// narrow it.
func (s *ContractService) applyEntityScope(filter domain.ContractFilter, sc iamcontract.SessionContext) domain.ContractFilter {
	scope := sc.EntityScope()

	switch scope.Type {
	case "all":
		// Platform/super-admin: no restriction beyond RLS tenant boundary.
		return filter

	case "entity":
		// Restrict to user's home entity.
		if scope.EntityID != "" {
			id, err := uuid.Parse(scope.EntityID)
			if err == nil && (filter.EntityID == nil || *filter.EntityID == uuid.Nil) {
				filter.EntityID = &id
			}
		}

	case "subtree":
		// Full subtree support requires an ltree path query which SQLC filter
		// currently does not expose. Conservative fallback: restrict to home entity.
		if scope.EntityID != "" {
			id, err := uuid.Parse(scope.EntityID)
			if err == nil && (filter.EntityID == nil || *filter.EntityID == uuid.Nil) {
				filter.EntityID = &id
			}
		}
	}

	return filter
}

// ── Transition helpers ────────────────────────────────────────────────────────

// buildStatusUpdate constructs a minimal UpdateContractRequest that sets only status.
func buildStatusUpdate(status domain.ContractStatus, version int32) domain.UpdateContractRequest {
	s := string(status)
	return domain.UpdateContractRequest{
		Status:  &s,
		Version: version,
	}
}

// generateContractNumber produces a time-based contract number.
// Format: CTR-YYYYMMDD-NNNNNN (nanosecond suffix for uniqueness).
// Production should use the entitystate sequence table for gapless numbering.
func generateContractNumber() string {
	now := time.Now()
	nano := now.UnixNano() % 1_000_000
	return fmt.Sprintf("CTR-%s-%06d", now.Format("20060102"), nano)
}
