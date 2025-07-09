package organization

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Repository defines the interface for organization data access
type Repository interface {
	CreateEntity(ctx context.Context, arg db.CreateEntityParams) (*Organization, error)
	GetEntity(ctx context.Context, id uuid.UUID) (*Organization, error)
	GetEntityByCode(ctx context.Context, code string) (*Organization, error)
	GetEntityByName(ctx context.Context, name string) (*Organization, error)
	GetEntityWithHierarchy(ctx context.Context, id uuid.UUID) (*db.GetEntityWithHierarchyInfoRow, error)
	ListEntities(ctx context.Context, arg db.ListEntitiesWithPaginationParams) ([]*Organization, error)
	UpdateEntity(ctx context.Context, arg db.UpdateEntityParams) (*Organization, error)
	DeleteEntity(ctx context.Context, id uuid.UUID, permanent bool) error
	RestoreEntity(ctx context.Context, id uuid.UUID) error
	GetNextSequence(ctx context.Context, arg db.GetNextSequenceNumberParams) (int32, error)
	ResetEntitySequence(ctx context.Context, arg db.ResetEntityStateSequenceParams) error
	GetEntityChildren(ctx context.Context, parentId uuid.UUID) ([]*Organization, error)
	GetEntityAncestors(ctx context.Context, id uuid.UUID) ([]db.GetEntityAncestorsRow, error)
	GetEntityTree(ctx context.Context, tenantId uuid.UUID) ([]db.GetEntityTreeStructureRow, error)
	ValidateEntityName(ctx context.Context, name string, id uuid.UUID) (bool, error)
	ValidateEntityCode(ctx context.Context, code string, id uuid.UUID) (bool, error)
	ValidateEntityParent(ctx context.Context, parentId, childId uuid.UUID) (bool, error)
	CheckCircularReference(ctx context.Context, parentId, childId uuid.UUID) (bool, error)
}

// repository implements Repository interface
type repository struct {
	store   db.Store
	logger  logger.Logger
	tracing tracing.TracingService
	metrics metrics.MetricsProvider
}

// NewRepository creates a new organization repository
func NewRepository(store db.Store, logger logger.Logger, tracing tracing.TracingService, metrics metrics.MetricsProvider) Repository {
	return &repository{
		store:   store,
		logger:  logger.WithFields(logger.Fields{"module": "repository", "package": "organization"}),
		tracing: tracing,
		metrics: metrics,
	}
}

func (r *repository) instrument(ctx context.Context, operation string, f func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository."+operation,
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", operation),
			attribute.String("db.table", "entities"),
		))
	defer span.End()

	r.logger.DebugContext(ctx, "db call", logger.Fields{"operation": operation})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{"operation": operation, "table": "entities"})
	defer timer.Stop()

	res, err := f(ctx)
	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{"operation": operation, "error_type": "sql_error"})
		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")
		r.logger.ErrorContext(ctx, "db call failed", logger.Fields{"error": err, "operation": operation})
		return nil, err
	}

	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{"operation": operation, "status": "success"})
	return res, nil
}

func (r *repository) CreateEntity(ctx context.Context, arg db.CreateEntityParams) (*Organization, error) {
	res, err := r.instrument(ctx, "CreateEntity", func(ctx context.Context) (interface{}, error) {
		entity, err := r.store.CreateEntity(ctx, arg)
		if err != nil {
			return nil, err
		}

		if arg.ParentID != nil {
			err = r.store.CreateHierarchyPath(ctx, db.CreateHierarchyPathParams{
				AncestorID:   *arg.ParentID,
				DescendantID: entity.Uuid,
				Depth:        1,
			})
			if err != nil {
				return nil, err
			}
			err = r.store.UpdateHierarchyPaths(ctx, entity.Uuid)
			if err != nil {
				return nil, err
			}
		}
		return FromSQLCEntity(entity), nil
	})
	if err != nil {
		return nil, err
	}
	return res.(*Organization), nil
}

func (r *repository) GetEntity(ctx context.Context, id uuid.UUID) (*Organization, error) {
	res, err := r.instrument(ctx, "GetEntity", func(ctx context.Context) (interface{}, error) {
		entity, err := r.store.GetEntity(ctx, id)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.ErrNotFound
			}
			return nil, err
		}
		return FromSQLCEntity(entity), nil
	})
	if err != nil {
		return nil, err
	}
	return res.(*Organization), nil
}

