# File Control Component

**FILE PURPOSE**: File upload control specifications for document management and attachment handling  
**SCOPE**: All file upload types, drag-and-drop, chunked uploads, and validation patterns  
**TARGET AUDIENCE**: Developers implementing file management, document attachments, and upload workflows

## 📋 Component Overview

The File Control component provides comprehensive file upload capabilities for forms requiring document attachments and media content uploads. It supports multiple file types, drag-and-drop interfaces, chunked uploads for large files, progress tracking, and advanced validation while maintaining accessibility and consistent user experience across different file upload scenarios.

### Schema Reference
- **Primary Schema**: `FileControlSchema.json`
- **Related Schemas**: `FormHorizontal.json`, `LabelAlign.json`, `SchemaApi.json`
- **Base Interface**: Form input control for file selection and upload management

## 🎨 JSON Schema Configuration

The File Control component is configured using JSON that conforms to the `FileControlSchema.json`. The JSON configuration renders interactive file upload interfaces with validation, progress tracking, and user-friendly drag-and-drop capabilities.

## Basic Usage

```json
{
    "type": "input-file",
    "name": "document",
    "label": "Upload Document",
    "accept": ".pdf,.doc,.docx"
}
```

This JSON configuration renders to a Templ component with styled file upload:

```go
// Generated from JSON schema
type FileControlProps struct {
    Type            string                 `json:"type"`
    Name            string                 `json:"name"`
    Label           interface{}            `json:"label"`
    Accept          string                 `json:"accept"`
    Multiple        bool                   `json:"multiple"`
    MaxSize         int64                  `json:"maxSize"`
    MaxLength       int                    `json:"maxLength"`
    AutoUpload      bool                   `json:"autoUpload"`
    Drag            bool                   `json:"drag"`
    UseChunk        interface{}            `json:"useChunk"`
    // ... additional props
}
```

## File Control Variants

### Basic Document Upload
**Purpose**: Simple document attachment for forms

**JSON Configuration:**
```json
{
    "type": "input-file",
    "name": "attachment",
    "label": "Attach Document",
    "placeholder": "Choose file to upload",
    "accept": ".pdf,.doc,.docx,.txt",
    "required": true,
    "maxSize": 10485760,
    "btnLabel": "Choose File",
    "autoUpload": true,
    "hideUploadButton": false,
    "className": "file-upload-basic"
}
```

### Multiple Image Upload
**Purpose**: Gallery and media collection uploads

**JSON Configuration:**
```json
{
    "type": "input-file",
    "name": "product_images",
    "label": "Product Images",
    "placeholder": "Drag images here or click to upload",
    "accept": ".jpg,.jpeg,.png,.gif,.webp",
    "multiple": true,
    "maxLength": 10,
    "maxSize": 5242880,
    "autoUpload": true,
    "drag": true,
    "btnLabel": "Add Images",
    "className": "file-upload-gallery",
    "joinValues": false,
    "extractValue": true,
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "Please upload at least one product image"
    }
}
```

### Large File Upload with Chunking
**Purpose**: Video, archive, and large document uploads

**JSON Configuration:**
```json
{
    "type": "input-file",
    "name": "training_video",
    "label": "Training Video Upload",
    "placeholder": "Upload training video file",
    "accept": ".mp4,.avi,.mkv,.mov",
    "maxSize": 2147483648,
    "autoUpload": true,
    "useChunk": true,
    "chunkSize": 10485760,
    "concurrency": 3,
    "drag": true,
    "btnLabel": "Upload Video",
    "className": "file-upload-large",
    "receiver": "/api/files/upload/video",
    "chunkApi": "/api/files/upload/chunk",
    "startChunkApi": "/api/files/upload/start",
    "finishChunkApi": "/api/files/upload/finish",
    "onEvent": {
        "success": {
            "actions": [
                {
                    "actionType": "toast",
                    "args": {
                        "msg": "Video uploaded successfully!",
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
                        "msg": "Upload failed. Please try again.",
                        "level": "error"
                    }
                }
            ]
        }
    }
}
```

