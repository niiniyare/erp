# Status Component

**FILE PURPOSE**: State indicator implementation and specifications  
**SCOPE**: All status variants, visual indicators, and semantic meanings  
**TARGET AUDIENCE**: Developers implementing status displays, state indicators, and system feedback

## 📋 Component Overview

The Status component provides clear visual indicators for system states, process status, and operational conditions. It combines semantic colors, icons, and text to communicate status information effectively while maintaining accessibility standards.

### Schema Reference
- **Primary Schema**: `StatusSchema.json`
- **Related Schemas**: `BadgeObject.json`, `IconSchema.json`
- **Base Interface**: Display element with semantic state meaning

## 🎨 Status Types

### Basic Status
**Purpose**: Simple state indicators with semantic colors

```go
// Basic status configuration
basicStatus := StatusProps{
    Level:       "success",
    Text:        "Online",
    ShowIcon:    true,
    ShowDot:     true,
}

// Generated Templ component
templ BasicStatus(props StatusProps) {
    <span class={ fmt.Sprintf("status status-%s status-%s", props.Level, props.Size) }
          role="status"
          aria-label={ fmt.Sprintf("Status: %s", props.Text) }>
        
        if props.ShowDot {
            <span class="status-dot" aria-hidden="true"></span>
        }
        
        if props.ShowIcon && props.Icon != "" {
            <span class="status-icon">
                @Icon(IconProps{
                    Name: props.Icon, 
                    Size: getStatusIconSize(props.Size),
                })
            </span>
        } else if props.ShowIcon {
            <span class="status-icon">
                @Icon(IconProps{
                    Name: getDefaultIcon(props.Level), 
                    Size: getStatusIconSize(props.Size),
                })
            </span>
        }
        
        if props.Text != "" {
            <span class="status-text">{ props.Text }</span>
        }
        
        if props.Timestamp != "" {
            <span class="status-timestamp">{ props.Timestamp }</span>
        }
    </span>
}

// Default icons by status level
func getDefaultIcon(level StatusLevel) string {
    switch level {
    case StatusSuccess:
        return "check-circle"
    case StatusWarning:
        return "alert-triangle"
    case StatusError:
        return "x-circle"
    case StatusInfo:
        return "info"
    case StatusPending:
        return "clock"
    default:
        return "circle"
    }
}
```

### Animated Status
**Purpose**: Status indicators with loading animations and transitions

```go
animatedStatus := StatusProps{
    Level:    "pending",
    Text:     "Processing",
    Animated: true,
    Duration: 2000,
}

templ AnimatedStatus(props StatusProps) {
    <span class={ fmt.Sprintf("status status-%s status-animated", props.Level) }
          style={ fmt.Sprintf("--animation-duration: %dms", props.Duration) }
          role="status"
          aria-live="polite">
        
        if props.ShowDot {
            <span class="status-dot status-dot-animated" aria-hidden="true"></span>
        }
        
        if props.ShowIcon {
            <span class="status-icon">
                @Icon(IconProps{
                    Name:      getDefaultIcon(props.Level),
                    Size:      getStatusIconSize(props.Size),
                    Animation: getStatusAnimation(props.Level),
                })
            </span>
        }
        
        <span class="status-text">{ props.Text }</span>
        
        if props.Progress >= 0 {
            <span class="status-progress">
                ({ fmt.Sprintf("%.0f%%", props.Progress*100) })
            </span>
        }
    </span>
}

func getStatusAnimation(level StatusLevel) string {
    switch level {
    case StatusPending:
        return "spin"
    case StatusLoading:
        return "pulse"
    default:
        return "none"
    }
}
```

### Status Card
**Purpose**: Detailed status display with additional context

