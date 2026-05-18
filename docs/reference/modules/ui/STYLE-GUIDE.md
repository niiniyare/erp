# Documentation Style Guide

> Applies to all files under `docs/reference/modules/ui/`

This guide is the rule set for writing and reviewing documentation in this module. When in doubt, consult these rules rather than personal preference.

---

## 1. Tone

Write as an experienced colleague explaining something to someone competent. Not a tutorial for beginners, not a formal specification. Direct, accurate, and slightly terse.

**Good:**

> `UIContext` does not exist. The correct type is `UISessionContext`.

> The cache key is built after `AuthzStage`. Building it before would be a privilege escalation defect.

> Use `ASTFn` for new pages. `Fn` is kept for the migration window only.

**Bad:**

> In this section, we will explore the various types available to the developer in the UI context system, which includes both the legacy and modern approaches.

> Please note that it is very important to always remember to use the correct type name.

> The system provides a robust caching mechanism that ensures optimal performance.

Rules:
- No "please", "note that", "it is important to", "we will explore"
- No throat-clearing sentences that repeat what the heading already says
- No hedging ("may", "might", "could") when the behavior is deterministic
- No marketing language ("robust", "powerful", "elegant", "seamless")
- Short sentences. One idea per sentence.

---

## 2. Section Pattern: Why → What → How → Caveats

Every major section follows this structure:

```markdown
## Section Title

**Why**: One or two sentences explaining why this exists / why it matters.
(Skip for reference sections where "why" is obvious.)

**What**: What it is, at a conceptual level. Not how to use it yet.

**How**: The actual usage, with a runnable code example.

**Caveats**: Edge cases, common mistakes, gotchas. Use a table or list.
```

Example using the permission key format:

```markdown
## Permission Keys

**Why**: `sess.Can()` checks a pre-resolved map keyed by `resource.action`. Getting the format wrong means the check silently returns false and the button disappears for everyone.

**What**: A permission key is a dot-separated string with the resource name first and the action second: `invoice.view`, `leave_request.approve`.

**How**:
    sess.Can("view", "invoice")      // checks "invoice.view" → true/false
    sess.Can("approve", "leave_request") // checks "leave_request.approve"

**Caveats**:
- `Can("invoice", "view")` checks "view.invoice" — always false — wrong order
- Keys are case-sensitive. "Invoice.View" is not "invoice.view"
- Use `CanAny()` for OR logic, `CanAll()` for AND logic
```

---

## 3. Code Blocks

**Rules:**

1. All code blocks must have a language tag. Use `go`, `json`, `bash`, `mermaid`, `markdown` as appropriate. An untagged block is rejected in review.

2. Code examples must be syntactically correct. If you are showing a fragment, use `// ...` to indicate omitted sections. Do not show code that would not compile or that would fail validation.

3. Go code examples must use the actual type names from the codebase. Do not invent types.

4. JSON examples must be valid JSON. Use `"..."` for string placeholders, not `<placeholder>`.

5. For long examples, show only the relevant parts. Use `// ...` for the rest.

**Good:**
```go
func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:  "/finance/invoices",
        Module: "finance",
        Title:  "Invoices",
        ASTFn:  PageSchema,
    })
}
```

**Bad:**
```
func init() {
    registry.RegisterPage(PageRegistration{
        Route: "finance/invoices",  // missing leading slash — invalid
        ...
    })
}
```

---

## 4. Status Badges

Use these exact formats to indicate implementation status. Consistent formatting allows grep and tooling to find them.

```markdown
Implemented — feature is in the codebase and working
Not implemented — feature does not exist yet
Partial — feature exists but is incomplete (explain what's missing inline)
Deprecated — exists but scheduled for removal (state the replacement)
```

In tables:

```markdown
| Feature | Status |
|---------|--------|
| Pipeline stages | Implemented |
| Persona detection | Not implemented |
| `PageFn` path | Deprecated (use `ASTPageFn`) |
```

Do not use checkmarks, crosses, or emoji for status in tables. They do not render consistently across all Markdown renderers.

---

## 5. Cross-Reference Rules

Always use relative paths from the current file's directory:

```markdown
See [Pipeline Deep Dive](../02-architecture/02-pipeline-deep-dive.md)
See [Glossary](../appendices/A-glossary.md#uicontext-vs-uisessioncontext-critical)
```

Never use absolute paths. Never use bare filenames without directory. Never link to GitHub URLs for internal docs.

Fragment identifiers (`#section-heading`) must match the actual heading text, lowercased and with spaces replaced by hyphens. Test links before committing.

When referencing code, use backtick-quoted paths:

```markdown
Defined in `internal/web/ui/types.go`.
Called by `AuthzStage` in `internal/web/stages/authz.go`.
```

---

## 6. Non-Technical Sections

Mark sections intended for non-technical readers with the `📖` marker in the heading:

```markdown
## 📖 Why This Matters
```

Rules for non-technical sections:
- Maximum 200 words
- No code blocks
- No type names or function names
- Explain using analogies or plain English descriptions of behavior
- Must stand alone — a non-technical reader should not need to read the technical sections to understand it

These sections are most valuable in `01-theory/` files and at the top of architecture documents. They are optional in `03-implementation/` and `04-reference/` files.

---

## 7. Technical Sections

Mark sections for developers with the `⚙️` marker in the heading:

```markdown
## ⚙️ Cache Key Construction
```

Rules for technical sections:
- Must include a runnable code example or a concrete command
- Must reference the actual file and line range (or function name) that implements the described behavior
- May use type names, function names, and data key constants freely
- If the section describes behavior that can be misconfigured, include a Caveats subsection

---

## 8. Vaporware Rule

**Never document a feature as if it exists when it does not.**

If a feature is planned but not implemented:
- Move it to `05-roadmap/02-planned.md`
- In other files, link to the roadmap entry rather than describing the feature
- If you must mention it (e.g., in a gap analysis), use the exact phrase "Not implemented" followed by a link to the roadmap

Violations of this rule are the most harmful documentation errors. A developer who reads "persona detection returns `sess.Persona()`" and writes code against it will waste hours debugging a missing method.

Test: before publishing any description of a feature, confirm it compiles and runs. If you cannot confirm it, it goes in the roadmap.

---

## 9. Audit Header Template

Every file in this module must have an audit header as the first element after the title:

```markdown
# File Title

> Last verified: YYYY-MM-DD | Code pointer: `path/to/relevant/file.go`
```

`Last verified` is the date the author last read the actual code and confirmed the documentation matches it. This is not a modification date — it is an accuracy date.

`Code pointer` is the primary file or package that implements what this document describes. Use the shortest accurate path relative to the repository root.

When updating a document after a code change, update the `Last verified` date.

---

## 10. Heading Conventions

**H1 (`#`)**: File title only. One per file. Must match the filename without the numeric prefix and extension (e.g., `02-pipeline-deep-dive.md` → `# Pipeline Deep Dive`).

**H2 (`##`)**: Major sections. Should be navigable independently. Use title case.

**H3 (`###`)**: Subsections within an H2. Use title case.

**H4 (`####`)**: Rarely needed. Use only for named items within a subsection (e.g., individual stage descriptions in the pipeline doc). Use sentence case.

**No bold headings**: Do not use `**Heading**` as a substitute for a real heading level. If it needs to be navigable or linkable, use `##` or `###`.

**Heading anchors**: GitHub-flavored Markdown auto-generates anchors from headings. Spaces become hyphens, uppercase becomes lowercase, special characters are dropped. Test any cross-reference fragment identifier against the actual rendered heading.
