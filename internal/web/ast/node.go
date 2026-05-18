// Package ast defines the typed UI Abstract Syntax Tree for the Go-first UI
// compilation system.
//
// Every AMIS schema node is represented as a concrete Go type that implements
// the Node interface. Nodes are immutable after construction — all methods use
// value receivers. Compilation to AMIS JSON happens exclusively via CompileTree,
// which enforces validation before emission.
//
// Architecture contract:
//   - No map[string]any in this package except as the output of Compile().
//   - No IAM imports — permission resolution happens before nodes are built.
//   - No I/O — nodes are pure data; PageFn / ASTPageFn remain side-effect-free.
//   - No runtime mutation — nodes are built once and compiled once per request.
package ast

// Node is the base interface for every typed AMIS schema node.
//
// Implementations must use value receivers on all methods to enforce
// immutability — a value receiver prevents callers from mutating the node
// through an interface variable.
//
// Every concrete node type must also provide a compile-time assertion:
//
//	var _ Node = PageNode{}
type Node interface {
	// NodeType returns the AMIS "type" field value (e.g. "page", "crud", "form").
	NodeType() string

	// Validate checks that the node's invariants hold.
	// Returns nil when valid. Returns a *sharedErrors.BusinessError with code
	// prefix "AST_" when invalid.
	//
	// Validate is called by CompileTree before Compile. Nodes must never
	// call Validate on themselves inside Compile — CompileTree owns that ordering.
	Validate() error

	// Compile emits the node as an AMIS-compliant map[string]any.
	// Must only be called after Validate() returns nil.
	// Structural invariants (syncLocation, transparent background, etc.) are
	// emitted unconditionally — they are not checked by NormalizeStage for
	// AST-compiled schemas.
	Compile() map[string]any
}

// ContainerNode is an optional extension for nodes that own child nodes
// (Page, Grid, Form, etc.). CompileTree recurses into children via this
// interface to collect all validation errors before any JSON is emitted.
type ContainerNode interface {
	Node
	Children() []Node
}