func (r *repository) GetEntityByCode(ctx context.Context, code string) (*Organization, error) {
	res, err := r.instrument(ctx, "GetEntityByCode", func(ctx context.Context) (interface{}, error) {
		entity, err := r.store.GetEntityByCode(ctx, &code)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.ErrNotFound
			}
			return nil, err
		}
		return FromSQLCEntity(entity), nil
	})
	if err != nil {
		return nil, err
	}
	return res.(*Organization), nil
}

func (r *repository) GetEntityByName(ctx context.Context, name string) (*Organization, error) {
	res, err := r.instrument(ctx, "GetEntityByName", func(ctx context.Context) (interface{}, error) {
		entity, err := r.store.GetEntityByName(ctx, name)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.ErrNotFound
			}
			return nil, err
		}
		return FromSQLCEntity(entity), nil
	})
	if err != nil {
		return nil, err
	}
	return res.(*Organization), nil
}

func (r *repository) GetEntityWithHierarchy(ctx context.Context, id uuid.UUID) (*db.GetEntityWithHierarchyInfoRow, error) {
	res, err := r.instrument(ctx, "GetEntityWithHierarchy", func(ctx context.Context) (interface{}, error) {
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			return nil, errors.ErrTenantIDNotInContext
		}
		return r.store.GetEntityWithHierarchyInfo(ctx, db.GetEntityWithHierarchyInfoParams{
			Uuid:     id,
			TenantID: tenantID,
		})
	})
	if err != nil {
		return nil, err
	}
	return res.(*db.GetEntityWithHierarchyInfoRow), nil
}

func (r *repository) ListEntities(ctx context.Context, arg db.ListEntitiesWithPaginationParams) ([]*Organization, error) {
	res, err := r.instrument(ctx, "ListEntities", func(ctx context.Context) (interface{}, error) {
		entities, err := r.store.ListEntitiesWithPagination(ctx, arg)
		if err != nil {
			return nil, err
		}
		orgs := make([]*Organization, len(entities))
		for i, e := range entities {
			orgs[i] = FromSQLCEntity(e)
		}
		return orgs, nil
	})
	if err != nil {
		return nil, err
	}
	return res.([]*Organization), nil
}

func (r *repository) UpdateEntity(ctx context.Context, arg db.UpdateEntityParams) (*Organization, error) {
	res, err := r.instrument(ctx, "UpdateEntity", func(ctx context.Context) (interface{}, error) {
		oldEntity, err := r.store.GetEntity(ctx, arg.Uuid)
		if err != nil {
			return nil, err
		}

		entity, err := r.store.UpdateEntity(ctx, arg)
		if err != nil {
			return nil, err
		}

		parentChanged := (arg.ParentID == nil && oldEntity.ParentID != nil) ||
			(arg.ParentID != nil && oldEntity.ParentID == nil) ||
			(arg.ParentID != nil && oldEntity.ParentID != nil && *arg.ParentID != *oldEntity.ParentID)

		if parentChanged {
			if arg.ParentID != nil {
				err := r.store.MoveEntityToNewParent(ctx, db.MoveEntityToNewParentParams{
					EntityID:    arg.Uuid,
					NewParentID: *arg.ParentID,
				})
				if err != nil {
					return nil, err
				}
			} else {
				// TODO: a query to move entity to root is needed
			}
		}

		return FromSQLCEntity(entity), nil
	})
	if err != nil {
		return nil, err
	}
	return res.(*Organization), nil
}

func (r *repository) DeleteEntity(ctx context.Context, id uuid.UUID, permanent bool) error {
	_, err := r.instrument(ctx, "DeleteEntity", func(ctx context.Context) (interface{}, error) {
		if permanent {
			_ = r.store.DeleteHierarchyPaths(ctx, id)
			// TODO: DeleteEntityStatesForEntity is not defined in the sql
			return nil, r.store.HardDeleteEntity(ctx, id)
		}
		return nil, r.store.SoftDeleteEntity(ctx, id)
	})
	return err
}