### Profile Picture Upload
**Purpose**: Single image upload with preview and crop

**JSON Configuration:**
```json
{
    "type": "input-file",
    "name": "profile_picture",
    "label": "Profile Picture",
    "placeholder": "Upload your profile photo",
    "accept": ".jpg,.jpeg,.png",
    "multiple": false,
    "maxSize": 2097152,
    "autoUpload": true,
    "asBase64": false,
    "drag": true,
    "btnLabel": "Choose Photo",
    "className": "file-upload-avatar",
    "capture": "user",
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "Profile picture is required"
    },
    "stateTextMap": {
        "init": "Select Photo",
        "pending": "Preparing...",
        "uploading": "Uploading Photo...",
        "error": "Upload Failed",
        "uploaded": "Photo Uploaded",
        "ready": "Ready"
    }
}
```

### CSV Import Upload
**Purpose**: Data import and bulk operations

**JSON Configuration:**
```json
{
    "type": "input-file",
    "name": "csv_import",
    "label": "Import Data File",
    "placeholder": "Upload CSV file for bulk import",
    "accept": ".csv,.xlsx,.xls",
    "required": true,
    "maxSize": 52428800,
    "autoUpload": false,
    "btnLabel": "Select Import File",
    "className": "file-upload-import",
    "templateUrl": "/api/import/template/employees.csv",
    "receiver": "/api/import/process",
    "documentation": "Please ensure your CSV file follows the required format. Download the template for reference.",
    "documentLink": "/docs/import-format",
    "validations": {
        "isRequired": true,
        "matchRegexp": "\\.(csv|xlsx|xls)$"
    },
    "validationErrors": {
        "isRequired": "Please select a file to import",
        "matchRegexp": "Only CSV and Excel files are allowed"
    }
}
```

## File Upload Configuration

### Accepted File Types
Common `accept` patterns:
- **Documents**: `.pdf,.doc,.docx,.txt,.rtf`
- **Images**: `.jpg,.jpeg,.png,.gif,.webp,.svg`
- **Videos**: `.mp4,.avi,.mkv,.mov,.wmv,.flv`
- **Audio**: `.mp3,.wav,.ogg,.aac,.flac`
- **Archives**: `.zip,.rar,.7z,.tar,.gz`
- **Spreadsheets**: `.csv,.xlsx,.xls,.ods`
- **All Files**: `*` (not recommended for security)

### Upload Strategies
- **`autoUpload: true`**: Immediate upload on file selection
- **`autoUpload: false`**: Manual upload with button click
- **`asBase64: true`**: Convert to base64 for small files
- **`asBlob: true`**: Use file data directly in form submission

### Chunked Upload Configuration
```json
{
    "useChunk": true,
    "chunkSize": 10485760,
    "concurrency": 3,
    "chunkApi": "/api/upload/chunk",
    "startChunkApi": "/api/upload/start",
    "finishChunkApi": "/api/upload/finish"
}
```

## Go Type Definitions

