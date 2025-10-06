package preview

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/niiniyare/erp/web/builder/core"
)

// PreviewRenderer handles real-time schema compilation and rendering
type PreviewRenderer struct {
	registry       *core.ComponentRegistry
	cache          *PreviewCache
	errorBoundary  *ErrorBoundary
	updateChannels map[string]chan *PreviewUpdate
	mu             sync.RWMutex
}

// PreviewUpdate represents a real-time update to the preview
type PreviewUpdate struct {
	SchemaID   string                 `json:"schemaId"`
	HTML       string                 `json:"html"`
	CSS        string                 `json:"css"`
	JavaScript string                 `json:"javascript"`
	Errors     []string               `json:"errors,omitempty"`
	Warnings   []string               `json:"warnings,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// PreviewCache handles caching for performance optimization
type PreviewCache struct {
	compiledSchemas map[string]*CachedPreview
	lastModified    map[string]time.Time
	mu              sync.RWMutex
	maxSize         int
	ttl             time.Duration
}

// CachedPreview represents a cached compiled preview
type CachedPreview struct {
	HTML       string
	CSS        string
	JavaScript string
	Checksum   string
	CreatedAt  time.Time
	AccessedAt time.Time
}

// ErrorBoundary handles compilation errors with visual feedback
type ErrorBoundary struct {
	errors map[string][]PreviewError
	mu     sync.RWMutex
}

// PreviewError represents a compilation or runtime error
type PreviewError struct {
	Type        string    `json:"type"` // "compilation", "runtime", "validation"
	Message     string    `json:"message"`
	Component   string    `json:"component"`
	Line        int       `json:"line,omitempty"`
	Column      int       `json:"column,omitempty"`
	Severity    string    `json:"severity"` // "error", "warning", "info"
	Timestamp   time.Time `json:"timestamp"`
	Suggestions []string  `json:"suggestions,omitempty"`
}

// NewPreviewRenderer creates a new preview renderer
func NewPreviewRenderer(registry *core.ComponentRegistry) *PreviewRenderer {
	return &PreviewRenderer{
		registry: registry,
		cache: &PreviewCache{
			compiledSchemas: make(map[string]*CachedPreview),
			lastModified:    make(map[string]time.Time),
			maxSize:         100,
			ttl:             5 * time.Minute,
		},
		errorBoundary: &ErrorBoundary{
			errors: make(map[string][]PreviewError),
		},
		updateChannels: make(map[string]chan *PreviewUpdate),
	}
}

// RenderPreview compiles and renders a schema for preview
func (r *PreviewRenderer) RenderPreview(schema *core.CompositionSchema) (*PreviewUpdate, error) {
	startTime := time.Now()

	// Check cache first
	if cached := r.getCachedPreview(schema.ID); cached != nil {
		return &PreviewUpdate{
			SchemaID:   schema.ID,
			HTML:       cached.HTML,
			CSS:        cached.CSS,
			JavaScript: cached.JavaScript,
			Timestamp:  time.Now(),
			Metadata: map[string]interface{}{
				"renderTime": time.Since(startTime).Milliseconds(),
				"fromCache":  true,
			},
		}, nil
	}

	// Clear previous errors
	r.errorBoundary.clearErrors(schema.ID)

	// Compile schema to preview
	html, css, js, errors := r.compileSchema(schema)

	// Create preview update
	update := &PreviewUpdate{
		SchemaID:   schema.ID,
		HTML:       html,
		CSS:        css,
		JavaScript: js,
		Errors:     r.formatErrors(errors),
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"renderTime":     time.Since(startTime).Milliseconds(),
			"fromCache":      false,
			"componentCount": len(schema.Components),
		},
	}

	// Cache successful compilation
	if len(errors) == 0 {
		r.cachePreview(schema.ID, html, css, js)
	}

	return update, nil
}

// RenderPreviewAsync renders preview asynchronously and sends updates via channel
func (r *PreviewRenderer) RenderPreviewAsync(schema *core.CompositionSchema, updates chan<- *PreviewUpdate) {
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Preview rendering panic: %v", rec)
				updates <- &PreviewUpdate{
					SchemaID:  schema.ID,
					Errors:    []string{fmt.Sprintf("Rendering panic: %v", rec)},
					Timestamp: time.Now(),
				}
			}
		}()

		update, err := r.RenderPreview(schema)
		if err != nil {
			updates <- &PreviewUpdate{
				SchemaID:  schema.ID,
				Errors:    []string{err.Error()},
				Timestamp: time.Now(),
			}
			return
		}

		updates <- update
	}()
}

// SubscribeToUpdates creates a channel for real-time updates
func (r *PreviewRenderer) SubscribeToUpdates(schemaID string) <-chan *PreviewUpdate {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan *PreviewUpdate, 10)
	r.updateChannels[schemaID] = ch
	return ch
}

// UnsubscribeFromUpdates removes a subscription
func (r *PreviewRenderer) UnsubscribeFromUpdates(schemaID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ch, exists := r.updateChannels[schemaID]; exists {
		close(ch)
		delete(r.updateChannels, schemaID)
	}
}

// NotifyUpdate sends updates to all subscribers
func (r *PreviewRenderer) NotifyUpdate(update *PreviewUpdate) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if ch, exists := r.updateChannels[update.SchemaID]; exists {
		select {
		case ch <- update:
		default:
			// Channel full, skip update
		}
	}
}

// compileSchema compiles a schema definition into HTML/CSS/JS
func (r *PreviewRenderer) compileSchema(schema *core.CompositionSchema) (string, string, string, []PreviewError) {
	var errors []PreviewError
	var htmlBuffer bytes.Buffer
	var cssBuffer bytes.Buffer
	var jsBuffer bytes.Buffer

	// Validate schema first
	if validationErrors := r.validateSchema(schema); len(validationErrors) > 0 {
		errors = append(errors, validationErrors...)
		return "", "", "", errors
	}

	// Generate HTML structure
	htmlBuffer.WriteString(r.generateHTMLWrapper(schema))

	// Render each component
	for _, instance := range schema.Components {
		componentHTML, componentCSS, componentJS, compErrors := r.renderComponent(&instance)

		if len(compErrors) > 0 {
			errors = append(errors, compErrors...)
			continue
		}

		htmlBuffer.WriteString(componentHTML)
		if componentCSS != "" {
			cssBuffer.WriteString(componentCSS)
			cssBuffer.WriteString("\n")
		}
		if componentJS != "" {
			jsBuffer.WriteString(componentJS)
			jsBuffer.WriteString("\n")
		}
	}

	htmlBuffer.WriteString("</div></body></html>")

	// Add responsive CSS
	cssBuffer.WriteString(r.generateResponsiveCSS())

	// Add interaction JavaScript
	jsBuffer.WriteString(r.generateInteractionJS())

	return htmlBuffer.String(), cssBuffer.String(), jsBuffer.String(), errors
}

// validateSchema validates a schema for common issues
func (r *PreviewRenderer) validateSchema(schema *core.CompositionSchema) []PreviewError {
	var errors []PreviewError

	// Check for required fields
	if schema.Name == "" {
		errors = append(errors, PreviewError{
			Type:      "validation",
			Message:   "Schema name is required",
			Severity:  "error",
			Timestamp: time.Now(),
		})
	}

	// Check for component conflicts
	usedIDs := make(map[string]bool)
	for _, instance := range schema.Components {
		if usedIDs[instance.ID] {
			errors = append(errors, PreviewError{
				Type:      "validation",
				Message:   fmt.Sprintf("Duplicate component ID: %s", instance.ID),
				Component: instance.ID,
				Severity:  "error",
				Timestamp: time.Now(),
			})
		}
		usedIDs[instance.ID] = true

		// Validate component exists in registry
		if !r.registry.HasComponent(instance.Type) {
			errors = append(errors, PreviewError{
				Type:      "validation",
				Message:   fmt.Sprintf("Unknown component type: %s", instance.Type),
				Component: instance.ID,
				Severity:  "error",
				Timestamp: time.Now(),
				Suggestions: []string{
					"Check component type spelling",
					"Ensure component is registered",
				},
			})
		}
	}

	return errors
}

// renderComponent renders a single component instance
func (r *PreviewRenderer) renderComponent(instance *core.ComponentInstance) (string, string, string, []PreviewError) {
	var errors []PreviewError

	// Get component definition
	componentDef := r.registry.GetComponent(instance.Type)
	if componentDef == nil {
		errors = append(errors, PreviewError{
			Type:      "compilation",
			Message:   fmt.Sprintf("Component not found: %s", instance.Type),
			Component: instance.ID,
			Severity:  "error",
			Timestamp: time.Now(),
		})
		return "", "", "", errors
	}

	// Merge properties with defaults
	mergedProps := r.mergeProperties(componentDef.DefaultProps, instance.Props)

	// Apply data bindings if present
	if instance.Data != nil {
		r.applyDataBindings(mergedProps, instance.Data)
	}

	// Generate component HTML
	html := r.generateComponentHTML(componentDef, instance, mergedProps)

	// Generate component-specific CSS
	css := r.generateComponentCSS(componentDef, instance, mergedProps)

	// Generate component-specific JavaScript
	js := r.generateComponentJS(componentDef, instance, mergedProps)

	return html, css, js, errors
}

// generateHTMLWrapper creates the base HTML structure
func (r *PreviewRenderer) generateHTMLWrapper(schema *core.CompositionSchema) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Preview</title>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
    <style id="preview-styles"></style>
</head>
<body class="bg-gray-50">
    <div id="preview-root" class="min-h-screen">`, schema.Name)
}