func (r *repository) RestoreEntity(ctx context.Context, id uuid.UUID) error {
	_, err := r.instrument(ctx, "RestoreEntity", func(ctx context.Context) (interface{}, error) {
		err := r.store.RestoreEntity(ctx, id)
		if err != nil {
			return nil, err
		}
		return nil, r.store.RebuildHierarchyPaths(ctx)
	})
	return err
}

func (r *repository) GetNextSequence(ctx context.Context, arg db.GetNextSequenceNumberParams) (int32, error) {
	res, err := r.instrument(ctx, "GetNextSequence", func(ctx context.Context) (interface{}, error) {
		return r.store.GetNextSequenceNumber(ctx, arg)
	})
	if err != nil {
		return 0, err
	}
	return res.(int32), nil
}

func (r *repository) ResetEntitySequence(ctx context.Context, arg db.ResetEntityStateSequenceParams) error {
	_, err := r.instrument(ctx, "ResetEntitySequence", func(ctx context.Context) (interface{}, error) {
		return nil, r.store.ResetEntityStateSequence(ctx, arg)
	})
	return err
}

func (r *repository) GetEntityChildren(ctx context.Context, parentId uuid.UUID) ([]*Organization, error) {
	res, err := r.instrument(ctx, "GetEntityChildren", func(ctx context.Context) (interface{}, error) {
		entities, err := r.store.GetEntityChildren(ctx, parentId)
		if err != nil {
			return nil, err
		}
		orgs := make([]*Organization, len(entities))
		for i, e := range entities {
			orgs[i] = FromSQLCEntity(e)
		}
		return orgs, nil
	})
	if err != nil {
		return nil, err
	}
	return res.([]*Organization), nil
}

func (r *repository) GetEntityAncestors(ctx context.Context, id uuid.UUID) ([]db.GetEntityAncestorsRow, error) {
	res, err := r.instrument(ctx, "GetEntityAncestors", func(ctx context.Context) (interface{}, error) {
		return r.store.GetEntityAncestors(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return res.([]db.GetEntityAncestorsRow), nil
}

func (r *repository) GetEntityTree(ctx context.Context, tenantId uuid.UUID) ([]db.GetEntityTreeStructureRow, error) {
	res, err := r.instrument(ctx, "GetEntityTree", func(ctx context.Context) (interface{}, error) {
		return r.store.GetEntityTreeStructure(ctx, tenantId)
	})
	if err != nil {
		return nil, err
	}
	return res.([]db.GetEntityTreeStructureRow), nil
}

func (r *repository) ValidateEntityName(ctx context.Context, name string, id uuid.UUID) (bool, error) {
	res, err := r.instrument(ctx, "ValidateEntityName", func(ctx context.Context) (interface{}, error) {
		return r.store.ValidateEntityName(ctx, db.ValidateEntityNameParams{
			Name: name,
			Uuid: id,
		})
	})
	if err != nil {
		return false, err
	}
	return res.(bool), nil
}

func (r *repository) ValidateEntityCode(ctx context.Context, code string, id uuid.UUID) (bool, error) {
	res, err := r.instrument(ctx, "ValidateEntityCode", func(ctx context.Context) (interface{}, error) {
		return r.store.ValidateEntityCode(ctx, db.ValidateEntityCodeParams{
			Code: code,
			Uuid: id,
		})
	})
	if err != nil {
		return false, err
	}
	return res.(bool), nil
}

func (r *repository) ValidateEntityParent(ctx context.Context, parentId, childId uuid.UUID) (bool, error) {
	res, err := r.instrument(ctx, "ValidateEntityParent", func(ctx context.Context) (interface{}, error) {
		return r.store.ValidateEntityParent(ctx, db.ValidateEntityParentParams{
			ParentID: &parentId,
			ChildID:  childId,
		})
	})
	if err != nil {
		return false, err
	}
	return res.(bool), nil
}

func (r *repository) CheckCircularReference(ctx context.Context, parentId, childId uuid.UUID) (bool, error) {
	res, err := r.instrument(ctx, "CheckCircularReference", func(ctx context.Context) (interface{}, error) {
		return r.store.CheckCircularReference(ctx, db.CheckCircularReferenceParams{
			DescendantID: parentId,
			AncestorID:   childId,
		})
	})
	if err != nil {
		return false, err
	}
	return res.(bool), nil
}
