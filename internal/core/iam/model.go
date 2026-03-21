package authz

// casbinModel is the Casbin CONF model used by every enforcer in this package.
//
// Key properties:
//   - deny-override: one explicit deny beats all allows
//   - role hierarchy: g(user, role, domain) — domain-scoped RBAC
//   - keyMatch2 on obj and act allows wildcard patterns like "invoice/*"
const casbinModel = `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, eft

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && keyMatch2(r.act, p.act)
`