// generateComponentHTML creates HTML for a component
func (r *PreviewRenderer) generateComponentHTML(def *core.ComponentDefinition, instance *core.ComponentInstance, props map[string]interface{}) string {
	// Use Go template to render component
	tmpl := template.New("component")
	tmpl, err := tmpl.Parse(def.Template.HTML)
	if err != nil {
		return fmt.Sprintf(`<div class="error">Template error: %s</div>`, err.Error())
	}

	var buf bytes.Buffer
	data := map[string]interface{}{
		"ID":       instance.ID,
		"Props":    props,
		"Position": instance.Position,
		"Size":     instance.Size,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf(`<div class="error">Render error: %s</div>`, err.Error())
	}

	return buf.String()
}

// generateComponentCSS creates CSS for a component
func (r *PreviewRenderer) generateComponentCSS(def *core.ComponentDefinition, instance *core.ComponentInstance, props map[string]interface{}) string {
	var css strings.Builder

	// Add positioning CSS
	css.WriteString(fmt.Sprintf(`#%s {
		position: absolute;
		left: %.0fpx;
		top: %.0fpx;`, instance.ID, instance.Position.X, instance.Position.Y))

	if instance.Size.Width > 0 {
		css.WriteString(fmt.Sprintf(`
		width: %.0fpx;`, instance.Size.Width))
	}

	if instance.Size.Height > 0 {
		css.WriteString(fmt.Sprintf(`
		height: %.0fpx;`, instance.Size.Height))
	}

	css.WriteString(`
	}
	`)

	// Add component-specific styles
	if def.Template.CSS != "" {
		css.WriteString(def.Template.CSS)
		css.WriteString("\n")
	}

	// Add custom styles from instance
	if instance.Style != nil && len(instance.Style.InlineStyles) > 0 {
		for selector, rules := range instance.Style.InlineStyles {
			css.WriteString(fmt.Sprintf(`%s { %s }
`, selector, rules))
		}
	}

	return css.String()
}