```go
statusCard := StatusProps{
    Level:       "warning",
    Title:       "System Maintenance",
    Text:        "Scheduled maintenance in progress",
    Description: "Some features may be temporarily unavailable",
    ShowIcon:    true,
    Dismissible: true,
}

templ StatusCard(props StatusProps) {
    <div class={ fmt.Sprintf("status-card status-card-%s", props.Level) }
         role="alert"
         aria-labelledby={ props.ID + "-title" }
         aria-describedby={ props.ID + "-description" }>
        
        <div class="status-card-header">
            if props.ShowIcon {
                <div class="status-card-icon">
                    @Icon(IconProps{
                        Name: getDefaultIcon(props.Level),
                        Size: "lg",
                    })
                </div>
            }
            
            <div class="status-card-content">
                if props.Title != "" {
                    <h4 id={ props.ID + "-title" } class="status-card-title">
                        { props.Title }
                    </h4>
                }
                
                if props.Text != "" {
                    <p class="status-card-text">{ props.Text }</p>
                }
            </div>
            
            if props.Dismissible {
                <button 
                    type="button"
                    class="status-card-dismiss"
                    @click={ props.OnDismiss }
                    aria-label="Dismiss status">
                    @Icon(IconProps{Name: "x", Size: "sm"})
                </button>
            }
        </div>
        
        if props.Description != "" {
            <div id={ props.ID + "-description" } class="status-card-description">
                { props.Description }
            </div>
        }
        
        if len(props.Actions) > 0 {
            <div class="status-card-actions">
                for _, action := range props.Actions {
                    @Button(ButtonProps{
                        Text:    action.Text,
                        Variant: action.Variant,
                        Size:    "sm",
                        OnClick: action.OnClick,
                    })
                }
            </div>
        }
    </div>
}
```

### Status List
**Purpose**: Multiple status items in organized display

```go
statusList := StatusListProps{
    Title:    "System Status",
    Items: []StatusItem{
        {Level: "success", Text: "API Server", Timestamp: "2 minutes ago"},
        {Level: "success", Text: "Database", Timestamp: "1 minute ago"},
        {Level: "warning", Text: "Cache Server", Timestamp: "5 minutes ago"},
        {Level: "error", Text: "Email Service", Timestamp: "10 minutes ago"},
    },
    ShowTimestamps: true,
    RefreshInterval: 30000,
}

templ StatusList(props StatusListProps) {
    <div class="status-list" 
         x-data={ fmt.Sprintf(`{
             items: %s,
             refreshInterval: %d,
             lastRefresh: Date.now(),
             async refresh() {
                 try {
                     const response = await fetch('/api/status');
                     this.items = await response.json();
                     this.lastRefresh = Date.now();
                 } catch (error) {
                     console.error('Failed to refresh status:', error);
                 }
             },
             formatTimestamp(timestamp) {
                 return new Date(timestamp).toLocaleTimeString();
             }
         }`, toJSON(props.Items), props.RefreshInterval) }
         x-init="setInterval(() => refresh(), refreshInterval)">
        
        <div class="status-list-header">
            <h3 class="status-list-title">{ props.Title }</h3>
            
            <button 
                type="button"
                class="status-refresh-btn"
                @click="refresh()"
                aria-label="Refresh status">
                @Icon(IconProps{Name: "refresh", Size: "sm"})
            </button>
        </div>
        
        <div class="status-list-content">
            <template x-for="item in items" :key="item.id">
                <div class="status-list-item">
                    <span :class="`status status-${item.level}`" role="status">
                        <span class="status-dot" aria-hidden="true"></span>
                        <span class="status-text" x-text="item.text"></span>
                    </span>
                    
                    if props.ShowTimestamps {
                        <span class="status-timestamp" x-text="formatTimestamp(item.timestamp)"></span>
                    }
                </div>
            </template>
        </div>
        
        <div class="status-list-footer">
            <span class="last-refresh">
                Last updated: <span x-text="formatTimestamp(lastRefresh)"></span>
            </span>
        </div>
    </div>
}
```

### Progress Status
**Purpose**: Status with progress indication

```go
progressStatus := StatusProps{
    Level:    "pending",
    Text:     "Uploading files",
    Progress: 0.65,
    ShowProgressBar: true,
    Animated: true,
}

templ ProgressStatus(props StatusProps) {
    <div class="status-progress-container" 
         x-data={ fmt.Sprintf(`{
             progress: %f,
             get percentage() {
                 return Math.round(this.progress * 100);
             }
         }`, props.Progress) }>
        
        <div class="status-progress-header">
            @BasicStatus(StatusProps{
                Level:    props.Level,
                Text:     props.Text,
                ShowIcon: props.ShowIcon,
                ShowDot:  props.ShowDot,
                Animated: props.Animated,
            })
            
            <span class="progress-percentage" x-text="`${percentage}%`">
                { fmt.Sprintf("%.0f%%", props.Progress*100) }
            </span>
        </div>
        
        if props.ShowProgressBar {
            <div class="progress-bar">
                <div class="progress-bar-fill" 
                     :style="`width: ${percentage}%`"
                     :class="`progress-${props.Level}`"></div>
            </div>
        }
        
        if props.EstimatedTime != "" {
            <div class="progress-meta">
                <span class="estimated-time">{ props.EstimatedTime } remaining</span>
            </div>
        }
    </div>
}
```

