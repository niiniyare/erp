---
title: "Real-Time UI Updates"
volume: "IV — Rendering Architecture"
chapter: "20-B"
phase: 2
status: draft
audience: [Platform Engineers, Backend Engineers, Web Engineers, Mobile Engineers]
---

# Chapter 20-B — Real-Time UI Updates

> **Volume:** IV — Rendering Architecture
> **Audience:** Platform Engineers, Backend Engineers, Web Engineers, Mobile Engineers
> **Prerequisites:** Chapter 20 — Action System, Chapter 19-B — Backend-Driven Navigation
> **Phase:** 2

---

## 20B.1 Real-Time UI Update Philosophy

ERP systems are not static report viewers. A pump attendant approving a sale, a finance manager posting a journal entry, a supplier confirming a delivery — these events change shared state that other users need to see immediately. In a Nairobi fuel station where the operations manager monitors live pump sessions on a tablet and the cashier processes payments at the counter, a 30-second delay in state propagation is not a UX inconvenience; it is an operational failure.

AwoERP's real-time update architecture is built around a single principle: **the server is always the source of truth; the UI is always a derivative view**. Real-time updates are not the primary read path — they are *invalidation signals* that tell the UI to refresh specific parts of its view from the server. This approach is robust under poor connectivity (a common reality in Kenyan deployments outside Nairobi), because a missed signal simply means a slightly stale view that self-corrects on the next poll or reconnection.

The platform uses Server-Sent Events (SSE) as the primary transport for web clients because SSE is unidirectional, works over standard HTTP/1.1, traverses corporate proxies without configuration, and requires no special server infrastructure beyond a long-lived HTTP connection. WebSocket is reserved for bidirectional scenarios (e.g. live chat support within a portal). Flutter mobile clients use a similar SSE connection or a platform-native equivalent.

---

## 20B.2 Use Cases

### 20B.2.1 Live Dashboard Widget Refresh