// generateComponentJS creates JavaScript for a component
func (r *PreviewRenderer) generateComponentJS(def *core.ComponentDefinition, instance *core.ComponentInstance, props map[string]interface{}) string {
	var js strings.Builder

	// Add component initialization
	js.WriteString(fmt.Sprintf(`
// Initialize component %s
(function() {
	const component = document.getElementById('%s');
	if (!component) return;
	
	// Component-specific logic
`, instance.ID, instance.ID))

	// Add component scripts
	if def.Template.JS != "" {
		js.WriteString(def.Template.JS)
		js.WriteString("\n")
	}

	// Add event handlers
	if instance.Events != nil {
		for _, handler := range instance.Events.Handlers {
			js.WriteString(fmt.Sprintf(`
	component.addEventListener('%s', function(e) {
		// Event handler for %s
		console.log('Event triggered:', e.type);
	});`, handler.Event, handler.ID))
		}
	}

	js.WriteString(`
})();
`)

	return js.String()
}

// Helper methods for caching and utilities
func (r *PreviewRenderer) getCachedPreview(schemaID string) *CachedPreview {
	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()

	cached, exists := r.cache.compiledSchemas[schemaID]
	if !exists {
		return nil
	}

	// Check TTL
	if time.Since(cached.CreatedAt) > r.cache.ttl {
		return nil
	}

	// Update access time
	cached.AccessedAt = time.Now()
	return cached
}