## 🎯 Props Interface

### Core Properties
```go
type StatusProps struct {
    // Identity
    ID     string `json:"id"`
    TestID string `json:"testid"`
    
    // Content
    Text        string `json:"text"`        // Status text
    Title       string `json:"title"`       // Title for cards
    Description string `json:"description"` // Additional description
    Icon        string `json:"icon"`        // Custom icon
    
    // Status Level
    Level StatusLevel `json:"level"` // success, warning, error, info, pending, etc.
    
    // Appearance
    Size     StatusSize    `json:"size"`     // xs, sm, md, lg
    Variant  StatusVariant `json:"variant"`  // dot, icon, card, badge
    
    // Features
    ShowIcon  bool `json:"showIcon"`  // Display icon
    ShowDot   bool `json:"showDot"`   // Display status dot
    Animated  bool `json:"animated"`  // Enable animations
    Duration  int  `json:"duration"`  // Animation duration
    
    // Progress
    Progress        float64 `json:"progress"`        // 0.0 to 1.0
    ShowProgressBar bool    `json:"showProgressBar"` // Show progress bar
    EstimatedTime   string  `json:"estimatedTime"`   // Time remaining
    
    // Timestamps
    Timestamp string `json:"timestamp"` // Status timestamp
    ShowTime  bool   `json:"showTime"`  // Show timestamp
    
    // Interaction
    Dismissible bool   `json:"dismissible"` // Allow dismissal
    OnDismiss   string `json:"onDismiss"`   // Dismiss handler
    OnClick     string `json:"onClick"`     // Click handler
    
    // Actions
    Actions []StatusAction `json:"actions"` // Action buttons
    
    // Styling
    Class string            `json:"className"`
    Style map[string]string `json:"style"`
    
    // Accessibility
    AriaLabel string `json:"ariaLabel"`
    Live      string `json:"live"` // polite, assertive, off
}
```

### List Properties
```go
type StatusListProps struct {
    // Identity
    ID    string `json:"id"`
    Title string `json:"title"`
    
    // Content
    Items []StatusItem `json:"items"`
    
    // Features
    ShowTimestamps  bool `json:"showTimestamps"`
    RefreshInterval int  `json:"refreshInterval"` // Milliseconds
    AutoRefresh     bool `json:"autoRefresh"`
    
    // Layout
    Compact  bool `json:"compact"`  // Compact display
    Grouped  bool `json:"grouped"`  // Group by level
    Sortable bool `json:"sortable"` // Allow sorting
    
    // Events
    OnRefresh string `json:"onRefresh"`
    OnItemClick string `json:"onItemClick"`
}

type StatusItem struct {
    ID        string      `json:"id"`
    Level     StatusLevel `json:"level"`
    Text      string      `json:"text"`
    Icon      string      `json:"icon"`
    Timestamp string      `json:"timestamp"`
    Metadata  interface{} `json:"metadata"`
}
```

### Action Properties
```go
type StatusAction struct {
    Text    string `json:"text"`
    Variant string `json:"variant"` // primary, secondary, danger
    OnClick string `json:"onClick"`
    Icon    string `json:"icon"`
}
```

### Status Levels
```go
type StatusLevel string

const (
    StatusSuccess StatusLevel = "success" // Green - completed, operational
    StatusWarning StatusLevel = "warning" // Yellow - caution, attention needed
    StatusError   StatusLevel = "error"   // Red - failed, critical
    StatusInfo    StatusLevel = "info"    // Blue - informational
    StatusPending StatusLevel = "pending" // Gray - in progress, waiting
    StatusLoading StatusLevel = "loading" // Blue - loading, processing
    StatusIdle    StatusLevel = "idle"    // Gray - inactive, standby
)
```

### Size Variants
```go
type StatusSize string

const (
    StatusXS StatusSize = "xs"    // 12px indicators
    StatusSM StatusSize = "sm"    // 14px indicators
    StatusMD StatusSize = "md"    // 16px indicators (default)
    StatusLG StatusSize = "lg"    // 20px indicators
)
```

