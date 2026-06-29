// Package org is a compatibility shim. New code should import awo.so/framework/platform/org.
package org

import platformorg "awo.so/framework/platform/org"

// ScopeLevel re-exported from platform/org.
type ScopeLevel = platformorg.ScopeLevel

const (
	ScopeLevelGlobal  = platformorg.ScopeLevelGlobal
	ScopeLevelTenant  = platformorg.ScopeLevelTenant
	ScopeLevelUnit    = platformorg.ScopeLevelUnit
)

// Scope re-exported from platform/org.
type Scope = platformorg.Scope

var (
	TenantOnly = platformorg.TenantOnly
	WithUnit   = platformorg.WithUnit
)

var (
	ErrCrossTenantAccess = platformorg.ErrCrossTenantAccess
	ErrCrossUnitAccess   = platformorg.ErrCrossUnitAccess
	ErrScopeInsufficient = platformorg.ErrScopeInsufficient
)

// NodeType re-exported from platform/org.
type NodeType = platformorg.NodeType
