# Video Component

**FILE PURPOSE**: Video player control specifications for multimedia content presentation and training delivery  
**SCOPE**: All video player types, frame navigation, live streaming, and media handling patterns  
**TARGET AUDIENCE**: Developers implementing video content delivery, training systems, and multimedia presentation

## 📋 Component Overview

The Video component provides comprehensive video playback capabilities for applications requiring rich media content delivery. It supports multiple video formats, frame-by-frame navigation, live streaming, configurable aspect ratios, playback speed control, and advanced video features while maintaining accessibility and consistent user experience across different video-related interfaces.

### Schema Reference
- **Primary Schema**: `VideoSchema.json`
- **Related Schemas**: `SchemaUrlPath.json`, `SchemaExpression.json`, `SchemaClassName.json`
- **Base Interface**: Media player control for video content presentation

## 🎨 JSON Schema Configuration

The Video component is configured using JSON that conforms to the `VideoSchema.json`. The JSON configuration renders interactive video players with customizable controls, frame navigation, live streaming support, and user-friendly media interfaces.

## Basic Usage

```json
{
    "type": "video",
    "src": "/videos/training-module.mp4",
    "poster": "/images/video-thumbnail.jpg",
    "aspectRatio": "16:9"
}
```

This JSON configuration renders to a Templ component with styled video player:

```go
// Generated from JSON schema
type VideoProps struct {
    Type            string                 `json:"type"`
    Src             interface{}            `json:"src"`
    Poster          interface{}            `json:"poster"`
    AspectRatio     string                 `json:"aspectRatio"`
    AutoPlay        bool                   `json:"autoPlay"`
    Loop            bool                   `json:"loop"`
    Muted           bool                   `json:"muted"`
    IsLive          bool                   `json:"isLive"`
    Rates           []float64              `json:"rates"`
    Frames          map[string]string      `json:"frames"`
    ColumnsCount    int                    `json:"columnsCount"`
    // ... additional props
}
```

## Video Player Variants

### Basic Video Player
**Purpose**: Simple video playback for content presentation

**JSON Configuration:**
```json
{
    "type": "video",
    "src": "/assets/videos/company-intro.mp4",
    "poster": "/assets/images/company-intro-thumb.jpg",
    "aspectRatio": "16:9",
    "autoPlay": false,
    "loop": false,
    "muted": false,
    "className": "video-player-basic"
}
```

### Training Video with Frame Navigation
**Purpose**: Educational content with precise navigation and speed controls

**JSON Configuration:**
```json
{
    "type": "video",
    "src": "${course.videoUrl}",
    "poster": "${course.thumbnailUrl}",
    "aspectRatio": "16:9",
    "autoPlay": false,
    "loop": false,
    "rates": [0.5, 0.75, 1.0, 1.25, 1.5, 2.0],
    "frames": {
        "00:30": "/frames/intro.jpg",
        "02:15": "/frames/chapter1.jpg",
        "05:45": "/frames/chapter2.jpg",
        "08:30": "/frames/summary.jpg"
    },
    "columnsCount": 4,
    "jumpFrame": true,
    "jumpBufferDuration": 2,
    "stopOnNextFrame": true,
    "className": "video-player-training",
    "playerClassName": "training-video-controls",
    "framesClassName": "training-frame-nav"
}
```

### Live Streaming Player
**Purpose**: Real-time video streaming for events and meetings

**JSON Configuration:**
```json
{
    "type": "video",
    "src": "${stream.hlsUrl}",
    "poster": "/images/live-placeholder.jpg",
    "aspectRatio": "16:9",
    "autoPlay": true,
    "loop": false,
    "muted": true,
    "isLive": true,
    "videoType": "application/x-mpegURL",
    "className": "video-player-live",
    "playerClassName": "live-stream-controls",
    "onEvent": {
        "play": {
            "actions": [
                {
                    "actionType": "toast",
                    "args": {
                        "msg": "Joining live stream: ${stream.title}"
                    }
                }
            ]
        },
        "error": {
            "actions": [
                {
                    "actionType": "toast",
                    "args": {
                        "msg": "Connection lost. Attempting to reconnect...",
                        "level": "warning"
                    }
                }
            ]
        }
    }
}
```

