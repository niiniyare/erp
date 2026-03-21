package repo

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	db "awo/db/sqlc"
	"awo/internal/core/iam/model"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
)

// personRepository implements PersonRepository interface
type personRepository struct {
	store   db.Store
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

// NewPersonRepository creates a new person repository
func NewPersonRepository(
	store db.Store,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) PersonRepository {
	return &personRepository{
		store:   store,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// Create creates a new person
func (r *personRepository) Create(ctx context.Context, person *model.Person) (*model.Person, error) {
	ctx, span := r.tracer.StartSpan(ctx, "person_repository.Create")
	defer span.End()

	var createdPerson *model.Person
	err := r.store.WithTenant(ctx, person.TenantID, func(ctx context.Context, store db.Store) error {
		// Convert domain model to SQLC params
		addressJSON, _ := json.Marshal(person.Address)
		securityAttrsJSON, _ := json.Marshal(person.SecurityAttributes)
		metadataJSON, _ := json.Marshal(person.Metadata)

		params := db.CreatePersonParams{
			EntityID:           person.EntityID,
			PersonType:         string(person.PersonType),
			FirstName:          person.FirstName,
			LastName:           person.LastName,
			MiddleName:         person.MiddleName,
			Email:              person.Email,
			Phone:              person.PhoneNumber,
			BirthDate:          person.BirthDate,
			NationalID:         person.NationalID,
			TaxID:              person.TaxID,
			Address:            addressJSON,
			SecurityAttributes: securityAttrsJSON,
			Metadata:           metadataJSON,
		}

		// Create person using SQLC
		dbPerson, err := store.CreatePerson(ctx, params)
		if err != nil {
			// Handle database-specific errors
			if isDuplicateKeyError(err) {
				return errors.NewBusinessError("PERSON_EMAIL_EXISTS", "Person with this email already exists")
			}
			return fmt.Errorf("failed to create person: %w", err)
		}

		// Convert back to domain model
		createdPerson = convertPersonToDomain(dbPerson)

		r.metrics.IncrementCounter("person_repository_create_success", nil)
		return nil
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("person_repository_create_error", nil)
		return nil, err
	}

	r.logger.InfoContext(ctx, "Person created successfully",
		logger.Fields{"person_id": createdPerson.ID, "tenant_id": person.TenantID})

	return createdPerson, nil
}

// GetByID retrieves a person by ID
func (r *personRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Person, error) {
	ctx, span := r.tracer.StartSpan(ctx, "person_repository.GetByID")
	defer span.End()

	tenantID := getTenantIDFromContext(ctx)

	var person *model.Person
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, store db.Store) error {
		dbPerson, err := store.GetPersonByID(ctx, id)
		if err != nil {
			if err == db.ErrNoRows {
				return errors.NewBusinessError("PERSON_NOT_FOUND", "Person not found")
			}
			return fmt.Errorf("failed to get person: %w", err)
		}

		// Convert to domain model
		person = convertPersonToDomain(dbPerson)
		return nil
	})
	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("person_repository_get_error", nil)
		return nil, err
	}

	r.metrics.IncrementCounter("person_repository_get_success", nil)
	return person, nil
}

// GetByEmail retrieves a person by email
func (r *personRepository) GetByEmail(ctx context.Context, email string) (*model.Person, error) {
	ctx, span := r.tracer.StartSpan(ctx, "person_repository.GetByEmail")
	defer span.End()

	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("GetByEmail not implemented")
}

// Update updates an existing person
func (r *personRepository) Update(ctx context.Context, person *model.Person) (*model.Person, error) {
	ctx, span := r.tracer.StartSpan(ctx, "person_repository.Update")
	defer span.End()

	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("Update not implemented")
}

// Delete soft deletes a person
func (r *personRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "person_repository.Delete")
	defer span.End()

	// TODO: Implement when SQLC query is available
	return fmt.Errorf("Delete not implemented")
}

// List lists persons with pagination
func (r *personRepository) List(ctx context.Context, limit, offset int) ([]*model.Person, error) {
	ctx, span := r.tracer.StartSpan(ctx, "person_repository.List")
	defer span.End()

	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("List not implemented")
}

// Count returns total number of persons
func (r *personRepository) Count(ctx context.Context) (int64, error) {
	// TODO: Implement when SQLC query is available
	return 0, fmt.Errorf("Count not implemented")
}

// ListByType lists persons by person type
func (r *personRepository) ListByType(ctx context.Context, personType model.PersonType) ([]*model.Person, error) {
	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("ListByType not implemented")
}

// SearchByName searches persons by name
func (r *personRepository) SearchByName(ctx context.Context, searchTerm string) ([]*model.Person, error) {
	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("SearchByName not implemented")
}

// GetPersonsForEntity retrieves all persons for an entity
func (r *personRepository) GetPersonsForEntity(ctx context.Context, entityID uuid.UUID) ([]*model.Person, error) {
	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("GetPersonsForEntity not implemented")
}

// Helper function to convert SQLC Person to domain model
func convertPersonToDomain(dbPerson *db.Person) *model.Person {
	person := &model.Person{
		ID:          dbPerson.ID,
		TenantID:    dbPerson.TenantID,
		EntityID:    dbPerson.EntityID,
		PersonType:  model.PersonType(dbPerson.PersonType),
		FirstName:   dbPerson.FirstName,
		LastName:    dbPerson.LastName,
		MiddleName:  dbPerson.MiddleName,
		Email:       dbPerson.Email,
		PhoneNumber: dbPerson.Phone,
		BirthDate:   dbPerson.BirthDate,
		NationalID:  dbPerson.NationalID,
		TaxID:       dbPerson.TaxID,
		IsActive:    getBoolValue(&dbPerson.IsActive),
		CreatedAt:   dbPerson.CreatedAt,
		UpdatedAt:   dbPerson.UpdatedAt,
	}

	// Unmarshal JSON fields - silently ignore errors for non-critical fields
	if dbPerson.Address != nil {
		_ = json.Unmarshal(dbPerson.Address, &person.Address)
	}
	if dbPerson.SecurityAttributes != nil {
		_ = json.Unmarshal(dbPerson.SecurityAttributes, &person.SecurityAttributes)
	}
	if dbPerson.Metadata != nil {
		_ = json.Unmarshal(dbPerson.Metadata, &person.Metadata)
	}

	return person
}
