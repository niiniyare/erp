# Audio Component

## Overview

The Audio component provides a comprehensive audio player with customizable controls, playback rate options, loop functionality, and both inline and full-width display modes. It supports various audio formats and integrates seamlessly with form data.

## Basic Usage

### Simple Audio Player
```json
{
  "type": "audio",
  "src": "/audio/welcome.mp3"
}
```

### Audio with Custom Controls
```json
{
  "type": "audio",
  "src": "/audio/presentation.mp3",
  "controls": ["play", "time", "process", "volume"],
  "loop": false,
  "autoPlay": false
}
```

### Inline Audio Player
```json
{
  "type": "audio",
  "src": "/audio/notification.mp3",
  "inline": true,
  "controls": ["play", "volume"]
}
```

## Complete Form Examples

### Podcast Player Interface
```json
{
  "type": "page",
  "title": "Podcast Library",
  "body": [
    {
      "type": "cards",
      "api": "/api/podcasts",
      "card": {
        "header": {
          "title": "${title}",
          "subTitle": "Episode ${episode_number} • ${duration}"
        },
        "body": [
          {
            "type": "tpl",
            "tpl": "<p class='text-gray-600 mb-4'>${description}</p>"
          },
          {
            "type": "audio",
            "src": "${audio_url}",
            "controls": ["play", "time", "process", "volume", "rates"],
            "rates": [0.75, 1.0, 1.25, 1.5, 2.0],
            "className": "mb-4"
          },
          {
            "type": "container",
            "className": "flex justify-between text-sm text-gray-500",
            "body": [
              {
                "type": "tpl",
                "tpl": "<span>Published: ${published_date}</span>"
              },
              {
                "type": "tpl",
                "tpl": "<span>Downloads: ${download_count}</span>"
              }
            ]
          }
        ],
        "actions": [
          {
            "type": "button",
            "label": "Download",
            "level": "link",
            "icon": "fa fa-download",
            "actionType": "link",
            "link": "${download_url}"
          }
        ]
      }
    }
  ]
}
```

