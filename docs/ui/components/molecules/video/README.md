# Video Component

## Overview

The Video component provides a comprehensive video player with advanced features including frame navigation, live streaming support, customizable aspect ratios, playback rate control, and interactive frame thumbnails.

## Basic Usage

### Simple Video Player
```json
{
  "type": "video",
  "src": "/videos/demo.mp4",
  "poster": "/images/video-poster.jpg"
}
```

### Video with Custom Controls
```json
{
  "type": "video",
  "src": "/videos/tutorial.mp4",
  "poster": "/images/tutorial-poster.jpg",
  "autoPlay": false,
  "loop": false,
  "muted": false,
  "aspectRatio": "16:9"
}
```

### Video with Frame Navigation
```json
{
  "type": "video",
  "src": "/videos/presentation.mp4",
  "poster": "/images/presentation-cover.jpg",
  "frames": {
    "00:30": "/images/frames/intro.jpg",
    "02:15": "/images/frames/overview.jpg",
    "05:45": "/images/frames/features.jpg",
    "08:20": "/images/frames/conclusion.jpg"
  },
  "columnsCount": 4,
  "jumpFrame": true
}
```

## Complete Form Examples

### Training Video Library
```json
{
  "type": "page",
  "title": "Training Library",
  "body": [
    {
      "type": "tabs",
      "tabs": [
        {
          "title": "Onboarding",
          "body": [
            {
              "type": "cards",
              "api": "/api/training/onboarding",
              "card": {
                "header": {
                  "title": "${title}",
                  "subTitle": "Duration: ${duration} • Level: ${level}"
                },
                "body": [
                  {
                    "type": "video",
                    "src": "${video_url}",
                    "poster": "${thumbnail}",
                    "aspectRatio": "16:9",
                    "rates": [0.75, 1.0, 1.25, 1.5],
                    "frames": "${chapter_frames}",
                    "columnsCount": 3,
                    "jumpFrame": true,
                    "onEvent": {
                      "ended": {
                        "actions": [
                          {
                            "actionType": "ajax",
                            "api": "/api/training/complete/${video_id}"
                          },
                          {
                            "actionType": "toast",
                            "msg": "Training module completed!"
                          }
                        ]
                      }
                    }
                  },
                  {
                    "type": "tpl",
                    "tpl": "<div class='mt-4'><h4 class='font-semibold'>Description:</h4><p class='text-gray-600'>${description}</p></div>"
                  }
                ],
                "actions": [
                  {
                    "type": "button",
                    "label": "Mark Complete",
                    "level": "primary",
                    "actionType": "ajax",
                    "api": "/api/training/complete/${video_id}",
                    "visibleOn": "${!completed}"
                  },
                  {
                    "type": "button",
                    "label": "Download Certificate",
                    "level": "success", 
                    "icon": "fa fa-download",
                    "actionType": "download",
                    "api": "/api/training/certificate/${video_id}",
                    "visibleOn": "${completed}"
                  }
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}
```