### Meeting Recording Player
**Purpose**: Corporate meeting recordings with presentation features

**JSON Configuration:**
```json
{
    "type": "video",
    "src": "${meeting.recordingUrl}",
    "poster": "${meeting.thumbnailUrl}",
    "aspectRatio": "16:9",
    "autoPlay": false,
    "loop": false,
    "muted": false,
    "rates": [0.75, 1.0, 1.25, 1.5, 2.0],
    "splitPoster": true,
    "className": "video-player-meeting",
    "style": {
        "maxWidth": "100%",
        "margin": "1rem 0",
        "border": "1px solid #e5e7eb",
        "borderRadius": "8px"
    },
    "onEvent": {
        "loadstart": {
            "actions": [
                {
                    "actionType": "setValue",
                    "componentId": "loading_status",
                    "args": { "value": "Loading meeting recording..." }
                }
            ]
        },
        "canplay": {
            "actions": [
                {
                    "actionType": "setValue",
                    "componentId": "loading_status", 
                    "args": { "value": "Ready to play" }
                }
            ]
        },
        "ended": {
            "actions": [
                {
                    "actionType": "ajax",
                    "api": {
                        "url": "/api/meetings/track-completion",
                        "method": "POST",
                        "data": {
                            "meetingId": "${meeting.id}",
                            "completedAt": "${NOW()}"
                        }
                    }
                }
            ]
        }
    }
}
```

## Video Configuration Options

### Aspect Ratios
- **`auto`**: Maintain original video dimensions
- **`4:3`**: Traditional television format
- **`16:9`**: Widescreen format (default for modern content)

### Video Types
Common video types for the `videoType` property:
- `video/mp4` - Standard MP4 format
- `video/webm` - WebM format
- `video/x-flv` - Flash video format
- `application/x-mpegURL` - HLS streaming format

### Frame Navigation
```json
{
    "frames": {
        "00:30": "/thumbnails/intro.jpg",
        "02:15": "/thumbnails/chapter1.jpg",
        "05:45": "/thumbnails/chapter2.jpg"
    },
    "columnsCount": 3,
    "jumpFrame": true,
    "jumpBufferDuration": 2,
    "stopOnNextFrame": false
}
```

## Go Type Definitions

```go
// VideoComponent represents the video player configuration
type VideoComponent struct {
    // Base component properties
    Type      string                 `json:"type"`
    ID        string                 `json:"id,omitempty"`
    ClassName interface{}            `json:"className,omitempty"`
    Style     map[string]interface{} `json:"style,omitempty"`
    
    // Visibility and state controls
    Hidden     bool        `json:"hidden,omitempty"`
    HiddenOn   interface{} `json:"hiddenOn,omitempty"`
    Visible    bool        `json:"visible,omitempty"`
    VisibleOn  interface{} `json:"visibleOn,omitempty"`
    Disabled   bool        `json:"disabled,omitempty"`
    DisabledOn interface{} `json:"disabledOn,omitempty"`
    
    // Video-specific properties
    Src                 interface{}       `json:"src"`                           // Video source URL
    Poster              interface{}       `json:"poster,omitempty"`              // Thumbnail image
    AspectRatio         string            `json:"aspectRatio,omitempty"`         // Video dimensions
    AutoPlay            bool              `json:"autoPlay,omitempty"`            // Auto-start playback
    Loop                bool              `json:"loop,omitempty"`                // Enable looping
    Muted               bool              `json:"muted,omitempty"`               // Start muted
    IsLive              bool              `json:"isLive,omitempty"`              // Live streaming mode
    VideoType           string            `json:"videoType,omitempty"`           // MIME type
    Rates               []float64         `json:"rates,omitempty"`               // Playback speeds
    
    // Frame navigation
    Frames              map[string]string `json:"frames,omitempty"`              // Time -> thumbnail mapping
    ColumnsCount        int               `json:"columnsCount,omitempty"`        // Frame grid columns
    JumpFrame           bool              `json:"jumpFrame,omitempty"`           // Enable frame jumping
    JumpBufferDuration  float64           `json:"jumpBufferDuration,omitempty"`  // Jump buffer seconds
    StopOnNextFrame     bool              `json:"stopOnNextFrame,omitempty"`     // Pause at next frame
    
    // Presentation options
    SplitPoster         bool        `json:"splitPoster,omitempty"`       // Separate poster display
    PlayerClassName     interface{} `json:"playerClassName,omitempty"`   // Player control styles
    FramesClassName     interface{} `json:"framesClassName,omitempty"`   // Frame navigation styles
    
    // Event handling
    OnEvent map[string]EventConfig `json:"onEvent,omitempty"`
    
    // Testing and debugging
    TestID       string      `json:"testid,omitempty"`
    TestIDBuilder interface{} `json:"testIdBuilder,omitempty"`
}

// Video aspect ratio constants
const (
    AspectRatioAuto = "auto"
    AspectRatio4x3  = "4:3"
    AspectRatio16x9 = "16:9"
)

// Common video types
const (
    VideoTypeMP4  = "video/mp4"
    VideoTypeWebM = "video/webm"
    VideoTypeFLV  = "video/x-flv"
    VideoTypeHLS  = "application/x-mpegURL"
)

// Standard playback rates
var StandardVideoRates = []float64{0.5, 0.75, 1.0, 1.25, 1.5, 2.0}
var DetailedVideoRates = []float64{0.25, 0.5, 0.75, 1.0, 1.25, 1.5, 1.75, 2.0}
```