```go
// FileControlComponent represents the file upload configuration
type FileControlComponent struct {
    // Base component properties
    Type      string                 `json:"type"`
    Name      string                 `json:"name"`
    Label     interface{}            `json:"label"`
    ClassName interface{}            `json:"className,omitempty"`
    Style     map[string]interface{} `json:"style,omitempty"`
    
    // Visibility and state controls
    Hidden     bool        `json:"hidden,omitempty"`
    HiddenOn   interface{} `json:"hiddenOn,omitempty"`
    Visible    bool        `json:"visible,omitempty"`
    VisibleOn  interface{} `json:"visibleOn,omitempty"`
    Disabled   bool        `json:"disabled,omitempty"`
    DisabledOn interface{} `json:"disabledOn,omitempty"`
    
    // File upload properties
    Accept          string      `json:"accept,omitempty"`           // Allowed file types
    Multiple        bool        `json:"multiple,omitempty"`         // Allow multiple files
    MaxSize         int64       `json:"maxSize,omitempty"`          // Max file size in bytes
    MaxLength       int         `json:"maxLength,omitempty"`        // Max number of files
    AutoUpload      bool        `json:"autoUpload,omitempty"`       // Auto-start upload
    Drag            bool        `json:"drag,omitempty"`             // Enable drag-and-drop
    Capture         string      `json:"capture,omitempty"`          // Camera capture mode
    
    // Upload behavior
    AsBase64        bool        `json:"asBase64,omitempty"`         // Convert to base64
    AsBlob          bool        `json:"asBlob,omitempty"`           // Use blob data
    UseChunk        interface{} `json:"useChunk,omitempty"`         // Enable chunked upload
    ChunkSize       int64       `json:"chunkSize,omitempty"`        // Chunk size in bytes
    Concurrency     int         `json:"concurrency,omitempty"`      // Concurrent chunks
    
    // API endpoints
    Receiver         interface{} `json:"receiver,omitempty"`         // Upload endpoint
    ChunkApi         interface{} `json:"chunkApi,omitempty"`         // Chunk upload endpoint
    StartChunkApi    string      `json:"startChunkApi,omitempty"`    // Start chunk endpoint
    FinishChunkApi   interface{} `json:"finishChunkApi,omitempty"`   // Finish chunk endpoint
    DownloadUrl      interface{} `json:"downloadUrl,omitempty"`      // Download URL pattern
    TemplateUrl      interface{} `json:"templateUrl,omitempty"`      // Template download URL
    
    // UI configuration
    BtnLabel            string      `json:"btnLabel,omitempty"`            // Upload button text
    BtnClassName        interface{} `json:"btnClassName,omitempty"`        // Button CSS classes
    BtnUploadClassName  interface{} `json:"btnUploadClassName,omitempty"`  // Upload button CSS
    HideUploadButton    bool        `json:"hideUploadButton,omitempty"`    // Hide upload button
    
    // Value handling
    JoinValues      bool        `json:"joinValues,omitempty"`       // Join multiple values
    ExtractValue    bool        `json:"extractValue,omitempty"`     // Extract value arrays
    Delimiter       string      `json:"delimiter,omitempty"`        // Value separator
    ValueField      string      `json:"valueField,omitempty"`       // Value field name
    NameField       string      `json:"nameField,omitempty"`        // Name field name
    UrlField        string      `json:"urlField,omitempty"`         // URL field name
    ResetValue      interface{} `json:"resetValue,omitempty"`       // Reset value
    
    // Documentation and help
    Documentation   string `json:"documentation,omitempty"`    // Help text
    DocumentLink    string `json:"documentLink,omitempty"`     // Help link URL
    
    // State management
    StateTextMap    FileStateMap `json:"stateTextMap,omitempty"`  // Upload state labels
    
    // Form integration
    Required            bool                    `json:"required,omitempty"`
    Placeholder         string                  `json:"placeholder,omitempty"`
    Validations         interface{}             `json:"validations,omitempty"`
    ValidationErrors    map[string]string       `json:"validationErrors,omitempty"`
    AutoFill            map[string]interface{}  `json:"autoFill,omitempty"`
    
    // Event handling
    OnEvent map[string]EventConfig `json:"onEvent,omitempty"`
    
    // Testing and debugging
    TestID       string      `json:"testid,omitempty"`
    TestIDBuilder interface{} `json:"testIdBuilder,omitempty"`
}

// FileStateMap represents upload state text mapping
type FileStateMap struct {
    Init      string `json:"init"`
    Pending   string `json:"pending"`
    Uploading string `json:"uploading"`
    Error     string `json:"error"`
    Uploaded  string `json:"uploaded"`
    Ready     string `json:"ready"`
}

// File upload constants
const (
    FileUploadType = "input-file"
    
    // Common file size limits
    FileSizeSmall  = 1024 * 1024      // 1MB
    FileSizeMedium = 5 * 1024 * 1024  // 5MB
    FileSizeLarge  = 10 * 1024 * 1024 // 10MB
    FileSizeHuge   = 100 * 1024 * 1024 // 100MB
    
    // Chunk sizes
    ChunkSizeSmall  = 1 * 1024 * 1024   // 1MB
    ChunkSizeMedium = 5 * 1024 * 1024   // 5MB
    ChunkSizeLarge  = 10 * 1024 * 1024  // 10MB
)

// Common file type patterns
var (
    AcceptImages     = ".jpg,.jpeg,.png,.gif,.webp,.svg"
    AcceptDocuments  = ".pdf,.doc,.docx,.txt,.rtf"
    AcceptSpreadsheets = ".csv,.xlsx,.xls,.ods"
    AcceptVideos     = ".mp4,.avi,.mkv,.mov,.wmv"
    AcceptAudio      = ".mp3,.wav,.ogg,.aac,.flac"
    AcceptArchives   = ".zip,.rar,.7z,.tar,.gz"
)
```