### Visual Variants
```go
type StatusVariant string

const (
    StatusDot   StatusVariant = "dot"   // Dot indicator only
    StatusIcon  StatusVariant = "icon"  // Icon indicator
    StatusCard  StatusVariant = "card"  // Card layout
    StatusBadge StatusVariant = "badge" // Badge style
)
```

## 🎨 Styling Implementation

### Base Status Styles
```css
.status {
    display: inline-flex;
    align-items: center;
    gap: var(--space-xs);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    line-height: 1.4;
    
    .status-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        flex-shrink: 0;
        
        &.status-dot-animated {
            animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
        }
    }
    
    .status-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    
    .status-text {
        color: var(--color-text-primary);
        line-height: inherit;
    }
    
    .status-timestamp {
        color: var(--color-text-tertiary);
        font-size: var(--font-size-xs);
        margin-left: var(--space-sm);
    }
    
    .status-progress {
        color: var(--color-text-secondary);
        font-size: var(--font-size-xs);
        margin-left: var(--space-xs);
    }
}
```

### Size Variants
```css
/* Extra Small */
.status-xs {
    font-size: var(--font-size-xs);
    gap: 2px;
    
    .status-dot {
        width: 4px;
        height: 4px;
    }
    
    .status-icon {
        width: 10px;
        height: 10px;
    }
}

/* Small */
.status-sm {
    font-size: var(--font-size-xs);
    gap: var(--space-xs);
    
    .status-dot {
        width: 5px;
        height: 5px;
    }
    
    .status-icon {
        width: 12px;
        height: 12px;
    }
}

/* Medium (Default) */
.status-md {
    font-size: var(--font-size-sm);
    gap: var(--space-xs);
    
    .status-dot {
        width: 6px;
        height: 6px;
    }
    
    .status-icon {
        width: 14px;
        height: 14px;
    }
}

/* Large */
.status-lg {
    font-size: var(--font-size-base);
    gap: var(--space-sm);
    
    .status-dot {
        width: 8px;
        height: 8px;
    }
    
    .status-icon {
        width: 16px;
        height: 16px;
    }
}
```

### Level-Specific Colors
```css
/* Success */
.status-success {
    .status-dot {
        background: var(--color-success);
    }
    
    .status-icon {
        color: var(--color-success);
    }
}

/* Warning */
.status-warning {
    .status-dot {
        background: var(--color-warning);
    }
    
    .status-icon {
        color: var(--color-warning);
    }
}

/* Error */
.status-error {
    .status-dot {
        background: var(--color-error);
    }
    
    .status-icon {
        color: var(--color-error);
    }
}

/* Info */
.status-info {
    .status-dot {
        background: var(--color-info);
    }
    
    .status-icon {
        color: var(--color-info);
    }
}

/* Pending */
.status-pending {
    .status-dot {
        background: var(--color-text-secondary);
    }
    
    .status-icon {
        color: var(--color-text-secondary);
    }
}

/* Loading */
.status-loading {
    .status-dot {
        background: var(--color-primary);
    }
    
    .status-icon {
        color: var(--color-primary);
    }
}

/* Idle */
.status-idle {
    .status-dot {
        background: var(--color-text-tertiary);
    }
    
    .status-icon {
        color: var(--color-text-tertiary);
    }
}
```

