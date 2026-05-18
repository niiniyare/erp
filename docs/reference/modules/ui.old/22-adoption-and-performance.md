[<-- Back to Index](README.md)

## Adoption & Performance

### Why Adoption Is a Design Problem

The best-designed system fails if people don't use it. In African enterprise markets, two factors uniquely shape adoption:

```markdown
1. The system must be FAST on variable connectivity (3G is common)
2. It must feel immediately useful WITHOUT a training course
```

### First-Run Experience

The first time a user logs in, the system should feel already set up for them:

```markdown
EMPLOYEE (first login):
├── Home screen immediately shows: leave balance, next shift, pending items
├── Nothing to configure
└── Can submit a leave request in the first 30 seconds

LINE MANAGER (first login):
├── Team is pre-loaded. No setup needed.
├── Pending approvals already surfaced
└── Dashboard shows department attendance/leave overview

HR STAFF (first login):
├── Setup checklist: jurisdiction → leave types → departments → grades
├── Progress is saved — stop and return any time
└── Checklist disappears once all steps are complete

ONBOARDING TIPS:
├── Contextual tips appear the FIRST TIME a feature is encountered
├── Not on login. Not as a tour that blocks the UI.
├── A small "?" badge dismisses permanently when clicked
└── Stored as dismissed in user preferences — never returns
```

### Performance on Variable Connectivity

**Skeleton Loading** — replace spinners with shape placeholders:

```markdown
SPINNER (bad):
  ⟳ Loading...   [blank white space]

SKELETON (good):
  ████████████████  ← table header shimmer
  ██████   ████   ← first row loading...
  ████     ██████  ← second row...
  Users see progress. Perceived wait time is lower.
```

**Lazy Loading** — load only what is on screen first:
- Dashboard: load the most important 2 widgets first, heavy charts last
- Table: default page size 25 rows (power users set to 100)

**Optimistic Updates** — update the UI before the server confirms:

```markdown
User clicks "Approve" → UI marks the request as approved immediately
  → If server confirms: done
  → If server returns error: revert + show error
Result: app feels instantaneous even on slow connections
```

**Chart backgrounds** — always `backgroundColor: "transparent"` in ECharts config, so the chart does not load with a white flash before the AMIS page background renders.

### Reducing Friction

**One-click from notification to action.** A line manager who receives a notification must approve directly from the notification — maximum two clicks from notification receipt to approved state.

**Undo instead of confirmation.** For reversible actions:
```markdown
CONFIRMATION (adds friction):
  "Are you sure you want to approve?"  [Yes] [No]

UNDO (faster, equally safe):
  ✓ Leave request approved — Amina Odhiambo    [Undo]  5s
```

Reserve confirmation dialogs for irreversible or high-consequence actions:
- Terminating an employee
- Reversing a posted payroll run
- Deleting a department with active staff

**Progress Preservation.** A user who starts a payroll wizard and closes the browser tab must return to exactly where they left off. Never silently lose user input.

### Measuring Adoption

Track a small set of engagement signals (with user awareness):

```markdown
TIME TO FIRST MEANINGFUL ACTION
→ How long does a new user take to complete their first real task?

TASK COMPLETION RATE
→ What percentage of initiated actions are completed vs abandoned?

RETURN VISIT FREQUENCY
→ Are users coming back regularly or only when forced?

SUPPORT TICKET TOPICS
→ Which features generate the most confusion?

FEATURE USAGE RATE
→ Which features are used by <10% of users who have access?
```

Features with low usage and high support tickets are **redesigned**, not just documented better.

---