## Templ Implementation

```templ
package molecules

import (
    "fmt"
    "strconv"
    "strings"
)

// FileUpload renders a file upload component
templ FileUpload(props FileControlComponent) {
    <div 
        id={ props.ID }
        class={ getFileUploadClassName(props) }
        style={ templ.Attributes(props.Style) }
        if props.Hidden { hidden }
        data-testid={ props.TestID }
    >
        @FileUploadLabel(props)
        
        <div class={ getFileUploadContainerClass(props) }>
            if props.Drag {
                @DragDropUpload(props)
            } else {
                @StandardFileInput(props)
            }
            
            if props.Documentation != "" {
                @FileUploadDocumentation(props)
            }
        </div>
        
        @FileUploadProgress(props)
        @FileUploadList(props)
    </div>
}

// FileUploadLabel renders the label section
templ FileUploadLabel(props FileControlComponent) {
    if props.Label != nil && props.Label != false {
        <label 
            for={ props.Name }
            class={ getFileLabelClassName(props) }
        >
            { getLabelText(props.Label) }
            if props.Required {
                <span class="text-red-500 ml-1">*</span>
            }
        </label>
    }
}

// DragDropUpload renders drag-and-drop upload area
templ DragDropUpload(props FileControlComponent) {
    <div 
        class="drag-drop-area border-2 border-dashed border-gray-300 rounded-lg p-8 text-center transition-colors hover:border-blue-400 hover:bg-blue-50"
        ondragover="handleDragOver(event)"
        ondrop="handleFileDrop(event)"
        onclick="triggerFileInput(this)"
    >
        <div class="drag-drop-content">
            <svg class="mx-auto h-12 w-12 text-gray-400 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                      d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/>
            </svg>
            <p class="text-lg text-gray-600 mb-2">
                if props.Placeholder != "" {
                    { props.Placeholder }
                } else {
                    Drag files here or click to upload
                }
            </p>
            <p class="text-sm text-gray-500">
                @FileUploadConstraints(props)
            </p>
        </div>
        
        @HiddenFileInput(props)
    </div>
}

// StandardFileInput renders standard file input
templ StandardFileInput(props FileControlComponent) {
    <div class="standard-file-input">
        @HiddenFileInput(props)
        
        <button 
            type="button"
            class={ getFileButtonClassName(props) }
            onclick="triggerFileInput(this)"
        >
            if props.BtnLabel != "" {
                { props.BtnLabel }
            } else {
                Choose File
            }
        </button>
        
        <span class="file-name-display ml-3 text-gray-600"></span>
    </div>
}

// HiddenFileInput renders the actual file input element
templ HiddenFileInput(props FileControlComponent) {
    <input 
        type="file"
        id={ props.Name }
        name={ props.Name }
        class="hidden"
        if props.Accept != "" { accept={ props.Accept } }
        if props.Multiple { multiple }
        if props.Capture != "" { capture={ props.Capture } }
        if props.Required { required }
        onchange="handleFileSelect(this)"
    />
}

// FileUploadConstraints renders upload constraints info
templ FileUploadConstraints(props FileControlComponent) {
    <div class="upload-constraints text-xs text-gray-500">
        if props.Accept != "" {
            <span>Allowed: { formatAcceptTypes(props.Accept) }</span>
        }
        if props.MaxSize > 0 {
            <span class="ml-2">Max: { formatFileSize(props.MaxSize) }</span>
        }
        if props.Multiple && props.MaxLength > 0 {
            <span class="ml-2">Up to { strconv.Itoa(props.MaxLength) } files</span>
        }
    </div>
}

// FileUploadDocumentation renders help documentation
templ FileUploadDocumentation(props FileControlComponent) {
    <div class="upload-documentation mt-2 p-3 bg-blue-50 border border-blue-200 rounded-md">
        <div class="flex items-start">
            <svg class="flex-shrink-0 h-4 w-4 text-blue-400 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
            </svg>
            <div class="ml-2">
                <p class="text-sm text-blue-700">{ props.Documentation }</p>
                if props.DocumentLink != "" {
                    <a href={ props.DocumentLink } class="text-sm text-blue-600 hover:text-blue-800 underline mt-1 inline-block">
                        Learn more
                    </a>
                }
                if props.TemplateUrl != nil {
                    <a href={ getTemplateUrl(props.TemplateUrl) } class="text-sm text-blue-600 hover:text-blue-800 underline mt-1 ml-4 inline-block">
                        Download template
                    </a>
                }
            </div>
        </div>
    </div>
}

// FileUploadProgress renders upload progress
templ FileUploadProgress(props FileControlComponent) {
    <div class="upload-progress mt-3 hidden">
        <div class="bg-gray-200 rounded-full h-2">
            <div class="bg-blue-600 h-2 rounded-full transition-all duration-300" style="width: 0%"></div>
        </div>
        <p class="text-sm text-gray-600 mt-1">Uploading...</p>
    </div>
}

// FileUploadList renders uploaded files list
templ FileUploadList(props FileControlComponent) {
    <div class="file-list mt-3">
        <!-- Files will be dynamically added here -->
    </div>
}

// Helper functions for CSS classes and utilities
func getFileUploadClassName(props FileControlComponent) string {
    classes := []string{"file-upload-component"}
    
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

func getFileUploadContainerClass(props FileControlComponent) string {
    classes := []string{"file-upload-container"}
    
    if props.Drag {
        classes = append(classes, "drag-enabled")
    }
    
    return strings.Join(classes, " ")
}

func getFileButtonClassName(props FileControlComponent) string {
    classes := []string{"file-upload-btn", "px-4", "py-2", "bg-blue-600", "text-white", "rounded-md", "hover:bg-blue-700", "focus:outline-none", "focus:ring-2", "focus:ring-blue-500", "focus:ring-offset-2"}
    
    if props.BtnClassName != nil {
        switch btnClass := props.BtnClassName.(type) {
        case string:
            if btnClass != "" {
                classes = append(classes, btnClass)
            }
        case []string:
            classes = append(classes, btnClass...)
        }
    }
    
    return strings.Join(classes, " ")
}

func getFileLabelClassName(props FileControlComponent) string {
    classes := []string{"block", "text-sm", "font-medium", "text-gray-700", "mb-2"}
    
    if props.LabelClassName != "" {
        classes = append(classes, props.LabelClassName)
    }
    
    return strings.Join(classes, " ")
}

func formatAcceptTypes(accept string) string {
    types := strings.Split(accept, ",")
    var formatted []string
    
    for _, t := range types {
        t = strings.TrimSpace(t)
        if strings.HasPrefix(t, ".") {
            formatted = append(formatted, strings.ToUpper(t[1:]))
        } else {
            formatted = append(formatted, t)
        }
    }
    
    return strings.Join(formatted, ", ")
}

func formatFileSize(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func getLabelText(label interface{}) string {
    switch l := label.(type) {
    case string:
        return l
    case bool:
        return ""
    default:
        return ""
    }
}

func getTemplateUrl(templateUrl interface{}) string {
    switch url := templateUrl.(type) {
    case string:
        return url
    default:
        return ""
    }
}
```