## Templ Implementation

```templ
package molecules

import (
    "strconv"
    "strings"
    "fmt"
)

// VideoPlayer renders a video player component
templ VideoPlayer(props VideoComponent) {
    <div 
        id={ props.ID }
        class={ getVideoClassName(props) }
        style={ templ.Attributes(props.Style) }
        if props.Hidden { hidden }
        data-testid={ props.TestID }
    >
        if props.SplitPoster && props.Poster != nil {
            <div class="video-poster-section mb-4">
                <img 
                    src={ props.Poster.(string) }
                    alt="Video thumbnail"
                    class="w-full h-auto rounded-lg shadow-sm"
                />
            </div>
        }
        
        <div class={ getVideoContainerClass(props) }>
            <video
                class={ getVideoPlayerClass(props) }
                if props.AutoPlay { autoplay }
                if props.Loop { loop }
                if props.Muted { muted }
                if props.Poster != nil && !props.SplitPoster { 
                    poster={ props.Poster.(string) }
                }
                controls
                preload="metadata"
                if props.IsLive { 
                    data-live="true"
                }
                if props.VideoType != "" {
                    data-video-type={ props.VideoType }
                }
            >
                <source src={ props.Src.(string) } if props.VideoType != "" { type={ props.VideoType } }/>
                if props.VideoType == "" {
                    <source src={ props.Src.(string) } type="video/mp4"/>
                    <source src={ props.Src.(string) } type="video/webm"/>
                }
                Your browser does not support the video element.
            </video>
            
            if props.IsLive {
                <div class="live-indicator">
                    <span class="live-dot"></span>
                    LIVE
                </div>
            }
        </div>
        
        if len(props.Rates) > 0 {
            <div class="video-controls mt-3">
                <div class="flex items-center space-x-4">
                    <div class="playback-speed-control">
                        <label class="block text-sm font-medium text-gray-700 mb-1">
                            Speed
                        </label>
                        <select 
                            class="form-select text-sm px-2 py-1 border border-gray-300 bg-white rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                            onchange="setVideoPlaybackRate(this)"
                        >
                            for _, rate := range props.Rates {
                                <option 
                                    value={ strconv.FormatFloat(rate, 'f', -1, 64) }
                                    if rate == 1.0 { selected }
                                >
                                    { strconv.FormatFloat(rate, 'f', -1, 64) }x
                                </option>
                            }
                        </select>
                    </div>
                </div>
            </div>
        }
        
        if len(props.Frames) > 0 {
            <div class={ getFramesClassName(props) }>
                <h4 class="text-sm font-medium text-gray-700 mb-2">Chapters</h4>
                <div class={ getFrameGridClass(props) }>
                    for timestamp, thumbnailUrl := range props.Frames {
                        <div 
                            class="frame-item cursor-pointer rounded-lg overflow-hidden shadow-sm hover:shadow-md transition-shadow"
                            onclick={ getFrameClickHandler(timestamp, props.JumpFrame, props.JumpBufferDuration) }
                        >
                            <img 
                                src={ thumbnailUrl }
                                alt={ fmt.Sprintf("Frame at %s", timestamp) }
                                class="w-full h-auto"
                            />
                            <div class="frame-timestamp text-xs text-center py-1 bg-gray-50">
                                { timestamp }
                            </div>
                        </div>
                    }
                </div>
            </div>
        }
    </div>
}

// Helper functions for CSS classes
func getVideoClassName(props VideoComponent) string {
    classes := []string{"video-component"}
    
    if props.ClassName != nil {
        switch className := props.ClassName.(type) {
        case string:
            if className != "" {
                classes = append(classes, className)
            }
        case []string:
            classes = append(classes, className...)
        }
    }
    
    return strings.Join(classes, " ")
}

func getVideoContainerClass(props VideoComponent) string {
    classes := []string{"video-container", "relative"}
    
    switch props.AspectRatio {
    case "4:3":
        classes = append(classes, "aspect-4-3")
    case "16:9":
        classes = append(classes, "aspect-16-9")
    default:
        classes = append(classes, "aspect-auto")
    }
    
    return strings.Join(classes, " ")
}

func getVideoPlayerClass(props VideoComponent) string {
    classes := []string{"video-player", "w-full", "h-full", "rounded-lg"}
    
    if props.PlayerClassName != nil {
        switch className := props.PlayerClassName.(type) {
        case string:
            if className != "" {
                classes = append(classes, className)
            }
        case []string:
            classes = append(classes, className...)
        }
    }
    
    return strings.Join(classes, " ")
}

func getFramesClassName(props VideoComponent) string {
    classes := []string{"video-frames", "mt-4"}
    
    if props.FramesClassName != nil {
        switch className := props.FramesClassName.(type) {
        case string:
            if className != "" {
                classes = append(classes, className)
            }
        case []string:
            classes = append(classes, className...)
        }
    }
    
    return strings.Join(classes, " ")
}

func getFrameGridClass(props VideoComponent) string {
    columns := props.ColumnsCount
    if columns <= 0 {
        columns = 4 // Default
    }
    
    gridClass := fmt.Sprintf("grid gap-2 grid-cols-%d", columns)
    return gridClass
}

func getFrameClickHandler(timestamp string, jumpFrame bool, bufferDuration float64) string {
    if !jumpFrame {
        return ""
    }
    
    buffer := int(bufferDuration)
    if buffer <= 0 {
        buffer = 0
    }
    
    return fmt.Sprintf("jumpToVideoTime('%s', %d)", timestamp, buffer)
}
```

