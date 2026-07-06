// Package iam implements the Awo platform Identity and Access Management module.
//
// IAM manages users, roles, and sessions across all tenants. It owns three
// system entities that are mandatory — declaring any as CustomDefinition causes
// the registry validator to panic at startup:
//
//   - iam_user    — canonical user identity (email, password hash, MFA, status)
//   - iam_role    — named permission set within a tenant (Casbin subject)
//   - iam_session — active session tokens (Redis-cached, DB as source of truth)
//
// Additionally it registers two custom entities:
//
//   - iam_api_token  — long-lived machine-to-machine bearer tokens
//   - iam_user_role  — join table linking users to roles within a tenant
//
// Registration happens in init() so definitions are available before the first
// request is served.
package iam

import "awo.so/awo/def"

func init() {
	def.Register(&UserDefinition)
	def.Register(&RoleDefinition)
	def.Register(&SessionDefinition)
	def.Register(&APITokenDefinition)
	def.Register(&UserRoleDefinition)
}
