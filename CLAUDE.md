## Pro Workflow

### Self-Correction
When corrected, propose rule → add to LEARNED after approval.

### Planning
Multi-file: plan first, wait for "proceed".

### Quality
After edits: lint, typecheck, test.

### LEARNED
- Never run `go build`, `go run`, `go vet`, or any Go toolchain commands — tell the user to run them instead (Termux sandbox blocks /tmp creation)