## CSS Styling

```css
/* Base video component styles */
.video-component {
    @apply w-full max-w-full;
}

/* Video container with aspect ratios */
.video-container {
    @apply relative w-full bg-black rounded-lg overflow-hidden;
}

.video-container.aspect-16-9 {
    aspect-ratio: 16 / 9;
}

.video-container.aspect-4-3 {
    aspect-ratio: 4 / 3;
}

.video-container.aspect-auto {
    @apply h-auto;
}

/* Video player styles */
.video-player {
    @apply object-cover;
    outline: none;
}

.video-player:focus {
    @apply ring-2 ring-blue-500 ring-opacity-50;
}

/* Live indicator */
.live-indicator {
    @apply absolute top-4 right-4 bg-red-600 text-white px-2 py-1 rounded-md text-xs font-medium flex items-center space-x-1;
}

.live-dot {
    @apply w-2 h-2 bg-white rounded-full animate-pulse;
}

/* Video controls */
.video-controls {
    @apply flex flex-wrap items-center space-x-4;
}

.playback-speed-control select {
    min-width: 80px;
}

/* Frame navigation */
.video-frames {
    @apply border-t border-gray-200 pt-4;
}

.frame-item {
    @apply bg-white border border-gray-200;
    transition: all 0.2s ease;
}

.frame-item:hover {
    @apply border-blue-300 transform scale-105;
}

.frame-timestamp {
    @apply text-gray-600 font-mono;
}

/* Video player variants */
.video-player-basic {
    @apply border border-gray-200 rounded-lg p-4 bg-white shadow-sm;
}

.video-player-training {
    @apply border border-blue-200 rounded-lg p-4 bg-blue-50 shadow-sm;
}

.video-player-training .video-frames {
    @apply border-blue-200;
}

.video-player-live {
    @apply border border-red-200 rounded-lg p-4 bg-red-50 shadow-md;
}

.video-player-meeting {
    @apply border border-gray-300 rounded-lg p-4 bg-gray-50 shadow-md;
}

/* Responsive grid adjustments */
@media (max-width: 640px) {
    .grid-cols-4 {
        @apply grid-cols-2;
    }
    
    .grid-cols-3 {
        @apply grid-cols-2;
    }
    
    .video-controls {
        @apply flex-col space-x-0 space-y-2;
    }
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
    .video-component {
        @apply text-gray-100;
    }
    
    .video-player-basic {
        @apply bg-gray-800 border-gray-600;
    }
    
    .frame-item {
        @apply bg-gray-800 border-gray-600;
    }
    
    .frame-timestamp {
        @apply bg-gray-700 text-gray-300;
    }
}

/* Loading and error states */
.video-component.loading .video-container {
    @apply bg-gray-200 animate-pulse;
}

.video-component.error .video-container {
    @apply bg-red-50 border-red-200;
}

/* Accessibility enhancements */
.video-component:focus-within {
    @apply ring-2 ring-blue-500 ring-opacity-50;
}

/* Poster section */
.video-poster-section img {
    @apply border border-gray-200;
}
```