func (r *PreviewRenderer) cachePreview(schemaID, html, css, js string) {
	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()

	// Clean cache if at max size
	if len(r.cache.compiledSchemas) >= r.cache.maxSize {
		r.cleanOldestCache()
	}

	r.cache.compiledSchemas[schemaID] = &CachedPreview{
		HTML:       html,
		CSS:        css,
		JavaScript: js,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
	}
}

func (r *PreviewRenderer) cleanOldestCache() {
	var oldestID string
	var oldestTime time.Time

	for id, cached := range r.cache.compiledSchemas {
		if oldestID == "" || cached.AccessedAt.Before(oldestTime) {
			oldestID = id
			oldestTime = cached.AccessedAt
		}
	}

	if oldestID != "" {
		delete(r.cache.compiledSchemas, oldestID)
	}
}

func (r *PreviewRenderer) mergeProperties(defaults, instance map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy defaults
	for k, v := range defaults {
		merged[k] = v
	}

	// Override with instance properties
	for k, v := range instance {
		merged[k] = v
	}

	return merged
}

func (r *PreviewRenderer) applyDataBindings(props map[string]interface{}, data *core.DataBinding) {
	// Apply data source bindings
	for _, binding := range data.Bindings {
		if sourceValue, exists := data.Source[binding.Path]; exists {
			props[binding.Property] = sourceValue
		}
	}
}

func (r *PreviewRenderer) generateResponsiveCSS() string {
	return `
/* Responsive utilities for preview */
.preview-container {
	transition: all 0.3s ease;
}

@media (max-width: 768px) {
	.preview-container {
		padding: 0.5rem;
	}
}

@media (max-width: 480px) {
	.preview-container {
		padding: 0.25rem;
	}
}
`
}

func (r *PreviewRenderer) generateInteractionJS() string {
	return `
// Preview interaction utilities
window.PreviewUtils = {
	highlightComponent: function(componentId) {
		const element = document.getElementById(componentId);
		if (element) {
			element.style.outline = '2px solid #3B82F6';
			element.style.outlineOffset = '2px';
		}
	},
	
	unhighlightComponent: function(componentId) {
		const element = document.getElementById(componentId);
		if (element) {
			element.style.outline = '';
			element.style.outlineOffset = '';
		}
	},
	
	selectComponent: function(componentId) {
		// Clear previous selection
		document.querySelectorAll('.selected-component').forEach(el => {
			el.classList.remove('selected-component');
		});
		
		// Select new component
		const element = document.getElementById(componentId);
		if (element) {
			element.classList.add('selected-component');
			// Notify parent frame
			if (window.parent) {
				window.parent.postMessage({
					type: 'componentSelected',
					componentId: componentId
				}, '*');
			}
		}
	}
};

// Add click handlers for component selection
document.addEventListener('click', function(e) {
	if (e.target.id && e.target.id !== 'preview-root') {
		PreviewUtils.selectComponent(e.target.id);
	}
});
`
}

func (r *PreviewRenderer) formatErrors(errors []PreviewError) []string {
	var formatted []string
	for _, err := range errors {
		formatted = append(formatted, fmt.Sprintf("[%s] %s", err.Type, err.Message))
	}
	return formatted
}

func (r *ErrorBoundary) clearErrors(schemaID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.errors, schemaID)
}

func (r *ErrorBoundary) addError(schemaID string, err PreviewError) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors[schemaID] = append(r.errors[schemaID], err)
}

func (r *ErrorBoundary) getErrors(schemaID string) []PreviewError {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.errors[schemaID]
}
