[<-- Back to Index](README.md)

> **🚫 NOT IMPLEMENTED.** This entire feature (inline alerts, toast notifications, notification
> panel, push notifications) does not exist in the current codebase. There is no backend handler,
> no frontend component, no database table, and no API endpoint for notifications.
>
> This document is the **design specification** for when notifications are built.
> See [Planned Features](../05-roadmap/02-planned.md#notification-system) for roadmap status.

## Notifications & Alerts

### The Four Notification Types

Use the right type for each situation. Mixing them up trains users to ignore them.

**Type 1 — Inline Alert (immediate, contextual)**

Appears on the current screen in response to something already visible. Not a popup. A banner within the relevant section.

```json
{
  "type":  "alert",
  "level": "warning",
  "body":  "This payroll run includes 3 employees whose timesheets have not been approved. They will be excluded from the run unless timesheets are approved first.",
  "actions": [
    { "type": "button", "label": "View Unapproved Timesheets",
      "actionType": "link", "link": "#timesheets?status=unapproved" }
  ]
}
```

Use for: validation warnings, data quality issues on the current screen, information the user needs before taking an action.

**Type 2 — Toast (transient feedback)**

Small message that appears briefly (3–5 seconds) after an action. Disappears automatically.

```javascript
// Triggered by AMIS onEvent success action
{
  "actionType": "toast",
  "args": {
    "msg":     "Leave request approved — Amina Odhiambo · 23–31 Dec",
    "msgType": "success",
    "position": "top-right"
  }
}
```

Use for: success confirmation after an action. Never for errors — errors must persist until resolved.

**Type 3 — Notification Panel (async, pull)**

The bell icon opens a panel listing notifications. Each item is actionable.

```markdown
NOTIFICATIONS                           Mark all read

  ● Payroll run ready for review                   2h ago
    November 2024 · 47 employees · KES 4.2M gross
    [Review Run →]

  ● Leave request pending your approval            1d ago
    Amina Odhiambo · Annual Leave · 23–31 Dec
    [Approve] [Decline]

  ✓ NSSF remittance schedule generated             2d ago
    November 2024 [Download →]
```

Go endpoint: `GET /api/v1/notifications` → returns list scoped to current user.

**Type 4 — Push Notification / Email (async, push)**

Used only for high-priority events when the user is not in the app.

Push triggers (configurable, opt-in by default):
- Leave request awaiting approval (> 24 hours pending)
- Payroll run ready for review
- Leave request approved or declined
- Payslip available

Email digest triggers:
- Daily summary of pending approvals
- Weekly payroll summary for finance managers
- Monthly statutory filing reminders

### Severity Levels

Consistent styling across all notification types:

| Level | Colour | Icon | Use For | Persistence |
|---|---|---|---|---|
| Info | Blue | ℹ | Tips, system notices | Auto-dismiss 5s |
| Success | Green | ✓ | Completed actions | Auto-dismiss 5s |
| Warning | Amber | ⚠ | Advisory — action recommended | Persist until dismissed |
| Error | Red | ✗ | Action failed — must be resolved | Persist until issue resolved |
| Critical | Red (bold) | ✗ | System-level issue | Persist until admin acknowledges |

### Alert vs Toast: When to Use Which

```markdown
SCENARIO                               TYPE          REASON
─────────────────────────────────────────────────────────────────
Form saves successfully                Toast         Transient — user knows it worked
Action fails (server error)            Alert/Toast   Error stays visible
Data issue visible on current screen   Inline Alert  Not a response to action — contextual
Pending approval added to inbox        Notification  Async — user not currently looking
Session about to expire                Inline Alert  Needs immediate attention
Network connection lost                Inline Alert  Persistent until resolved
```

### Notification Rules

```markdown
✅ Notifications are scoped to the user — managers only see their team's
✅ Every notification has a direct action link — no navigation required
✅ Notification context is complete — "Amina Odhiambo's leave (23–31 Dec)"
   NOT "You have a pending item"
✅ Older than 30 days are archived, not deleted
✅ User can configure which event types generate notifications
✅ Every email notification has a one-click unsubscribe
```

---