## JavaScript Helper Functions

```javascript
// Video playback rate control
function setVideoPlaybackRate(selectElement) {
    const video = selectElement.closest('.video-component').querySelector('video');
    const rate = parseFloat(selectElement.value);
    if (video) {
        video.playbackRate = rate;
    }
}

// Frame jumping functionality
function jumpToVideoTime(timestamp, bufferSeconds = 0) {
    const video = event.target.closest('.video-component').querySelector('video');
    if (video) {
        const timeInSeconds = parseTimestamp(timestamp) - bufferSeconds;
        video.currentTime = Math.max(0, timeInSeconds);
        
        // Highlight the current frame
        highlightCurrentFrame(timestamp);
    }
}

// Parse timestamp string to seconds
function parseTimestamp(timestamp) {
    const parts = timestamp.split(':');
    let seconds = 0;
    
    if (parts.length === 3) {
        // HH:MM:SS
        seconds = parseInt(parts[0]) * 3600 + parseInt(parts[1]) * 60 + parseInt(parts[2]);
    } else if (parts.length === 2) {
        // MM:SS
        seconds = parseInt(parts[0]) * 60 + parseInt(parts[1]);
    }
    
    return seconds;
}

// Highlight current frame
function highlightCurrentFrame(timestamp) {
    const frameItems = document.querySelectorAll('.frame-item');
    frameItems.forEach(item => item.classList.remove('active'));
    
    const currentFrame = document.querySelector(`[onclick*="${timestamp}"]`);
    if (currentFrame) {
        currentFrame.classList.add('active');
    }
}
```

## Accessibility Features

- **Keyboard Navigation**: Full keyboard control for video playback
- **Screen Reader Support**: Proper ARIA labels and semantic HTML
- **Caption Support**: Compatible with subtitle and caption tracks
- **Focus Management**: Clear focus indicators and logical tab order
- **High Contrast**: Compatible with high contrast mode
- **Reduced Motion**: Respects user's motion preferences

## Testing Strategy

### Unit Tests
```go
func TestVideoComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    VideoComponent
        expected string
    }{
        {
            name: "basic video player",
            props: VideoComponent{
                Type: "video",
                Src:  "/test.mp4",
                AspectRatio: "16:9",
            },
            expected: "video player with 16:9 aspect ratio",
        },
        {
            name: "training video with frames",
            props: VideoComponent{
                Type: "video",
                Src:  "/training.mp4",
                Frames: map[string]string{
                    "01:00": "/thumb1.jpg",
                    "02:00": "/thumb2.jpg",
                },
                ColumnsCount: 2,
            },
            expected: "video player with frame navigation",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test video component rendering
            result := renderVideoComponent(tt.props)
            assert.Contains(t, result, "video-component")
            
            if len(tt.props.Frames) > 0 {
                assert.Contains(t, result, "video-frames")
            }
        })
    }
}
```