### Status Card Styles
```css
.status-card {
    border-radius: var(--radius-lg);
    border: 1px solid var(--color-border-light);
    background: var(--color-bg-surface);
    transition: var(--transition-base);
    
    .status-card-header {
        display: flex;
        align-items: flex-start;
        gap: var(--space-md);
        padding: var(--space-lg);
    }
    
    .status-card-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 40px;
        height: 40px;
        border-radius: var(--radius-md);
        flex-shrink: 0;
    }
    
    .status-card-content {
        flex: 1;
        min-width: 0;
        
        .status-card-title {
            font-size: var(--font-size-base);
            font-weight: var(--font-weight-semibold);
            color: var(--color-text-primary);
            margin: 0 0 var(--space-xs) 0;
            line-height: 1.3;
        }
        
        .status-card-text {
            color: var(--color-text-secondary);
            font-size: var(--font-size-sm);
            line-height: 1.4;
            margin: 0;
        }
    }
    
    .status-card-dismiss {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 24px;
        height: 24px;
        background: none;
        border: none;
        color: var(--color-text-tertiary);
        cursor: pointer;
        border-radius: var(--radius-xs);
        transition: var(--transition-quick);
        flex-shrink: 0;
        
        &:hover {
            background: var(--color-bg-hover);
            color: var(--color-text-secondary);
        }
        
        &:focus {
            outline: none;
            box-shadow: 0 0 0 2px var(--color-primary-light);
        }
    }
    
    .status-card-description {
        padding: 0 var(--space-lg) var(--space-lg) calc(56px + var(--space-lg));
        color: var(--color-text-secondary);
        font-size: var(--font-size-sm);
        line-height: 1.5;
    }
    
    .status-card-actions {
        display: flex;
        gap: var(--space-sm);
        padding: 0 var(--space-lg) var(--space-lg) calc(56px + var(--space-lg));
    }
}

/* Card level-specific styles */
.status-card-success {
    border-color: var(--color-success-light);
    background: var(--color-success-bg);
    
    .status-card-icon {
        background: var(--color-success-light);
        color: var(--color-success);
    }
}

.status-card-warning {
    border-color: var(--color-warning-light);
    background: var(--color-warning-bg);
    
    .status-card-icon {
        background: var(--color-warning-light);
        color: var(--color-warning-dark);
    }
}

.status-card-error {
    border-color: var(--color-error-light);
    background: var(--color-error-bg);
    
    .status-card-icon {
        background: var(--color-error-light);
        color: var(--color-error);
    }
}

.status-card-info {
    border-color: var(--color-info-light);
    background: var(--color-info-bg);
    
    .status-card-icon {
        background: var(--color-info-light);
        color: var(--color-info);
    }
}
```

### Status List Styles
```css
.status-list {
    border: 1px solid var(--color-border-light);
    border-radius: var(--radius-lg);
    background: var(--color-bg-surface);
    
    .status-list-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: var(--space-lg);
        border-bottom: 1px solid var(--color-border-light);
        
        .status-list-title {
            font-size: var(--font-size-lg);
            font-weight: var(--font-weight-semibold);
            color: var(--color-text-primary);
            margin: 0;
        }
        
        .status-refresh-btn {
            display: flex;
            align-items: center;
            justify-content: center;
            width: 32px;
            height: 32px;
            background: none;
            border: none;
            color: var(--color-text-secondary);
            cursor: pointer;
            border-radius: var(--radius-md);
            transition: var(--transition-quick);
            
            &:hover {
                background: var(--color-bg-hover);
                color: var(--color-text-primary);
            }
            
            &:focus {
                outline: none;
                box-shadow: 0 0 0 2px var(--color-primary-light);
            }
        }
    }
    
    .status-list-content {
        .status-list-item {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: var(--space-md) var(--space-lg);
            border-bottom: 1px solid var(--color-border-light);
            transition: var(--transition-quick);
            
            &:last-child {
                border-bottom: none;
            }
            
            &:hover {
                background: var(--color-bg-hover);
            }
            
            .status-timestamp {
                color: var(--color-text-tertiary);
                font-size: var(--font-size-xs);
            }
        }
    }
    
    .status-list-footer {
        padding: var(--space-md) var(--space-lg);
        border-top: 1px solid var(--color-border-light);
        background: var(--color-bg-secondary);
        border-radius: 0 0 var(--radius-lg) var(--radius-lg);
        
        .last-refresh {
            color: var(--color-text-tertiary);
            font-size: var(--font-size-xs);
        }
    }
}
```

### Progress Status Styles
```css
.status-progress-container {
    .status-progress-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: var(--space-sm);
        
        .progress-percentage {
            font-size: var(--font-size-sm);
            font-weight: var(--font-weight-semibold);
            color: var(--color-text-secondary);
        }
    }
    
    .progress-bar {
        width: 100%;
        height: 4px;
        background: var(--color-bg-secondary);
        border-radius: var(--radius-full);
        overflow: hidden;
        
        .progress-bar-fill {
            height: 100%;
            border-radius: var(--radius-full);
            transition: width 0.3s ease;
            
            &.progress-success {
                background: var(--color-success);
            }
            
            &.progress-warning {
                background: var(--color-warning);
            }
            
            &.progress-error {
                background: var(--color-error);
            }
            
            &.progress-info {
                background: var(--color-info);
            }
            
            &.progress-pending {
                background: var(--color-text-secondary);
            }
        }
    }
    
    .progress-meta {
        margin-top: var(--space-xs);
        
        .estimated-time {
            color: var(--color-text-tertiary);
            font-size: var(--font-size-xs);
        }
    }
}
```