Dashboard widgets showing KPIs (today's fuel sales in litres, cash collected, invoices pending) must reflect current data without requiring manual refresh. The widget's `service` component in amis polls a data source URL, and SSE `datasource_invalidate` events accelerate the refresh cycle by triggering an immediate re-fetch rather than waiting for the poll interval.

### 20B.2.2 Approval Queue Count Badge Updates

Navigation badges showing pending approval counts (leave requests, purchase orders, invoices for payment) update in real time as items enter or leave the approval queue. The badge SSE topic (`badge:approvals:pending:{user_id}`) receives `badge_update` events whenever the queue count changes for the current user.

### 20B.2.3 Push-Driven Form Field Updates (e.g. FX Rate Feed)

Finance forms for foreign currency transactions (common for Kenyan importers dealing in USD, EUR, GBP) display live exchange rates. The KES/USD rate field subscribes to an FX rate SSE topic and updates in real time via `field_patch` events as the treasury team or an automated feed updates rates.

### 20B.2.4 Workflow Status Transitions

Temporal workflow state transitions (an approval workflow moving from PENDING to APPROVED) are pushed to the UI via SSE `component_replace` events that update the workflow status indicator component without requiring a full page reload.

### 20B.2.5 Inventory Level Alerts

When a fuel tank level crosses a threshold (e.g. below 20% capacity), an SSE alert event updates the tank level indicator on the operations dashboard and triggers a notification badge on the inventory nav item.

---

## 20B.3 Transport Mechanisms

### 20B.3.1 Server-Sent Events (SSE) — Primary for Web

SSE is implemented as a long-lived `GET /api/v1/ui/events` endpoint. The connection is tenant-scoped and authenticated:

```go
// internal/api/handlers/sse/handler.go

type SSEHandler struct {
    hub    *SSEHub
    authz  AuthzService
    audit  AuditService
}

func (h *SSEHandler) Subscribe(c *fiber.Ctx) error {
    tenantID := shared.GetTenantID(c.UserContext())
    userID   := shared.GetUserID(c.UserContext())

    // Validate topics from query param
    requestedTopics := c.Query("topics") // comma-separated
    allowedTopics, err := h.authz.FilterSSETopics(c.UserContext(), userID, tenantID, splitTopics(requestedTopics))
    if err != nil {
        return fiber.ErrForbidden
    }

    // Set SSE headers
    c.Set("Content-Type", "text/event-stream")
    c.Set("Cache-Control", "no-cache")
    c.Set("Connection", "keep-alive")
    c.Set("X-Accel-Buffering", "no") // disable Nginx buffering

    // Create subscriber
    sub := h.hub.Subscribe(tenantID, userID, allowedTopics)
    defer h.hub.Unsubscribe(sub)

    ctx := c.UserContext()
    for {
        select {
        case event, ok := <-sub.Events:
            if !ok {
                return nil
            }
            fmt.Fprintf(c.Response().BodyWriter(), "event: %s\ndata: %s\n\n", event.Type, event.JSON())
            c.Response().Flush()

        case <-ctx.Done():
            return nil
        }
    }
}
```

### 20B.3.2 WebSocket — For Bidirectional Channels

WebSocket is reserved for surfaces that require bidirectional communication. Currently this applies to:
- Live chat between portal users and tenant staff
- Collaborative editing of documents (draft purchase orders with multiple editors)

WebSocket connections use the same authentication and tenant context as SSE but require explicit opt-in from the surface definition.

### 20B.3.3 Long Polling — Fallback

For clients that cannot maintain SSE connections (certain corporate proxy configurations), a long-polling fallback is available at `GET /api/v1/ui/poll?topics=...&last_event_id=...`. The server holds the request for up to 30 seconds, returning immediately if any event arrives for the subscribed topics.

### 20B.3.4 Temporal Workflow Signals to UI

Temporal workflows can push state changes to connected UI clients via the SSE hub. The workflow activity calls `SSEHub.Publish()` after a state transition:

```go
// internal/core/ui/sse/hub.go

type SSEHub struct {
    mu          sync.RWMutex
    subscribers map[string][]*Subscriber // key: "tenant_id:topic"
    metrics     *SSEMetrics
}

func (h *SSEHub) Publish(tenantID uuid.UUID, topic string, event UIEvent) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    key := fmt.Sprintf("%s:%s", tenantID, topic)
    for _, sub := range h.subscribers[key] {
        select {
        case sub.Events <- event:
        default:
            h.metrics.DroppedEvents.Inc()
        }
    }
}

// Called from a Temporal workflow activity
func (a *ApprovalActivities) NotifyApprovalComplete(ctx context.Context, params NotifyParams) error {
    event := UIEvent{
        Type: "component_replace",
        Payload: map[string]any{
            "component_id": fmt.Sprintf("approval-status-%s", params.DocumentID),
            "schema": buildApprovedStatusSchema(params),
        },
    }
    a.sseHub.Publish(params.TenantID, fmt.Sprintf("workflow:%s", params.WorkflowID), event)
    return nil
}
```

---

## 20B.4 Update Payload Schema

All SSE events carry a typed payload in the `data` field. The event type determines the shape:

```go
// internal/core/ui/sse/events.go

type UIEvent struct {
    ID        string         `json:"id"`          // Monotonic event ID per tenant
    Type      UIEventType    `json:"type"`
    TenantID  uuid.UUID      `json:"tenant_id"`
    Timestamp time.Time      `json:"ts"`
    Payload   map[string]any `json:"payload"`
}

type UIEventType string

const (
    EventFieldPatch           UIEventType = "field_patch"
    EventComponentReplace     UIEventType = "component_replace"
    EventDataSourceInvalidate UIEventType = "datasource_invalidate"
    EventBadgeUpdate          UIEventType = "badge_update"
    EventNavDiff              UIEventType = "nav_diff"
    EventNavInvalidated       UIEventType = "nav_invalidated"
    EventSessionExpired       UIEventType = "session_expired"
    EventAlert                UIEventType = "alert"
)
```

### 20B.4.1 Targeted Field Update (`field_patch`)

Updates specific fields within the currently open form or view without re-rendering:

```json
{
  "id": "evt-001",
  "type": "field_patch",
  "ts": "2026-06-06T08:05:00Z",
  "payload": {
    "surface_id": "finance.journal-entry.new",
    "patches": [
      { "field": "fx_rate_usd_kes", "value": 131.50, "display": "KES 131.50 / USD" },
      { "field": "fx_rate_eur_kes", "value": 142.20, "display": "KES 142.20 / EUR" }
    ]
  }
}
```

The amis renderer applies field patches by dispatching an amis action to update the form data context:

```javascript
window.addEventListener('awo:sse', function(e) {
    var event = e.detail;
    if (event.type === 'field_patch' && event.payload.surface_id === currentSurfaceId) {
        event.payload.patches.forEach(function(patch) {
            amisInstance.updateData({ [patch.field]: patch.value });
        });
    }
});
```

### 20B.4.2 Full Component Refresh (`component_replace`)

Replaces an entire component's schema and data:

```json
{
  "type": "component_replace",
  "payload": {
    "component_id": "widget-fuel-sales-today",
    "schema": {
      "type": "statistic",
      "title": "Today's Fuel Sales",
      "value": 84250,
      "unit": "Litres",
      "trend": "+12% vs yesterday"
    }
  }
}
```

### 20B.4.3 Data Source Invalidation Signal (`datasource_invalidate`)

Signals that a named data source should be re-fetched:

```json
{
  "type": "datasource_invalidate",
  "payload": {
    "datasource_ids": ["finance.accounts.list", "finance.dashboard.summary"],
    "reason": "journal_posted",
    "affected_resource_id": "jnl-uuid-123"
  }
}
```

The amis `service` component responds to this by re-fetching its API.

### 20B.4.4 Badge Count Update (`badge_update`)

```json
{
  "type": "badge_update",
  "payload": {
    "topic": "badge:approvals:pending:user-uuid",
    "count": 5,
    "delta": 1
  }
}
```

### 20B.4.5 Navigation Tree Diff (`nav_diff`)

```json
{
  "type": "nav_diff",
  "payload": {
    "since_version": 14,
    "new_version": 15,
    "added": [{ "group": "lpg", "item": { "key": "lpg.reports", "label": "LPG Reports" } }],
    "removed": [],
    "modified": []
  }
}
```

---

## 20B.5 Subscription Model

### 20B.5.1 Surface-Scoped Subscriptions

A client subscribes to topics relevant to the currently active surface. When a user navigates to `finance.invoices.list`, the client subscribes to `datasource:finance.invoices:{tenant_id}`. When they navigate away, the client unsubscribes.

```javascript
// amis surface mount/unmount lifecycle hooks
function onSurfaceMount(surfaceId, tenantId) {
    sseClient.subscribe([
        `datasource:${surfaceId}:${tenantId}`,
        `badge:${surfaceId}:${tenantId}`
    ]);
}

function onSurfaceUnmount(surfaceId, tenantId) {
    sseClient.unsubscribe([
        `datasource:${surfaceId}:${tenantId}`,
        `badge:${surfaceId}:${tenantId}`
    ]);
}
```

### 20B.5.2 Resource-Scoped Subscriptions

When a user opens a specific document (e.g., PO #PO-2026-001), the client subscribes to `resource:procurement.po:{po_id}:{tenant_id}`. This enables collaborative editing notifications and workflow status updates specific to that document.

### 20B.5.3 Tenant-Scoped Broadcast

Platform-admin events (module enable/disable, feature flag change) are broadcast to all sessions for a tenant using the topic `tenant:{tenant_id}:broadcast`. These events typically carry `nav_invalidated` payloads.

---

## 20B.6 Authorization for Real-Time Updates

SSE topic access is governed by the same Casbin permission model as API endpoints. The `SSEHub` validates topic access at subscription time:

```go
func (h *AuthzService) FilterSSETopics(ctx context.Context, userID, tenantID uuid.UUID, topics []string) ([]string, error) {
    var allowed []string
    for _, topic := range topics {
        permission := sseTopicToPermission(topic) // e.g. "datasource:finance.invoices" → "finance.invoices.read"
        if slices.Contains(getSessionPermissions(ctx), permission) {
            allowed = append(allowed, topic)
        }
    }
    return allowed, nil
}
```

> **⚠ Warning:** Never publish a `field_patch` or `component_replace` event containing data the subscribing user is not authorised to see. The publisher must verify the target user's permissions before including sensitive data in the event payload.

---

## 20B.7 Backpressure and Rate Limiting

Each SSE subscriber has a buffered channel with a configurable capacity (default: 100 events). If the channel is full (client is slow or disconnected), events are dropped and a `DroppedEvents` metric counter is incremented. On reconnection, the client uses `Last-Event-ID` to request missed events from a short-lived event log (Redis sorted set, 5-minute TTL):

```go
const sseChannelBuffer = 100
const sseEventLogTTL = 5 * time.Minute

type Subscriber struct {
    TenantID uuid.UUID
    UserID   uuid.UUID
    Topics   []string
    Events   chan UIEvent // buffered
}
```

### Rate Limiting

Temporal workflows and API handlers are rate-limited to 50 SSE publishes per second per tenant to prevent a high-volume batch operation from flooding connected clients:

```go
var tenantSSELimiter = rate.NewLimiter(rate.Limit(50), 100)

func (h *SSEHub) PublishRateLimited(tenantID uuid.UUID, topic string, event UIEvent) error {
    if !tenantSSELimiter.Allow() {
        return ErrSSERateLimited
    }
    h.Publish(tenantID, topic, event)
    return nil
}
```

---

## 20B.8 Reconnection and Missed Update Recovery

SSE clients implement automatic reconnection with exponential back-off using the browser's native SSE reconnect mechanism (`retry:` field) combined with `Last-Event-ID`:

```javascript
// web/src/sse/client.js

class AwoSSEClient {
    constructor(url, topics) {
        this.url = url;
        this.topics = topics;
        this.lastEventId = localStorage.getItem('sse-last-event-id') || '';
        this.reconnectDelay = 1000; // ms, doubles on each failure up to 30s
    }

    connect() {
        const url = new URL(this.url);
        url.searchParams.set('topics', this.topics.join(','));

        this.source = new EventSource(url.toString());

        this.source.onopen = () => {
            this.reconnectDelay = 1000; // reset
        };

        this.source.onmessage = (e) => {
            this.lastEventId = e.lastEventId;
            localStorage.setItem('sse-last-event-id', e.lastEventId);
            this.handleEvent(JSON.parse(e.data));
        };

        this.source.onerror = () => {
            this.source.close();
            setTimeout(() => this.connect(), this.reconnectDelay);
            this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000);
        };
    }
}
```

On reconnection with a `Last-Event-ID`, the server checks the Redis event log and replays missed events within the 5-minute window. Events older than 5 minutes are not replayed — instead a `datasource_invalidate` event for all subscribed topics is sent, triggering a full data refresh.

---

## 20B.9 Observability for Real-Time Channels

```go
var (
    sseActiveConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "awo_sse_active_connections",
        Help: "Current active SSE connections",
    }, []string{"portal"})

    sseEventsPublished = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "awo_sse_events_published_total",
        Help: "Total SSE events published",
    }, []string{"event_type", "portal"})

    sseEventsDropped = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "awo_sse_events_dropped_total",
        Help: "SSE events dropped due to full subscriber channel",
    }, []string{"event_type"})

    sseConnectionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "awo_sse_connection_duration_seconds",
        Buckets: []float64{30, 60, 300, 600, 1800, 3600},
    }, []string{"portal", "disconnect_reason"})
)
```

OpenTelemetry spans are created for each SSE publish operation to trace the end-to-end latency from event source (e.g. a Temporal activity completing) to SSE channel write.

> See Chapter 38 — Observability for alert thresholds and runbook references for SSE health degradation.