### Music Player Dashboard
```json
{
  "type": "page",
  "title": "Music Player",
  "body": [
    {
      "type": "grid",
      "columns": [
        {
          "md": 8,
          "body": [
            {
              "type": "card",
              "header": {
                "title": "Now Playing",
                "subTitle": "${current_track.artist} - ${current_track.album}"
              },
              "body": [
                {
                  "type": "container",
                  "className": "flex items-center space-x-4 mb-6",
                  "body": [
                    {
                      "type": "image",
                      "src": "${current_track.cover_art}",
                      "className": "w-20 h-20 rounded-lg"
                    },
                    {
                      "type": "container",
                      "body": [
                        {
                          "type": "tpl",
                          "tpl": "<h3 class='text-xl font-bold'>${current_track.title}</h3>"
                        },
                        {
                          "type": "tpl",
                          "tpl": "<p class='text-gray-600'>${current_track.artist}</p>"
                        }
                      ]
                    }
                  ]
                },
                {
                  "type": "audio",
                  "src": "${current_track.audio_url}",
                  "controls": ["play", "time", "process", "volume"],
                  "autoPlay": false,
                  "loop": false,
                  "onEvent": {
                    "ended": {
                      "actions": [
                        {
                          "actionType": "ajax",
                          "api": "/api/music/next-track"
                        }
                      ]
                    }
                  }
                }
              ]
            }
          ]
        },
        {
          "md": 4,
          "body": [
            {
              "type": "card",
              "header": {
                "title": "Playlist",
                "subTitle": "${playlist.length} tracks"
              },
              "body": [
                {
                  "type": "list",
                  "source": "${playlist}",
                  "listItem": {
                    "body": [
                      {
                        "type": "container",
                        "className": "flex items-center justify-between p-2 hover:bg-gray-50 rounded",
                        "body": [
                          {
                            "type": "container",
                            "body": [
                              {
                                "type": "tpl",
                                "tpl": "<div class='font-medium'>${title}</div>"
                              },
                              {
                                "type": "tpl",
                                "tpl": "<div class='text-sm text-gray-500'>${artist}</div>"
                              }
                            ]
                          },
                          {
                            "type": "audio",
                            "src": "${audio_url}",
                            "inline": true,
                            "controls": ["play"]
                          }
                        ]
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

### Language Learning Audio Lessons
```json
{
  "type": "form",
  "title": "Language Lessons",
  "body": [
    {
      "type": "select",
      "name": "lesson_level",
      "label": "Lesson Level",
      "options": [
        {"label": "Beginner", "value": "beginner"},
        {"label": "Intermediate", "value": "intermediate"},
        {"label": "Advanced", "value": "advanced"}
      ]
    },
    {
      "type": "cards",
      "api": "/api/lessons?level=${lesson_level}",
      "card": {
        "header": {
          "title": "Lesson ${lesson_number}: ${title}",
          "subTitle": "${difficulty_level} • ${estimated_time}"
        },
        "body": [
          {
            "type": "tpl",
            "tpl": "<div class='mb-4'><strong>Objective:</strong> ${objective}</div>"
          },
          {
            "type": "container",
            "className": "space-y-4",
            "body": [
              {
                "type": "tpl",
                "tpl": "<h4 class='font-semibold'>Listen and Repeat:</h4>"
              },
              {
                "type": "audio",
                "src": "${pronunciation_audio}",
                "controls": ["play", "time", "process"],
                "rates": [0.5, 0.75, 1.0],
                "loop": true
              },
              {
                "type": "tpl",
                "tpl": "<div class='bg-blue-50 p-3 rounded'><strong>Text:</strong> ${lesson_text}</div>"
              },
              {
                "type": "tpl",
                "tpl": "<h4 class='font-semibold mt-6'>Full Conversation:</h4>"
              },
              {
                "type": "audio",
                "src": "${conversation_audio}",
                "controls": ["play", "time", "process", "rates"],
                "rates": [0.75, 1.0, 1.25]
              }
            ]
          }
        ],
        "actions": [
          {
            "type": "button",
            "label": "Mark Complete",
            "level": "primary",
            "actionType": "ajax",
            "api": "/api/lessons/${lesson_id}/complete"
          }
        ]
      }
    }
  ]
}
```

### Audio Feedback Collection
```json
{
  "type": "form",
  "title": "Voice Feedback",
  "api": "/api/feedback/submit",
  "body": [
    {
      "type": "input-text",
      "name": "name",
      "label": "Your Name",
      "required": true
    },
    {
      "type": "select",
      "name": "feedback_type",
      "label": "Feedback Type",
      "options": [
        {"label": "Product Review", "value": "product"},
        {"label": "Service Experience", "value": "service"},
        {"label": "General Suggestion", "value": "suggestion"}
      ]
    },
    {
      "type": "textarea",
      "name": "written_feedback",
      "label": "Written Feedback",
      "placeholder": "Please share your thoughts..."
    },
    {
      "type": "static",
      "label": "Voice Recording",
      "value": "Record your feedback for a more personal touch"
    },
    {
      "type": "audio",
      "src": "${voice_recording_url}",
      "controls": ["play", "time", "process", "volume"],
      "visibleOn": "${voice_recording_url}",
      "className": "mt-2"
    },
    {
      "type": "button",
      "label": "${voice_recording_url ? 'Re-record' : 'Start Recording'}",
      "level": "info",
      "icon": "fa fa-microphone",
      "actionType": "dialog",
      "dialog": {
        "title": "Voice Recording",
        "body": {
          "type": "tpl",
          "tpl": "<div class='text-center p-8'><p>Voice recording functionality would be implemented here</p></div>"
        }
      }
    }
  ]
}
```

### Training Module with Audio Instructions
```json
{
  "type": "wizard",
  "title": "Safety Training Module",
  "steps": [
    {
      "title": "Introduction",
      "body": [
        {
          "type": "tpl",
          "tpl": "<h3>Welcome to Safety Training</h3><p>Please listen to the complete audio instruction before proceeding.</p>"
        },
        {
          "type": "audio",
          "src": "/audio/training/introduction.mp3",
          "controls": ["play", "time", "process"],
          "onEvent": {
            "ended": {
              "actions": [
                {
                  "actionType": "setValue",
                  "componentId": "intro_completed",
                  "value": true
                }
              ]
            }
          }
        },
        {
          "type": "checkbox",
          "name": "intro_completed",
          "label": "I have listened to the complete introduction",
          "disabled": true
        }
      ]
    },
    {
      "title": "Equipment Overview",
      "body": [
        {
          "type": "grid",
          "columns": [
            {
              "md": 6,
              "body": [
                {
                  "type": "image",
                  "src": "/images/training/equipment.jpg",
                  "className": "w-full rounded-lg"
                }
              ]
            },
            {
              "md": 6,
              "body": [
                {
                  "type": "audio",
                  "src": "/audio/training/equipment-overview.mp3",
                  "controls": ["play", "time", "process", "rates"],
                  "rates": [0.75, 1.0, 1.25]
                },
                {
                  "type": "tpl",
                  "tpl": "<div class='mt-4'><h4>Key Points:</h4><ul class='list-disc list-inside text-sm text-gray-600'><li>Proper handling procedures</li><li>Safety precautions</li><li>Maintenance requirements</li></ul></div>"
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

## Property Reference

### Core Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `type` | `string` | - | **Required.** Must be `"audio"` |
| `src` | `string` | - | **Required.** Audio source URL or path |

### Playback Control

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `autoPlay` | `boolean` | `false` | Start playing automatically |
| `loop` | `boolean` | `false` | Loop audio when it ends |
| `rates` | `array` | `[1.0]` | Available playback speed options |

### Interface Control

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `inline` | `boolean` | `false` | Display in compact inline mode |
| `controls` | `array` | `["play", "time", "process", "volume"]` | Visible control elements |

### Control Options

| Control | Description |
|---------|-------------|
| `"play"` | Play/pause button |
| `"time"` | Current time and duration display |
| `"process"` | Progress bar/scrubber |
| `"volume"` | Volume control slider |
| `"rates"` | Playback speed selector |

### Style & Appearance

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `className` | `string` | - | Additional CSS classes |
| `style` | `object` | - | Inline styles |

### Common Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `id` | `string` | - | Unique component identifier |
| `testid` | `string` | - | Test automation identifier |
| `disabled` | `boolean` | `false` | Disable audio player |
| `hidden` | `boolean` | `false` | Hide the component |
| `visible` | `boolean` | `true` | Component visibility |
| `static` | `boolean` | `false` | Static display mode |

## Event Handling

### Available Events

| Event | Description | Data |
|-------|-------------|------|
| `play` | Audio starts playing | `{currentTime: number, duration: number}` |
| `pause` | Audio is paused | `{currentTime: number, duration: number}` |
| `ended` | Audio playback finished | `{duration: number}` |
| `timeupdate` | Playback position changed | `{currentTime: number, duration: number}` |
| `volumechange` | Volume level changed | `{volume: number}` |
| `ratechange` | Playback rate changed | `{playbackRate: number}` |
| `loadstart` | Audio loading started | `{src: string}` |
| `loadeddata` | Audio data loaded | `{duration: number}` |
| `error` | Audio loading error | `{error: string}` |

### Event Configuration Examples

```json
{
  "type": "audio",
  "src": "/audio/lesson.mp3",
  "onEvent": {
    "play": {
      "actions": [
        {
          "actionType": "setValue",
          "componentId": "playback_status",
          "value": "playing"
        }
      ]
    },
    "ended": {
      "actions": [
        {
          "actionType": "setValue",
          "componentId": "lesson_completed",
          "value": true
        },
        {
          "actionType": "toast",
          "msg": "Audio lesson completed!"
        }
      ]
    },
    "timeupdate": {
      "actions": [
        {
          "actionType": "setValue",
          "componentId": "progress_percentage",
          "value": "${Math.round((currentTime / duration) * 100)}"
        }
      ]
    }
  }
}
```

## Audio Format Support

### Supported Formats

| Format | Extension | Browser Support |
|--------|-----------|-----------------|
| MP3 | `.mp3` | Universal |
| WAV | `.wav` | Universal |
| OGG | `.ogg` | Firefox, Chrome |
| AAC | `.aac` | Safari, Chrome |
| M4A | `.m4a` | Safari, Chrome |

### Format Selection Best Practices

```json
{
  "type": "audio",
  "src": "${browser === 'safari' ? '/audio/file.m4a' : '/audio/file.mp3'}",
  "controls": ["play", "time", "process"]
}
```

## Styling & Theming

### CSS Classes

- `.audio-player` - Base audio player container
- `.audio-player--inline` - Inline display mode
- `.audio-player__controls` - Controls container
- `.audio-player__play` - Play/pause button
- `.audio-player__time` - Time display
- `.audio-player__progress` - Progress bar
- `.audio-player__volume` - Volume control
- `.audio-player__rates` - Rate selector

### Custom Styling Examples

```json
{
  "type": "audio",
  "src": "/audio/background.mp3",
  "className": "custom-audio-player",
  "style": {
    "borderRadius": "8px",
    "backgroundColor": "#f8fafc",
    "border": "1px solid #e2e8f0",
    "padding": "12px"
  },
  "controls": ["play", "time", "volume"]
}
```

### Responsive Design
```json
{
  "type": "audio",
  "src": "/audio/responsive.mp3",
  "className": "w-full md:w-auto",
  "inline": "${device === 'mobile'}",
  "controls": "${device === 'mobile' ? ['play'] : ['play', 'time', 'process', 'volume']}"
}
```

## Accessibility

### ARIA Support
- `role="application"` for complex audio players
- `aria-label` for control buttons
- `aria-valuemin/max/now` for sliders
- Keyboard navigation support

### Keyboard Navigation
- `Space` - Play/pause
- `Arrow Left/Right` - Skip backward/forward
- `Arrow Up/Down` - Volume control
- `M` - Mute/unmute

### Best Practices
- Provide transcript alternatives
- Include audio descriptions when needed
- Support keyboard navigation
- Ensure sufficient color contrast

## Integration Patterns

### With Form Data
```json
{
  "type": "audio",
  "src": "${audio_field}",
  "controls": ["play", "time", "process"],
  "visibleOn": "${audio_field}"
}
```

### With API Loading
```json
{
  "type": "audio",
  "src": "/api/audio/${audio_id}/stream",
  "controls": ["play", "time", "process", "volume"],
  "onEvent": {
    "loadstart": {
      "actions": [
        {
          "actionType": "setValue",
          "componentId": "loading_status",
          "value": "Loading audio..."
        }
      ]
    }
  }
}
```

### With Conditional Display
```json
{
  "type": "audio",
  "src": "${has_audio ? audio_url : ''}",
  "visibleOn": "${has_audio}",
  "controls": ["play", "time", "process"]
}
```

## Best Practices

### Performance
- Use appropriate audio formats for target browsers
- Implement progressive loading for long audio files
- Avoid autoplay on mobile devices
- Compress audio files appropriately

### User Experience
- Provide clear visual feedback for all states
- Include loading indicators for network audio
- Support both keyboard and mouse interaction
- Show audio duration when available

### Accessibility
- Always provide alternative text content
- Support keyboard navigation
- Include transcripts for important audio
- Test with screen readers

## Go Type Definition

```go
// AudioComponent represents an audio player component
type AudioComponent struct {
    BaseComponent
    
    // Core Properties
    Src string `json:"src"`
    
    // Playback Control
    AutoPlay bool      `json:"autoPlay,omitempty"`
    Loop     bool      `json:"loop,omitempty"`
    Rates    []float64 `json:"rates,omitempty"`
    
    // Interface Control
    Inline   bool     `json:"inline,omitempty"`
    Controls []string `json:"controls,omitempty"`
    
    // Style Properties
    ClassName string                 `json:"className,omitempty"`
    Style     map[string]interface{} `json:"style,omitempty"`
    
    // Event Configuration
    OnEvent map[string]EventConfig `json:"onEvent,omitempty"`
}

// AudioFactory creates Audio components from JSON configuration
func AudioFactory(config map[string]interface{}) (*AudioComponent, error) {
    component := &AudioComponent{
        BaseComponent: BaseComponent{
            Type: "audio",
        },
        AutoPlay: false,
        Loop:     false,
        Inline:   false,
        Controls: []string{"play", "time", "process", "volume"},
        Rates:    []float64{1.0},
    }
    
    return component, mapConfig(config, component)
}

// Render generates the Templ template for the audio player
func (c *AudioComponent) Render() templ.Component {
    return audio.Audio(audio.AudioProps{
        Src:       c.Src,
        AutoPlay:  c.AutoPlay,
        Loop:      c.Loop,
        Rates:     c.Rates,
        Inline:    c.Inline,
        Controls:  c.Controls,
        ClassName: c.ClassName,
        Style:     c.Style,
        OnEvent:   c.OnEvent,
    })
}

// ValidateAudioFormat checks if the audio format is supported
func (c *AudioComponent) ValidateAudioFormat() error {
    if c.Src == "" {
        return fmt.Errorf("audio source is required")
    }
    
    supportedFormats := []string{".mp3", ".wav", ".ogg", ".aac", ".m4a"}
    for _, format := range supportedFormats {
        if strings.HasSuffix(strings.ToLower(c.Src), format) {
            return nil
        }
    }
    
    return fmt.Errorf("unsupported audio format")
}
```

## Related Components

- **[Video](../video/)** - Video player component
- **[Image](../atoms/image/)** - Image display component
- **[Button](../atoms/button/)** - Control button component
- **[Progress](../atoms/progress/)** - Progress indicator component
- **[Card](../card/)** - Container for audio players