## CSS Styling

```css
/* Base file upload component styles */
.file-upload-component {
    @apply w-full;
}

.file-upload-container {
    @apply relative;
}

/* Drag and drop area */
.drag-drop-area {
    @apply cursor-pointer transition-all duration-200;
}

.drag-drop-area:hover {
    @apply border-blue-400 bg-blue-50;
}

.drag-drop-area.drag-over {
    @apply border-blue-500 bg-blue-100;
}

.drag-drop-content {
    @apply pointer-events-none;
}

/* Standard file input button */
.file-upload-btn {
    @apply transition-colors duration-200;
}

.file-upload-btn:disabled {
    @apply bg-gray-400 cursor-not-allowed;
}

/* Upload progress */
.upload-progress {
    @apply transition-opacity duration-300;
}

.upload-progress.active {
    @apply block;
}

/* File list */
.file-list {
    @apply space-y-2;
}

.file-item {
    @apply flex items-center justify-between p-3 bg-gray-50 border border-gray-200 rounded-lg;
}

.file-item.uploading {
    @apply bg-blue-50 border-blue-200;
}

.file-item.error {
    @apply bg-red-50 border-red-200;
}

.file-item.success {
    @apply bg-green-50 border-green-200;
}

.file-info {
    @apply flex items-center space-x-3;
}

.file-icon {
    @apply flex-shrink-0 w-8 h-8 text-gray-400;
}

.file-details {
    @apply min-w-0 flex-1;
}

.file-name {
    @apply text-sm font-medium text-gray-900 truncate;
}

.file-size {
    @apply text-xs text-gray-500;
}

.file-actions {
    @apply flex items-center space-x-2;
}

.file-remove {
    @apply text-gray-400 hover:text-red-500 cursor-pointer;
}

/* Upload state indicators */
.upload-state {
    @apply text-xs font-medium;
}

.upload-state.pending {
    @apply text-yellow-600;
}

.upload-state.uploading {
    @apply text-blue-600;
}

.upload-state.error {
    @apply text-red-600;
}

.upload-state.success {
    @apply text-green-600;
}

/* File upload variants */
.file-upload-basic {
    @apply border border-gray-200 rounded-lg p-4 bg-white;
}

.file-upload-gallery {
    @apply border-2 border-dashed border-gray-300 rounded-xl p-6 bg-gray-50;
}

.file-upload-large {
    @apply border-2 border-dashed border-blue-300 rounded-xl p-8 bg-blue-50;
}

.file-upload-avatar {
    @apply border border-gray-200 rounded-lg p-4 bg-white text-center;
}

.file-upload-import {
    @apply border border-yellow-200 rounded-lg p-4 bg-yellow-50;
}

/* Documentation styles */
.upload-documentation {
    @apply transition-all duration-200;
}

/* Responsive adjustments */
@media (max-width: 640px) {
    .drag-drop-area {
        @apply p-4;
    }
    
    .file-item {
        @apply flex-col items-start space-y-2;
    }
    
    .file-actions {
        @apply w-full justify-end;
    }
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
    .file-upload-component {
        @apply text-gray-100;
    }
    
    .drag-drop-area {
        @apply border-gray-600 bg-gray-800;
    }
    
    .drag-drop-area:hover {
        @apply border-blue-400 bg-gray-700;
    }
    
    .file-upload-btn {
        @apply bg-blue-600 hover:bg-blue-700;
    }
    
    .file-item {
        @apply bg-gray-800 border-gray-600;
    }
}

/* Accessibility enhancements */
.file-upload-component:focus-within {
    @apply ring-2 ring-blue-500 ring-opacity-50;
}

.drag-drop-area:focus-visible {
    @apply outline-none ring-2 ring-blue-500 ring-opacity-50;
}

/* Loading and error states */
.file-upload-component.loading .drag-drop-area {
    @apply opacity-50 pointer-events-none;
}

.file-upload-component.error .drag-drop-area {
    @apply border-red-300 bg-red-50;
}
```