### Product Demo Showcase
```json
{
  "type": "page",
  "title": "Product Demos",
  "body": [
    {
      "type": "grid",
      "columns": [
        {
          "md": 8,
          "body": [
            {
              "type": "video",
              "src": "${selected_demo.video_url}",
              "poster": "${selected_demo.poster}",
              "aspectRatio": "16:9",
              "autoPlay": false,
              "rates": [0.5, 0.75, 1.0, 1.25, 1.5, 2.0],
              "frames": "${selected_demo.chapters}",
              "columnsCount": 2,
              "jumpFrame": true,
              "stopOnNextFrame": false,
              "playerClassName": "rounded-lg shadow-lg",
              "framesClassName": "mt-4"
            }
          ]
        },
        {
          "md": 4,
          "body": [
            {
              "type": "card",
              "header": {
                "title": "Demo Library",
                "subTitle": "Select a demo to watch"
              },
              "body": [
                {
                  "type": "list",
                  "source": "${demo_library}",
                  "listItem": {
                    "body": [
                      {
                        "type": "container",
                        "className": "p-3 border rounded hover:bg-gray-50 cursor-pointer ${selected_demo.id === id ? 'bg-blue-50 border-blue-200' : ''}",
                        "body": [
                          {
                            "type": "container",
                            "className": "flex items-center space-x-3",
                            "body": [
                              {
                                "type": "image",
                                "src": "${thumbnail}",
                                "className": "w-16 h-12 rounded object-cover"
                              },
                              {
                                "type": "container",
                                "body": [
                                  {
                                    "type": "tpl",
                                    "tpl": "<h4 class='font-medium'>${title}</h4>"
                                  },
                                  {
                                    "type": "tpl",
                                    "tpl": "<p class='text-sm text-gray-500'>${duration}</p>"
                                  }
                                ]
                              }
                            ]
                          }
                        ],
                        "onEvent": {
                          "click": {
                            "actions": [
                              {
                                "actionType": "setValue",
                                "componentId": "selected_demo",
                                "value": "${.}"
                              }
                            ]
                          }
                        }
                      }
                    ]
                  }
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

### Live Streaming Interface
```json
{
  "type": "page",
  "title": "Live Events",
  "body": [
    {
      "type": "card",
      "header": {
        "title": "🔴 Live: ${event_title}",
        "subTitle": "Started ${start_time} • ${viewer_count} viewers"
      },
      "body": [
        {
          "type": "video",
          "src": "${live_stream_url}",
          "poster": "${event_poster}",
          "isLive": true,
          "autoPlay": true,
          "muted": false,
          "aspectRatio": "16:9",
          "videoType": "application/x-mpegURL",
          "playerClassName": "w-full rounded-lg",
          "onEvent": {
            "play": {
              "actions": [
                {
                  "actionType": "ajax",
                  "api": "/api/events/${event_id}/join"
                }
              ]
            },
            "pause": {
              "actions": [
                {
                  "actionType": "ajax",
                  "api": "/api/events/${event_id}/pause"
                }
              ]
            }
          }
        },
        {
          "type": "grid",
          "className": "mt-6",
          "columns": [
            {
              "md": 8,
              "body": [
                {
                  "type": "tpl",
                  "tpl": "<h3 class='text-xl font-bold mb-2'>${event_title}</h3>"
                },
                {
                  "type": "tpl",
                  "tpl": "<p class='text-gray-600 mb-4'>${event_description}</p>"
                },
                {
                  "type": "tpl",
                  "tpl": "<div class='flex items-center space-x-4 text-sm text-gray-500'><span>👥 ${viewer_count} watching</span><span>⏱️ Started ${start_time}</span></div>"
                }
              ]
            },
            {
              "md": 4,
              "body": [
                {
                  "type": "container",
                  "className": "space-y-2",
                  "body": [
                    {
                      "type": "button",
                      "label": "Share Stream",
                      "level": "default",
                      "block": true,
                      "icon": "fa fa-share",
                      "actionType": "copy",
                      "content": "${window.location.href}"
                    },
                    {
                      "type": "button",
                      "label": "Join Chat",
                      "level": "primary",
                      "block": true,
                      "icon": "fa fa-comments",
                      "actionType": "drawer",
                      "drawer": {
                        "title": "Live Chat",
                        "body": {
                          "type": "iframe",
                          "src": "/chat/${event_id}"
                        }
                      }
                    }
                  ]
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

### Video Tutorial with Interactive Elements
```json
{
  "type": "form",
  "title": "Interactive Tutorial",
  "body": [
    {
      "type": "video",
      "src": "/videos/interactive-tutorial.mp4",
      "poster": "/images/tutorial-cover.jpg",
      "aspectRatio": "16:9",
      "frames": {
        "01:30": "/images/frames/step1.jpg",
        "03:45": "/images/frames/step2.jpg", 
        "06:20": "/images/frames/step3.jpg",
        "09:10": "/images/frames/step4.jpg",
        "12:30": "/images/frames/conclusion.jpg"
      },
      "columnsCount": 5,
      "jumpFrame": true,
      "stopOnNextFrame": true,
      "jumpBufferDuration": 2,
      "onEvent": {
        "timeupdate": {
          "actions": [
            {
              "actionType": "setValue",
              "componentId": "current_progress",
              "value": "${Math.round((currentTime / duration) * 100)}"
            }
          ]
        },
        "pause": {
          "actions": [
            {
              "actionType": "setValue",
              "componentId": "paused_at",
              "value": "${currentTime}"
            }
          ]
        }
      }
    },
    {
      "type": "container",
      "className": "mt-6 p-4 bg-gray-50 rounded-lg",
      "body": [
        {
          "type": "tpl",
          "tpl": "<h4 class='font-semibold mb-2'>Tutorial Progress</h4>"
        },
        {
          "type": "progress",
          "value": "${current_progress}",
          "className": "mb-4"
        },
        {
          "type": "tpl",
          "tpl": "<p class='text-sm text-gray-600'>Complete the tutorial to unlock the next module</p>"
        }
      ]
    },
    {
      "type": "container",
      "className": "mt-4",
      "visibleOn": "${current_progress >= 90}",
      "body": [
        {
          "type": "alert",
          "level": "success",
          "body": "🎉 Tutorial completed! You can now proceed to the assessment."
        },
        {
          "type": "button",
          "label": "Take Assessment",
          "level": "primary",
          "className": "mt-3",
          "actionType": "link",
          "link": "/assessment/tutorial-basic"
        }
      ]
    }
  ]
}
```

### Video Conference Recording Viewer
```json
{
  "type": "page",
  "title": "Meeting Recordings",
  "body": [
    {
      "type": "form",
      "body": [
        {
          "type": "input-date-range",
          "name": "date_range",
          "label": "Meeting Date Range"
        },
        {
          "type": "select",
          "name": "department",
          "label": "Department",
          "source": "/api/departments"
        }
      ]
    },
    {
      "type": "cards",
      "api": "/api/meetings/recordings?start=${date_range.start}&end=${date_range.end}&dept=${department}",
      "card": {
        "header": {
          "title": "${meeting_title}",
          "subTitle": "${meeting_date} • ${duration} • ${attendee_count} attendees"
        },
        "body": [
          {
            "type": "video",
            "src": "${recording_url}",
            "poster": "${thumbnail || '/images/default-meeting-thumb.jpg'}",
            "aspectRatio": "16:9",
            "rates": [1.0, 1.25, 1.5, 2.0],
            "frames": "${agenda_chapters}",
            "columnsCount": 3,
            "jumpFrame": true,
            "splitPoster": false
          },
          {
            "type": "container",
            "className": "mt-4",
            "body": [
              {
                "type": "tpl",
                "tpl": "<div class='grid grid-cols-2 gap-4 text-sm'><div><strong>Organizer:</strong> ${organizer}</div><div><strong>Duration:</strong> ${duration}</div><div><strong>Attendees:</strong> ${attendee_count}</div><div><strong>Department:</strong> ${department}</div></div>"
              }
            ]
          }
        ],
        "actions": [
          {
            "type": "button",
            "label": "Download",
            "level": "default",
            "icon": "fa fa-download",
            "actionType": "download",
            "api": "/api/meetings/download/${recording_id}"
          },
          {
            "type": "button",
            "label": "Share",
            "level": "default",
            "icon": "fa fa-share",
            "actionType": "copy",
            "content": "${share_url}"
          },
          {
            "type": "button",
            "label": "Transcript",
            "level": "link",
            "icon": "fa fa-file-text",
            "actionType": "dialog",
            "dialog": {
              "title": "Meeting Transcript",
              "body": {
                "type": "iframe",
                "src": "/transcripts/${recording_id}"
              }
            }
          }
        ]
      }
    }
  ]
}
```

## Property Reference

### Core Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `type` | `string` | - | **Required.** Must be `"video"` |
| `src` | `string` | - | **Required.** Video source URL or path |
| `poster` | `string` | - | Video poster/thumbnail image URL |

### Playback Control

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `autoPlay` | `boolean` | `false` | Start playing automatically |
| `loop` | `boolean` | `false` | Loop video when it ends |
| `muted` | `boolean` | `false` | Start with audio muted |
| `rates` | `array` | `[1.0]` | Available playback speed options |

### Display Configuration

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `aspectRatio` | `string` | `"auto"` | Video aspect ratio: `"auto"`, `"4:3"`, `"16:9"` |
| `splitPoster` | `boolean` | `false` | Display video and poster separately |
| `videoType` | `string` | - | Video MIME type (e.g., `"video/mp4"`) |

### Frame Navigation

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `frames` | `object` | - | Frame thumbnails with timestamps |
| `columnsCount` | `number` | `4` | Number of frame thumbnails per row |
| `jumpFrame` | `boolean` | `true` | Allow clicking frames to jump to time |
| `jumpBufferDuration` | `number` | `0` | Seconds to add when jumping to frame |
| `stopOnNextFrame` | `boolean` | `false` | Pause when reaching next frame marker |

### Live Streaming

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `isLive` | `boolean` | `false` | Mark as live stream content |

### Styling

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `className` | `string` | - | Additional CSS classes |
| `style` | `object` | - | Inline styles |
| `playerClassName` | `string` | - | CSS classes for video player |
| `framesClassName` | `string` | - | CSS classes for frame thumbnails |

### Common Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `id` | `string` | - | Unique component identifier |
| `testid` | `string` | - | Test automation identifier |
| `disabled` | `boolean` | `false` | Disable video player |
| `hidden` | `boolean` | `false` | Hide the component |
| `visible` | `boolean` | `true` | Component visibility |
| `static` | `boolean` | `false` | Static display mode |

## Event Handling

### Available Events

| Event | Description | Data |
|-------|-------------|------|
| `play` | Video starts playing | `{currentTime: number, duration: number}` |
| `pause` | Video is paused | `{currentTime: number, duration: number}` |
| `ended` | Video playback finished | `{duration: number}` |
| `timeupdate` | Playback position changed | `{currentTime: number, duration: number}` |
| `volumechange` | Volume level changed | `{volume: number}` |
| `ratechange` | Playback rate changed | `{playbackRate: number}` |
| `loadstart` | Video loading started | `{src: string}` |
| `loadeddata` | Video data loaded | `{duration: number, videoWidth: number, videoHeight: number}` |
| `error` | Video loading error | `{error: string}` |
| `fullscreenchange` | Fullscreen mode changed | `{fullscreen: boolean}` |
| `frameclick` | Frame thumbnail clicked | `{timestamp: string, frameUrl: string}` |

### Event Configuration Examples

```json
{
  "type": "video",
  "src": "/videos/training.mp4",
  "frames": {
    "02:30": "/frames/section1.jpg",
    "05:15": "/frames/section2.jpg"
  },
  "onEvent": {
    "play": {
      "actions": [
        {
          "actionType": "setValue",
          "componentId": "video_status",
          "value": "playing"
        }
      ]
    },
    "ended": {
      "actions": [
        {
          "actionType": "toast",
          "msg": "Video completed!"
        },
        {
          "actionType": "ajax",
          "api": "/api/videos/complete/${video_id}"
        }
      ]
    },
    "frameclick": {
      "actions": [
        {
          "actionType": "toast",
          "msg": "Jumped to ${timestamp}"
        }
      ]
    }
  }
}
```

## Frame Navigation Configuration

### Frame Object Format
```json
{
  "frames": {
    "00:30": "/images/frames/intro.jpg",
    "02:15": "/images/frames/chapter1.jpg",
    "05:45": "/images/frames/chapter2.jpg",
    "08:20": "/images/frames/conclusion.jpg"
  }
}
```

### Advanced Frame Configuration
```json
{
  "type": "video",
  "src": "/videos/course.mp4",
  "frames": "${video_chapters}",
  "columnsCount": "${device === 'mobile' ? 2 : 4}",
  "jumpFrame": true,
  "jumpBufferDuration": 1,
  "stopOnNextFrame": false,
  "framesClassName": "grid gap-2 mt-4"
}
```

## Video Format Support

### Supported Formats

| Format | Extension | Browser Support |
|--------|-----------|-----------------|
| MP4 | `.mp4` | Universal |
| WebM | `.webm` | Chrome, Firefox |
| OGV | `.ogv` | Firefox, Chrome |
| MOV | `.mov` | Safari, Chrome |
| AVI | `.avi` | Limited |

### Streaming Formats

| Format | Type | Use Case |
|--------|------|----------|
| HLS | `application/x-mpegURL` | Live streaming, adaptive bitrate |
| DASH | `application/dash+xml` | Adaptive streaming |
| RTMP | `rtmp://` | Live streaming (requires Flash) |

## Styling & Theming

### CSS Classes

- `.video-player` - Base video player container
- `.video-player__video` - Video element
- `.video-player__poster` - Poster image
- `.video-player__controls` - Player controls
- `.video-player__frames` - Frame thumbnails container
- `.video-player__frame` - Individual frame thumbnail
- `.video-player--live` - Live streaming indicator
- `.video-player--loading` - Loading state

### Custom Styling Examples

```json
{
  "type": "video",
  "src": "/videos/demo.mp4",
  "className": "custom-video-wrapper",
  "playerClassName": "rounded-xl shadow-2xl overflow-hidden",
  "framesClassName": "mt-6 grid grid-cols-4 gap-3",
  "style": {
    "maxWidth": "800px",
    "margin": "0 auto"
  }
}
```

### Responsive Configuration
```json
{
  "type": "video",
  "src": "/videos/responsive.mp4",
  "aspectRatio": "16:9",
  "className": "w-full max-w-4xl mx-auto",
  "columnsCount": "${device === 'mobile' ? 2 : device === 'tablet' ? 3 : 4}"
}
```

## Accessibility

### ARIA Support
- `role="application"` for video player
- `aria-label` for control buttons
- `aria-describedby` for video descriptions
- Keyboard navigation support

### Keyboard Navigation
- `Space` - Play/pause
- `Arrow Left/Right` - Skip backward/forward
- `Arrow Up/Down` - Volume control
- `M` - Mute/unmute
- `F` - Toggle fullscreen

### Best Practices
- Provide captions for accessibility
- Include video transcripts
- Support keyboard navigation
- Use sufficient color contrast for controls

## Integration Patterns

### With Form Data
```json
{
  "type": "video",
  "src": "${video_url}",
  "poster": "${thumbnail_url}",
  "visibleOn": "${video_url}",
  "frames": "${video_chapters}"
}
```

### With API Loading
```json
{
  "type": "video",
  "src": "/api/videos/${video_id}/stream",
  "poster": "/api/videos/${video_id}/thumbnail",
  "api": "/api/videos/${video_id}/details",
  "frames": "${video_frames}"
}
```

### With Conditional Display
```json
{
  "type": "video",
  "src": "${premium_content ? premium_video : free_video}",
  "poster": "${premium_content ? premium_poster : free_poster}",
  "aspectRatio": "16:9"
}
```

## Best Practices

### Performance
- Use appropriate video formats for target browsers
- Implement adaptive bitrate streaming for large videos
- Optimize poster images for quick loading
- Consider lazy loading for multiple videos

### User Experience
- Always provide poster images
- Include clear loading indicators
- Support both keyboard and mouse interaction
- Provide frame navigation for long videos

### Accessibility
- Always provide captions and transcripts
- Support keyboard navigation
- Include video descriptions
- Test with screen readers

## Go Type Definition

```go
// VideoComponent represents a video player component
type VideoComponent struct {
    BaseComponent
    
    // Core Properties
    Src    string `json:"src"`
    Poster string `json:"poster,omitempty"`
    
    // Playback Control
    AutoPlay bool      `json:"autoPlay,omitempty"`
    Loop     bool      `json:"loop,omitempty"`
    Muted    bool      `json:"muted,omitempty"`
    Rates    []float64 `json:"rates,omitempty"`
    
    // Display Configuration
    AspectRatio string `json:"aspectRatio,omitempty"`
    SplitPoster bool   `json:"splitPoster,omitempty"`
    VideoType   string `json:"videoType,omitempty"`
    
    // Frame Navigation
    Frames              map[string]string `json:"frames,omitempty"`
    ColumnsCount        int               `json:"columnsCount,omitempty"`
    JumpFrame           bool              `json:"jumpFrame,omitempty"`
    JumpBufferDuration  float64           `json:"jumpBufferDuration,omitempty"`
    StopOnNextFrame     bool              `json:"stopOnNextFrame,omitempty"`
    
    // Live Streaming
    IsLive bool `json:"isLive,omitempty"`
    
    // Styling
    ClassName       string `json:"className,omitempty"`
    Style           map[string]interface{} `json:"style,omitempty"`
    PlayerClassName string `json:"playerClassName,omitempty"`
    FramesClassName string `json:"framesClassName,omitempty"`
    
    // Event Configuration
    OnEvent map[string]EventConfig `json:"onEvent,omitempty"`
}

// VideoFactory creates Video components from JSON configuration
func VideoFactory(config map[string]interface{}) (*VideoComponent, error) {
    component := &VideoComponent{
        BaseComponent: BaseComponent{
            Type: "video",
        },
        AutoPlay:     false,
        Loop:         false,
        Muted:        false,
        AspectRatio:  "auto",
        SplitPoster:  false,
        ColumnsCount: 4,
        JumpFrame:    true,
        IsLive:       false,
        Rates:        []float64{1.0},
    }
    
    return component, mapConfig(config, component)
}

// Render generates the Templ template for the video player
func (c *VideoComponent) Render() templ.Component {
    return video.Video(video.VideoProps{
        Src:                 c.Src,
        Poster:              c.Poster,
        AutoPlay:            c.AutoPlay,
        Loop:                c.Loop,
        Muted:               c.Muted,
        Rates:               c.Rates,
        AspectRatio:         c.AspectRatio,
        SplitPoster:         c.SplitPoster,
        VideoType:           c.VideoType,
        Frames:              c.Frames,
        ColumnsCount:        c.ColumnsCount,
        JumpFrame:           c.JumpFrame,
        JumpBufferDuration:  c.JumpBufferDuration,
        StopOnNextFrame:     c.StopOnNextFrame,
        IsLive:              c.IsLive,
        ClassName:           c.ClassName,
        Style:               c.Style,
        PlayerClassName:     c.PlayerClassName,
        FramesClassName:     c.FramesClassName,
        OnEvent:             c.OnEvent,
    })
}

// ValidateVideoFormat checks if the video format is supported
func (c *VideoComponent) ValidateVideoFormat() error {
    if c.Src == "" {
        return fmt.Errorf("video source is required")
    }
    
    supportedFormats := []string{".mp4", ".webm", ".ogv", ".mov", ".avi"}
    for _, format := range supportedFormats {
        if strings.HasSuffix(strings.ToLower(c.Src), format) {
            return nil
        }
    }
    
    // Check for streaming formats
    streamingFormats := []string{"application/x-mpegURL", "application/dash+xml"}
    for _, format := range streamingFormats {
        if c.VideoType == format {
            return nil
        }
    }
    
    return fmt.Errorf("unsupported video format")
}
```

## Related Components

- **[Audio](../audio/)** - Audio player component
- **[Image](../atoms/image/)** - Image display component
- **[Carousel](../carousel/)** - Image/content carousel
- **[Button](../atoms/button/)** - Control button component
- **[Card](../card/)** - Container for video players