### Animation Styles
```css
/* Status animations */
@keyframes pulse {
    0%, 100% {
        opacity: 1;
        transform: scale(1);
    }
    50% {
        opacity: 0.7;
        transform: scale(1.1);
    }
}

@keyframes spin {
    from {
        transform: rotate(0deg);
    }
    to {
        transform: rotate(360deg);
    }
}

.status-animated {
    .status-dot-animated {
        animation: pulse var(--animation-duration, 2s) cubic-bezier(0.4, 0, 0.6, 1) infinite;
    }
    
    .status-icon {
        animation: var(--animation-type, none) var(--animation-duration, 1s) linear infinite;
    }
}
```

## ⚙️ Advanced Features

### Real-time Status Updates
```go
templ LiveStatus(props StatusProps) {
    <div class="live-status" 
         x-data={ fmt.Sprintf(`{
             status: '%s',
             level: '%s',
             lastUpdate: Date.now(),
             connection: null,
             connect() {
                 this.connection = new WebSocket('/ws/status');
                 this.connection.onmessage = (event) => {
                     const data = JSON.parse(event.data);
                     this.status = data.status;
                     this.level = data.level;
                     this.lastUpdate = Date.now();
                 };
                 this.connection.onclose = () => {
                     setTimeout(() => this.connect(), 5000);
                 };
             },
             disconnect() {
                 if (this.connection) {
                     this.connection.close();
                 }
             }
         }`, props.Text, props.Level) }
         x-init="connect()"
         x-destroy="disconnect()">
        
        <span :class="`status status-${level}`" role="status" aria-live="polite">
            <span class="status-dot" aria-hidden="true"></span>
            <span class="status-text" x-text="status"></span>
        </span>
    </div>
}
```

### Status History
```go
templ StatusHistory(props StatusHistoryProps) {
    <div class="status-history" 
         x-data={ fmt.Sprintf(`{
             history: %s,
             showAll: false,
             get visibleHistory() {
                 return this.showAll ? this.history : this.history.slice(0, 5);
             }
         }`, toJSON(props.History)) }>
        
        <div class="history-header">
            <h4>Status History</h4>
        </div>
        
        <div class="history-timeline">
            <template x-for="item in visibleHistory" :key="item.id">
                <div class="timeline-item">
                    <div class="timeline-marker">
                        <span :class="`status-dot status-${item.level}`"></span>
                    </div>
                    <div class="timeline-content">
                        <span class="timeline-status" x-text="item.status"></span>
                        <span class="timeline-time" x-text="item.timestamp"></span>
                    </div>
                </div>
            </template>
        </div>
        
        <button x-show="history.length > 5 && !showAll" 
                @click="showAll = true"
                class="show-more-btn">
            Show all history
        </button>
    </div>
}
```

### Conditional Status
```go
templ ConditionalStatus(props StatusProps) {
    <div class="conditional-status" 
         x-data={ fmt.Sprintf(`{
             conditions: %s,
             currentStatus: '%s',
             get computedStatus() {
                 for (const condition of this.conditions) {
                     if (this.evaluateCondition(condition.condition)) {
                         return condition.status;
                     }
                 }
                 return this.currentStatus;
             },
             evaluateCondition(condition) {
                 // Implement condition evaluation logic
                 return eval(condition);
             }
         }`, toJSON(props.Conditions), props.Level) }>
        
        <span :class="`status status-${computedStatus.level}`" 
              role="status" 
              aria-live="polite">
            <span class="status-dot" aria-hidden="true"></span>
            <span class="status-text" x-text="computedStatus.text"></span>
        </span>
    </div>
}
```

## 📱 Responsive Design