## JavaScript Helper Functions

```javascript
// File input trigger
function triggerFileInput(element) {
    const input = element.closest('.file-upload-component').querySelector('input[type="file"]');
    if (input) {
        input.click();
    }
}

// File selection handler
function handleFileSelect(input) {
    const files = Array.from(input.files);
    const container = input.closest('.file-upload-component');
    
    files.forEach(file => {
        if (validateFile(file, input)) {
            addFileToList(file, container);
            if (shouldAutoUpload(input)) {
                uploadFile(file, container);
            }
        }
    });
}

// Drag and drop handlers
function handleDragOver(event) {
    event.preventDefault();
    event.currentTarget.classList.add('drag-over');
}

function handleFileDrop(event) {
    event.preventDefault();
    event.currentTarget.classList.remove('drag-over');
    
    const files = Array.from(event.dataTransfer.files);
    const input = event.currentTarget.querySelector('input[type="file"]');
    const container = event.currentTarget.closest('.file-upload-component');
    
    files.forEach(file => {
        if (validateFile(file, input)) {
            addFileToList(file, container);
            if (shouldAutoUpload(input)) {
                uploadFile(file, container);
            }
        }
    });
}

// File validation
function validateFile(file, input) {
    const accept = input.getAttribute('accept');
    const maxSize = parseInt(input.dataset.maxSize || '0');
    
    // Validate file type
    if (accept && !isFileTypeAllowed(file, accept)) {
        showError(`File type ${getFileExtension(file.name)} is not allowed`);
        return false;
    }
    
    // Validate file size
    if (maxSize > 0 && file.size > maxSize) {
        showError(`File size exceeds ${formatFileSize(maxSize)} limit`);
        return false;
    }
    
    return true;
}

// File type validation
function isFileTypeAllowed(file, accept) {
    const fileExtension = getFileExtension(file.name).toLowerCase();
    const allowedTypes = accept.toLowerCase().split(',').map(t => t.trim());
    
    return allowedTypes.some(type => {
        if (type === '*') return true;
        if (type.startsWith('.')) return type === fileExtension;
        if (type.includes('/')) return file.type.match(type.replace('*', '.*'));
        return false;
    });
}

// Utility functions
function getFileExtension(filename) {
    return '.' + filename.split('.').pop().toLowerCase();
}

function formatFileSize(bytes) {
    const sizes = ['B', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 B';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
}

function shouldAutoUpload(input) {
    return input.dataset.autoUpload === 'true';
}

function showError(message) {
    // Implementation depends on your toast/notification system
    console.error(message);
}
```