### Integration Tests
- Test video loading and playback functionality
- Verify frame navigation and jumping
- Test live streaming capabilities
- Validate accessibility compliance
- Test responsive behavior across devices

## Real-World ERP Use Cases

### 1. Employee Training Platform
```json
{
    "type": "video",
    "src": "${course.videoUrl}",
    "poster": "${course.thumbnailUrl}",
    "aspectRatio": "16:9",
    "rates": [0.5, 0.75, 1.0, 1.25, 1.5],
    "frames": "${course.chapters}",
    "columnsCount": 4,
    "jumpFrame": true,
    "stopOnNextFrame": true,
    "onEvent": {
        "ended": {
            "actions": [
                {
                    "actionType": "ajax",
                    "api": {
                        "url": "/api/training/complete",
                        "method": "POST",
                        "data": {
                            "courseId": "${course.id}",
                            "completedAt": "${NOW()}",
                            "watchTime": "${video.currentTime}"
                        }
                    }
                },
                {
                    "actionType": "toast",
                    "args": {
                        "msg": "Training completed successfully!",
                        "level": "success"
                    }
                }
            ]
        }
    }
}
```

### 2. Product Demo Presentations
```json
{
    "type": "video",
    "src": "/videos/product-demos/${product.id}.mp4",
    "poster": "/images/product-thumbnails/${product.id}.jpg",
    "aspectRatio": "16:9",
    "autoPlay": false,
    "loop": false,
    "splitPoster": true,
    "rates": [0.75, 1.0, 1.25, 1.5],
    "className": "product-demo-player",
    "onEvent": {
        "play": {
            "actions": [
                {
                    "actionType": "ajax",
                    "api": {
                        "url": "/api/analytics/video-play",
                        "method": "POST",
                        "data": {
                            "productId": "${product.id}",
                            "userId": "${user.id}",
                            "timestamp": "${NOW()}"
                        }
                    }
                }
            ]
        }
    }
}
```

### 3. Customer Support Video Library
```json
{
    "type": "video",
    "src": "/api/support/videos/${ticket.category}/${solution.id}",
    "poster": "/images/support-thumbnails/default.jpg",
    "aspectRatio": "16:9",
    "autoPlay": false,
    "rates": [0.75, 1.0, 1.25, 1.5, 2.0],
    "className": "support-video-player",
    "onEvent": {
        "ended": {
            "actions": [
                {
                    "actionType": "dialog",
                    "dialog": {
                        "title": "Was this helpful?",
                        "body": "Did this video help resolve your issue?",
                        "actions": [
                            {
                                "label": "Yes, it helped",
                                "actionType": "ajax",
                                "api": "/api/support/feedback/helpful"
                            },
                            {
                                "label": "No, I need more help",
                                "actionType": "url",
                                "url": "/support/contact"
                            }
                        ]
                    }
                }
            ]
        }
    }
}
```

### 4. Live Company Events
```json
{
    "type": "video",
    "src": "${event.streamUrl}",
    "aspectRatio": "16:9",
    "autoPlay": true,
    "muted": true,
    "isLive": true,
    "videoType": "application/x-mpegURL",
    "className": "live-event-player",
    "onEvent": {
        "loadstart": {
            "actions": [
                {
                    "actionType": "toast",
                    "args": {
                        "msg": "Connecting to live stream...",
                        "level": "info"
                    }
                }
            ]
        },
        "canplay": {
            "actions": [
                {
                    "actionType": "toast",
                    "args": {
                        "msg": "Connected! You are now viewing the live event.",
                        "level": "success"
                    }
                }
            ]
        },
        "error": {
            "actions": [
                {
                    "actionType": "toast",
                    "args": {
                        "msg": "Stream connection lost. Attempting to reconnect...",
                        "level": "warning"
                    }
                }
            ]
        }
    }
}
```

The Video component provides comprehensive multimedia presentation capabilities essential for modern ERP systems, supporting training delivery, product demonstrations, customer support, and live event streaming with full accessibility and professional styling.