### Mobile Optimizations
```css
/* Mobile-specific status styling */
@media (max-width: 479px) {
    .status {
        font-size: var(--font-size-xs);
        
        .status-timestamp {
            display: none; /* Hide timestamps on mobile */
        }
    }
    
    .status-card {
        .status-card-header {
            padding: var(--space-md);
            gap: var(--space-sm);
        }
        
        .status-card-icon {
            width: 32px;
            height: 32px;
        }
        
        .status-card-description {
            padding: 0 var(--space-md) var(--space-md) calc(40px + var(--space-sm));
        }
        
        .status-card-actions {
            padding: 0 var(--space-md) var(--space-md) calc(40px + var(--space-sm));
            flex-direction: column;
            align-items: stretch;
        }
    }
    
    .status-list {
        .status-list-header {
            padding: var(--space-md);
        }
        
        .status-list-item {
            padding: var(--space-sm) var(--space-md);
            flex-direction: column;
            align-items: flex-start;
            gap: var(--space-xs);
        }
    }
}
```

### Touch Interactions
```css
/* Touch-friendly interactions */
@media (hover: none) {
    .status-card-dismiss {
        min-width: 32px;
        min-height: 32px;
        
        &:hover {
            background: none;
        }
    }
    
    .status-refresh-btn {
        min-width: 44px;
        min-height: 44px;
        
        &:hover {
            background: none;
        }
    }
    
    .status-list-item {
        &:hover {
            background: var(--color-bg-surface);
        }
        
        &:active {
            background: var(--color-bg-pressed);
        }
    }
}
```

## ♿ Accessibility Implementation

### ARIA Support
```go
func (status StatusProps) GetAriaAttributes() map[string]string {
    attrs := make(map[string]string)
    
    // Live region for dynamic status
    if status.Live != "" {
        attrs["aria-live"] = status.Live
        attrs["aria-atomic"] = "true"
    } else {
        attrs["aria-live"] = "polite"
    }
    
    // Role
    attrs["role"] = "status"
    
    // Accessible name
    if status.AriaLabel != "" {
        attrs["aria-label"] = status.AriaLabel
    } else {
        attrs["aria-label"] = fmt.Sprintf("Status: %s %s", status.Level, status.Text)
    }
    
    return attrs
}
```

### Screen Reader Support
```go
templ AccessibleStatus(props StatusProps) {
    <span class={ props.GetClasses() }
          for attrName, attrValue := range props.GetAriaAttributes() {
              { attrName }={ attrValue }
          }>
        
        <!-- Visual indicators are hidden from screen readers -->
        <span class="status-dot" aria-hidden="true"></span>
        
        <!-- Text content is announced -->
        <span class="status-text">{ props.Text }</span>
        
        <!-- Screen reader only context -->
        <span class="sr-only">
            Status level: { string(props.Level) }
        </span>
        
        if props.Timestamp != "" {
            <span class="sr-only">
                Last updated: { props.Timestamp }
            </span>
        }
    </span>
}
```

### Keyboard Navigation
```css
/* Focus management */
.status-card-dismiss,
.status-refresh-btn {
    &:focus {
        outline: none;
        box-shadow: 0 0 0 2px var(--color-primary-light);
        border-radius: var(--radius-sm);
    }
    
    &:focus-visible {
        box-shadow: 0 0 0 2px var(--color-primary);
    }
}

/* High contrast mode support */
@media (prefers-contrast: high) {
    .status-dot {
        border: 1px solid currentColor;
    }
    
    .status-card {
        border-width: 2px;
        border-style: solid;
    }
    
    .progress-bar-fill {
        border: 1px solid currentColor;
    }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
    .status-dot-animated,
    .status-icon {
        animation: none;
    }
    
    .progress-bar-fill {
        transition: none;
    }
}
```

## 🧪 Testing Guidelines

### Unit Tests
```go
func TestStatusComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    StatusProps
        expected []string
    }{
        {
            name: "success status",
            props: StatusProps{
                Level:    "success",
                Text:     "Online",
                ShowDot:  true,
                ShowIcon: true,
            },
            expected: []string{"status", "status-success", "Online", "status-dot"},
        },
        {
            name: "error status with card",
            props: StatusProps{
                Level:   "error",
                Title:   "System Error",
                Text:    "Database connection failed",
                Variant: "card",
            },
            expected: []string{"status-card", "status-card-error", "System Error"},
        },
        {
            name: "animated pending status",
            props: StatusProps{
                Level:    "pending",
                Text:     "Processing",
                Animated: true,
            },
            expected: []string{"status-animated", "Processing"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            component := renderStatus(tt.props)
            for _, expected := range tt.expected {
                assert.Contains(t, component, expected)
            }
        })
    }
}
```