## Accessibility Features

- **Keyboard Navigation**: Full keyboard access for file selection
- **Screen Reader Support**: Proper ARIA labels and file descriptions
- **Focus Management**: Clear focus indicators and logical tab order
- **Progress Announcements**: Screen reader updates during upload
- **Error Communication**: Clear error messages and validation feedback
- **High Contrast**: Compatible with high contrast mode

## Testing Strategy

### Unit Tests
```go
func TestFileControlComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    FileControlComponent
        expected string
    }{
        {
            name: "basic file upload",
            props: FileControlComponent{
                Type: "input-file",
                Name: "document",
                Accept: ".pdf,.doc",
                MaxSize: FileSizeMedium,
            },
            expected: "file upload with document types",
        },
        {
            name: "multiple image upload with drag",
            props: FileControlComponent{
                Type: "input-file",
                Name: "images",
                Accept: AcceptImages,
                Multiple: true,
                Drag: true,
                MaxLength: 5,
            },
            expected: "multiple image upload with drag-and-drop",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test file control component rendering
            result := renderFileControlComponent(tt.props)
            assert.Contains(t, result, "file-upload-component")
            
            if tt.props.Drag {
                assert.Contains(t, result, "drag-drop-area")
            }
            
            if tt.props.Multiple {
                assert.Contains(t, result, "multiple")
            }
        })
    }
}
```

