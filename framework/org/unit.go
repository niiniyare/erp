// Package org is a compatibility shim. New code should import awo.so/framework/platform/org.
package org

import platformorg "awo.so/framework/platform/org"

// Unit re-exported from platform/org.
type Unit = platformorg.Unit

// UnitPath re-exported from platform/org.
type UnitPath = platformorg.UnitPath

// UnitType re-exported from platform/org.
type UnitType = platformorg.UnitType

// ValidationStatus re-exported from platform/org.
type ValidationStatus = platformorg.ValidationStatus

const (
	UnitTypeCompany    = platformorg.UnitTypeCompany
	UnitTypeSubsidiary = platformorg.UnitTypeSubsidiary
	UnitTypeRegion     = platformorg.UnitTypeRegion
	UnitTypeBranch     = platformorg.UnitTypeBranch
	UnitTypeLocation   = platformorg.UnitTypeLocation
	UnitTypeDepartment = platformorg.UnitTypeDepartment
	UnitTypeDivision   = platformorg.UnitTypeDivision
	UnitTypeCostCenter = platformorg.UnitTypeCostCenter
	UnitTypeProject    = platformorg.UnitTypeProject
	UnitTypeBudgetUnit = platformorg.UnitTypeBudgetUnit

	ValidationStatusPending = platformorg.ValidationStatusPending
	ValidationStatusValid   = platformorg.ValidationStatusValid
	ValidationStatusWarning = platformorg.ValidationStatusWarning
	ValidationStatusError   = platformorg.ValidationStatusError
)