### Accessibility Tests
```javascript
describe('Status Accessibility', () => {
    test('has proper role and live region', () => {
        const status = render(<Status level="success" text="Online" />);
        const statusElement = screen.getByRole('status');
        expect(statusElement).toHaveAttribute('aria-live', 'polite');
    });
    
    test('announces status changes', () => {
        const { rerender } = render(<Status level="pending" text="Loading" />);
        const statusElement = screen.getByRole('status');
        
        rerender(<Status level="success" text="Complete" />);
        expect(statusElement).toHaveTextContent('Complete');
    });
    
    test('status cards have proper alert role', () => {
        const status = render(
            <Status 
                level="error" 
                title="Error" 
                text="Something went wrong" 
                variant="card" 
            />
        );
        const alertElement = screen.getByRole('alert');
        expect(alertElement).toBeInTheDocument();
    });
});
```

### Visual Regression Tests
```javascript
test.describe('Status Visual Tests', () => {
    test('all status levels and variants', async ({ page }) => {
        await page.goto('/components/status');
        
        // Test all levels
        const levels = ['success', 'warning', 'error', 'info', 'pending'];
        for (const level of levels) {
            await expect(page.locator(`[data-level="${level}"]`)).toHaveScreenshot(`status-${level}.png`);
        }
        
        // Test all variants
        const variants = ['dot', 'icon', 'card', 'badge'];
        for (const variant of variants) {
            await expect(page.locator(`[data-variant="${variant}"]`)).toHaveScreenshot(`status-${variant}.png`);
        }
        
        // Test animations
        await expect(page.locator('[data-animated="true"]')).toHaveScreenshot('status-animated.png');
    });
});
```

## 📚 Usage Examples

### System Status Dashboard
```go
templ SystemStatusDashboard() {
    @StatusList(StatusListProps{
        Title: "System Components",
        Items: []StatusItem{
            {Level: "success", Text: "API Gateway", Timestamp: "2 minutes ago"},
            {Level: "success", Text: "User Service", Timestamp: "1 minute ago"},
            {Level: "warning", Text: "Cache Layer", Timestamp: "5 minutes ago"},
            {Level: "error", Text: "Email Service", Timestamp: "10 minutes ago"},
        },
        ShowTimestamps:  true,
        RefreshInterval: 30000,
    })
}
```

### User Status Indicators
```go
templ UserStatus() {
    <div class="user-info">
        <img src="/avatar.jpg" alt="User avatar" class="avatar" />
        <div class="user-details">
            <span class="user-name">John Doe</span>
            @BasicStatus(StatusProps{
                Level:   "success",
                Text:    "Online",
                ShowDot: true,
                Size:    "sm",
            })
        </div>
    </div>
}
```

### Process Status with Progress
```go
templ FileUploadStatus() {
    @ProgressStatus(StatusProps{
        Level:           "pending",
        Text:            "Uploading document.pdf",
        Progress:        0.67,
        ShowProgressBar: true,
        EstimatedTime:   "2 minutes",
        Animated:        true,
    })
}
```

### Alert Status Card
```go
templ MaintenanceAlert() {
    @StatusCard(StatusProps{
        Level:       "warning",
        Title:       "Scheduled Maintenance",
        Text:        "System will be offline for 30 minutes",
        Description: "We're updating our servers to improve performance. Service will resume at 3:00 PM EST.",
        ShowIcon:    true,
        Dismissible: true,
        Actions: []StatusAction{
            {Text: "Learn More", Variant: "secondary", OnClick: "showMaintenanceDetails"},
            {Text: "Dismiss", Variant: "primary", OnClick: "dismissAlert"},
        },
    })
}
```

## 🔗 Related Components

- **[Badge](../badge/)** - Status indicators and labels
- **[Icon](../icon/)** - Visual symbols for status
- **[Alert](../../molecules/alert/)** - Enhanced status messages
- **[Progress](../progress/)** - Progress tracking elements

---

**Component Status**: ✅ Production Ready  
**Schema Reference**: `StatusSchema.json`  
**CSS Classes**: `.status`, `.status-{level}`, `.status-{size}`, `.status-{variant}`  
**Accessibility**: WCAG 2.1 AA Compliant