### Integration Tests
- Test file selection and upload functionality
- Verify drag-and-drop behavior
- Test chunked upload for large files
- Validate file type and size restrictions
- Test accessibility compliance
- Test responsive behavior across devices

## Real-World ERP Use Cases

### 1. Employee Document Management
```json
{
    "type": "input-file",
    "name": "employee_documents",
    "label": "Employee Documents",
    "accept": ".pdf,.doc,.docx,.jpg,.jpeg,.png",
    "multiple": true,
    "maxLength": 10,
    "maxSize": 10485760,
    "autoUpload": true,
    "drag": true,
    "btnLabel": "Upload Documents",
    "className": "file-upload-hr",
    "receiver": "/api/hr/employees/${employee.id}/documents",
    "onEvent": {
        "success": {
            "actions": [
                {
                    "actionType": "reload",
                    "componentId": "employee_documents_list"
                },
                {
                    "actionType": "toast",
                    "args": {
                        "msg": "Documents uploaded successfully",
                        "level": "success"
                    }
                }
            ]
        }
    }
}
```

### 2. Invoice Attachment System
```json
{
    "type": "input-file",
    "name": "invoice_attachments",
    "label": "Invoice Attachments",
    "accept": ".pdf,.jpg,.jpeg,.png",
    "multiple": true,
    "maxLength": 5,
    "maxSize": 5242880,
    "autoUpload": true,
    "drag": true,
    "className": "file-upload-invoice",
    "receiver": "/api/accounting/invoices/${invoice.id}/attachments",
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "Please attach invoice documentation"
    }
}
```

### 3. Product Catalog Images
```json
{
    "type": "input-file",
    "name": "product_images",
    "label": "Product Images",
    "accept": ".jpg,.jpeg,.png,.webp",
    "multiple": true,
    "maxLength": 10,
    "maxSize": 5242880,
    "autoUpload": true,
    "drag": true,
    "className": "file-upload-product",
    "receiver": "/api/catalog/products/${product.id}/images",
    "extractValue": true,
    "onEvent": {
        "success": {
            "actions": [
                {
                    "actionType": "reload",
                    "componentId": "product_gallery"
                }
            ]
        }
    }
}
```

### 4. Bulk Data Import
```json
{
    "type": "input-file",
    "name": "bulk_import",
    "label": "Import Data File",
    "accept": ".csv,.xlsx,.xls",
    "required": true,
    "maxSize": 52428800,
    "autoUpload": false,
    "className": "file-upload-import",
    "receiver": "/api/import/process",
    "templateUrl": "/api/import/templates/${import.type}",
    "documentation": "Upload your data file for bulk import. Ensure the file follows the required format.",
    "documentLink": "/docs/import-guidelines",
    "onEvent": {
        "success": {
            "actions": [
                {
                    "actionType": "redirect",
                    "args": {
                        "url": "/import/status/${import.id}"
                    }
                }
            ]
        }
    }
}
```

The File Control component provides comprehensive file upload capabilities essential for modern ERP systems, supporting document management, media uploads, bulk data import, and attachment workflows with full accessibility and professional styling.