package activities

import (
	"context"
	"encoding/json"
	"fmt"

	// "github.com/niiniyare/erp/internal/core/abac" // TODO: Fix import path
	"github.com/niiniyare/erp/internal/core/identity"
	// "github.com/niiniyare/erp/internal/core/resource" // TODO: Fix import path - resource package doesn't exist
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// AttributeCollectionActivities implements the Temporal activities for collecting attributes.
type AttributeCollectionActivities struct {
	identityService identity.Service
	resourceRepo    resource.Repository // Assuming a resource repository interface
	logger          logger.Logger
}

// NewAttributeCollectionActivities creates a new instance of AttributeCollectionActivities.
func NewAttributeCollectionActivities(
	identityService identity.Service,
	resourceRepo resource.Repository,
	logger logger.Logger,
) *AttributeCollectionActivities {
	return &AttributeCollectionActivities{
		identityService: identityService,
		resourceRepo:    resourceRepo,
		logger:          logger,
	}
}

// CollectUserAttributesActivity collects attributes for a given user.
func (a *AttributeCollectionActivities) CollectUserAttributesActivity(ctx context.Context, input abac.CollectUserAttributesActivityInput) (*abac.CollectUserAttributesActivityOutput, error) {
	a.logger.InfoContext(ctx, "Starting CollectUserAttributesActivity", logger.Fields{"user_id": input.UserID})

	user, err := a.identityService.GetUserByID(ctx, input.UserID)
	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to get user by ID", logger.Fields{"user_id": input.UserID, "error": err})
		return nil, errors.WrapSimpleError(err, "USER_FETCH_FAILED", 500)
	}

	attributes := make(map[string]interface{})
	// User attributes
	attributes["id"] = user.ID
	attributes["email"] = user.Email
	attributes["user_type"] = user.UserType
	attributes["account_status"] = user.AccountStatus
	attributes["mfa_enabled"] = user.MfaEnabled

	// Person attributes
	if user.PersonID != nil {
		person, err := a.identityService.GetPersonByID(ctx, *user.PersonID)
		if err == nil {
			attributes["person_id"] = person.ID
			attributes["person_type"] = person.PersonType
			// Unmarshal JSONB attributes
			var personAttrs map[string]interface{}
			if err := json.Unmarshal(person.SecurityAttributes, &personAttrs); err == nil {
				for k, v := range personAttrs {
					attributes["person_"+k] = v
				}
			}
		}
	}

	// Employee attributes
	if user.EmployeeID != nil {
		employee, err := a.identityService.GetEmployeeByID(ctx, *user.EmployeeID)
		if err == nil {
			attributes["employee_id"] = employee.ID
			attributes["position_title"] = employee.PositionTitle
			attributes["security_level"] = employee.SecurityLevel
			// Unmarshal JSONB attributes
			var employeeAttrs map[string]interface{}
			if err := json.Unmarshal(employee.AccessAttributes, &employeeAttrs); err == nil {
				for k, v := range employeeAttrs {
					attributes["employee_"+k] = v
				}
			}
		}
	}

	output := &abac.CollectUserAttributesActivityOutput{
		Attributes: attributes,
	}

	a.logger.InfoContext(ctx, "Finished CollectUserAttributesActivity", logger.Fields{"user_id": input.UserID, "attribute_count": len(attributes)})
	return output, nil
}

// CollectResourceAttributesActivity collects attributes for a given resource.
func (a *AttributeCollectionActivities) CollectResourceAttributesActivity(ctx context.Context, input abac.CollectResourceAttributesActivityInput) (*abac.CollectResourceAttributesActivityOutput, error) {
	a.logger.InfoContext(ctx, "Starting CollectResourceAttributesActivity", logger.Fields{"resource_id": input.ResourceID})

	res, err := a.resourceRepo.GetByID(ctx, input.ResourceID)
	if err != nil {
		a.logger.ErrorContext(ctx, "Failed to get resource by ID", logger.Fields{"resource_id": input.ResourceID, "error": err})
		return nil, errors.WrapSimpleError(err, "RESOURCE_FETCH_FAILED", 500)
	}

	attributes := make(map[string]interface{})
	attributes["id"] = res.ID
	attributes["name"] = res.Name
	attributes["resource_type"] = res.ResourceType
	attributes["path"] = res.Path

	// Unmarshal JSONB attributes
	var resourceAttrs map[string]interface{}
	if err := json.Unmarshal(res.ResourceAttributes, &resourceAttrs); err == nil {
		for k, v := range resourceAttrs {
			attributes[k] = v
		}
	}

	output := &abac.CollectResourceAttributesActivityOutput{
		Attributes: attributes,
	}

	a.logger.InfoContext(ctx, "Finished CollectResourceAttributesActivity", logger.Fields{"resource_id": input.ResourceID, "attribute_count": len(attributes)})
	return output, nil
}

// BuildEnvironmentContextActivity builds the environmental context for the evaluation.
func (a *AttributeCollectionActivities) BuildEnvironmentContextActivity(ctx context.Context, input abac.BuildEnvironmentContextActivityInput) (*abac.BuildEnvironmentContextActivityOutput, error) {
	a.logger.InfoContext(ctx, "Starting BuildEnvironmentContextActivity")

	contextMap := make(map[string]interface{})
	contextMap["timestamp"] = input.Timestamp
	contextMap["ip_address"] = input.IPAddress
	contextMap["user_agent"] = input.UserAgent

	// In a real implementation, you might do IP geolocation, device fingerprinting, etc.
	// For now, we'll just use the provided location data.
	if input.Location != nil {
		for k, v := range input.Location {
			contextMap["location_"+k] = v
		}
	}

	output := &abac.BuildEnvironmentContextActivityOutput{
		Context: contextMap,
	}

	a.logger.InfoContext(ctx, "Finished BuildEnvironmentContextActivity", logger.Fields{"context_keys": len(contextMap)})
	return output, nil
}
