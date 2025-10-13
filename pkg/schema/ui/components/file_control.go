// Package components - FileControlSchema for file upload components  
// Based on JSON schema: FileControlSchema.json
package components

import (
	"encoding/json"
	"fmt"
)

// FileControlSize represents the size options for file controls
type FileControlSize string

const (
	FileControlSizeXS   FileControlSize = "xs"   // Extra small size
	FileControlSizeSM   FileControlSize = "sm"   // Small size  
	FileControlSizeMD   FileControlSize = "md"   // Medium size (default)
	FileControlSizeLG   FileControlSize = "lg"   // Large size
	FileControlSizeFull FileControlSize = "full" // Full width size
)

// FileUploadMode represents how files should be processed
type FileUploadMode string

const (
	FileUploadModeNormal   FileUploadMode = "normal"   // Standard file upload to server
	FileUploadModeBase64   FileUploadMode = "base64"   // Convert to base64 and include in form
	FileUploadModeBlob     FileUploadMode = "blob"     // Use file data directly as form value
)

// ChunkUploadMode represents the chunked upload behavior
type ChunkUploadMode string

const (
	ChunkUploadAuto     ChunkUploadMode = "auto" // Automatically use chunks for large files
	ChunkUploadEnabled  ChunkUploadMode = "true" // Always use chunked upload
	ChunkUploadDisabled ChunkUploadMode = "false" // Never use chunked upload
)

// FileCaptureMode represents camera capture options for mobile devices
type FileCaptureMode string

const (
	FileCaptureUser        FileCaptureMode = "user"        // Front-facing camera
	FileCaptureEnvironment FileCaptureMode = "environment" // Back-facing camera
	FileCaptureCamera      FileCaptureMode = "camera"      // Any camera
	FileCaptureCamcorder   FileCaptureMode = "camcorder"   // Video capture
	FileCaptureFile        FileCaptureMode = "file"        // File system
)

// CropDragMode represents the drag behavior for image cropping
type CropDragMode string

const (
	CropDragModeCrop CropDragMode = "crop" // Create new crop box
	CropDragModeMove CropDragMode = "move" // Move the canvas
	CropDragModeNone CropDragMode = "none" // Do nothing
)

// CropViewMode represents the view restriction mode for cropper
type CropViewMode int

const (
	CropViewModeRestrict0 CropViewMode = 0 // No restrictions  
	CropViewModeRestrict1 CropViewMode = 1 // Restrict crop box to not exceed canvas size
	CropViewModeRestrict2 CropViewMode = 2 // Minimum canvas size fit contained in container  
	CropViewModeRestrict3 CropViewMode = 3 // Minimum canvas size fill fit container
)

// CropConfig represents image cropping configuration
type CropConfig struct {
	// Aspect ratio of the crop box (width / height)
	AspectRatio float64 `json:"aspectRatio,omitempty"`
	// Whether to show rotation controls before cropping
	RotateBefore bool `json:"rotateBefore,omitempty"`
	// Whether to upload only the cropped image
	UploadCropOnly bool `json:"uploadCropOnly,omitempty"`
	// Resize width after cropping (pixels)
	ResizeWidth int `json:"resizeWidth,omitempty"`
	// Resize height after cropping (pixels) 
	ResizeHeight int `json:"resizeHeight,omitempty"`
	// JPEG quality for resized image (0-1)
	ResizeQuality float64 `json:"resizeQuality,omitempty"`
	
	// Cropper.js Configuration
	// Define the view mode of the cropper
	ViewMode CropViewMode `json:"viewMode,omitempty"`
	// Define the drag mode of the cropper
	DragMode CropDragMode `json:"dragMode,omitempty"`
	// Show the dashed lines above the crop box
	Guides bool `json:"guides,omitempty"`
	// Show the center indicator above the crop box
	Center bool `json:"center,omitempty"`
	// Show the white modal above the crop box
	Highlight bool `json:"highlight,omitempty"`
	// Show the grid background of the container
	Background bool `json:"background,omitempty"`
	// Enable to crop the image automatically when initialized
	AutoCrop bool `json:"autoCrop,omitempty"`
	// Define the percentage of automatic cropping area when initializing (0-1)
	AutoCropArea float64 `json:"autoCropArea,omitempty"`
	// Enable to move the image
	Movable bool `json:"movable,omitempty"`
	// Enable to rotate the image
	Rotatable bool `json:"rotatable,omitempty"`
	// Enable to scale the image
	Scalable bool `json:"scalable,omitempty"`
	// Enable to zoom the image
	Zoomable bool `json:"zoomable,omitempty"`
	// Enable to zoom the image by touching
	ZoomOnTouch bool `json:"zoomOnTouch,omitempty"`
	// Enable to zoom the image by mouse wheeling
	ZoomOnWheel bool `json:"zoomOnWheel,omitempty"`
	// Define zoom ratio when zooming the image by mouse wheeling
	WheelZoomRatio float64 `json:"wheelZoomRatio,omitempty"`
	// Enable to move the crop box by dragging
	CropBoxMovable bool `json:"cropBoxMovable,omitempty"`
	// Enable to resize the crop box by dragging
	CropBoxResizable bool `json:"cropBoxResizable,omitempty"`
	// Enable to toggle drag mode on double click
	ToggleDragModeOnDblclick bool `json:"toggleDragModeOnDblclick,omitempty"`
}

// CompressOptions represents file compression settings
type CompressOptions struct {
	// Maximum width for image compression
	MaxWidth int `json:"maxWidth,omitempty"`
	// Maximum height for image compression
	MaxHeight int `json:"maxHeight,omitempty"`
	// JPEG quality for compression (0-1)
	Quality float64 `json:"quality,omitempty"`
	// File size threshold to trigger compression (bytes)
	ConvertSize int `json:"convertSize,omitempty"`
	// Target MIME type after compression
	MimeType string `json:"mimeType,omitempty"`
}

// FileControlSchema represents a file upload control component
// This component handles file selection, upload, and management with advanced features
// like chunked upload, image cropping, and compression
// Based on AMis File control: https://aisuda.bce.baidu.com/amis/zh-CN/components/form/input-file
type FileControlSchema struct {
	BaseComponentProps

	// Component type identifier (required)
	Type string `json:"type"` // Must be "input-file"

	// Field name for form submission, supports multi-level paths (e.g., "a.b.c") (required)
	Name string `json:"name"`

	// Display label for the file control
	Label string `json:"label,omitempty"`

	// Default file value or URL
	Value any `json:"value,omitempty"`

	// Placeholder text shown in empty state
	Placeholder string `json:"placeholder,omitempty"`

	// Size of the file control
	Size FileControlSize `json:"size,omitempty"`

	// Whether the field is required
	Required bool `json:"required,omitempty"`

	// Whether the control is read-only
	ReadOnly bool `json:"readOnly,omitempty"`

	// Whether multiple files can be selected
	Multiple bool `json:"multiple,omitempty"`

	// File Selection Configuration
	// Maximum number of files that can be selected
	MaxLength int `json:"maxLength,omitempty"`
	// Maximum file size allowed (bytes)  
	MaxSize any `json:"maxSize,omitempty"` // number or string like "10MB"
	// File type filter (e.g., ".jpg,.png,image/*")
	Accept string `json:"accept,omitempty"`
	// Camera capture mode for mobile devices
	Capture FileCaptureMode `json:"capture,omitempty"`

	// Upload Configuration
	// Whether to start upload automatically after file selection
	AutoUpload bool `json:"autoUpload,omitempty"`
	// Main upload API endpoint
	Receiver *APIConfig `json:"receiver,omitempty"`
	// Alternative API configuration
	API *APIConfig `json:"api,omitempty"`
	// File deletion API endpoint
	DeleteAPI *APIConfig `json:"deleteApi,omitempty"`
	// File download API endpoint configuration
	DownloadAPI *APIConfig `json:"downloadApi,omitempty"`
	// Template file download URL
	TemplateURL *APIConfig `json:"templateUrl,omitempty"`

	// Field Mapping Configuration
	// Form field name for the file (default: "file")
	FileField string `json:"fileField,omitempty"`
	// Field name for storing original filename
	NameField string `json:"nameField,omitempty"`
	// Field name for the file value/URL
	ValueField string `json:"valueField,omitempty"`
	// Field name for the file URL
	URLField string `json:"urlField,omitempty"`

	// Data Processing Configuration
	// Whether to convert files to base64 format for small files
	AsBase64 bool `json:"asBase64,omitempty"`
	// Whether to use file blob data directly in form
	AsBlob bool `json:"asBlob,omitempty"`
	// Whether to join multiple values with delimiter
	JoinValues bool `json:"joinValues,omitempty"`
	// Whether to extract value from response
	ExtractValue bool `json:"extractValue,omitempty"`
	// Delimiter for joining values when multiple files
	Delimiter string `json:"delimiter,omitempty"`

	// Chunked Upload Configuration
	// Whether to use chunked upload for large files
	UseChunk any `json:"useChunk,omitempty"` // "auto", true, or false
	// Size of each chunk (bytes, default: 5MB)
	ChunkSize int `json:"chunkSize,omitempty"`
	// Number of concurrent chunk uploads
	Concurrency int `json:"concurrency,omitempty"`
	// API endpoint for starting chunked upload
	StartChunkAPI *APIConfig `json:"startChunkApi,omitempty"`
	// API endpoint for uploading individual chunks
	ChunkAPI *APIConfig `json:"chunkApi,omitempty"`
	// API endpoint for finishing chunked upload
	FinishChunkAPI *APIConfig `json:"finishChunkApi,omitempty"`

	// Image Processing Configuration
	// Image cropping configuration
	Crop *CropConfig `json:"crop,omitempty"`
	// Whether to enable image compression
	Compress bool `json:"compress,omitempty"`
	// Image compression options
	CompressOptions *CompressOptions `json:"compressOptions,omitempty"`
	// Whether to show compression options to user
	ShowCompressOptions bool `json:"showCompressOptions,omitempty"`

	// UI Configuration
	// Text for the upload button
	BtnLabel string `json:"btnLabel,omitempty"`
	// CSS class for the upload button
	BtnClassName string `json:"btnClassName,omitempty"`
	// Whether to enable drag and drop upload
	Drag bool `json:"drag,omitempty"`
	// Information text shown in drag area
	DropInfo string `json:"dropInfo,omitempty"`
	// Custom icons for different file types
	FileIcons map[string]string `json:"fileIcons,omitempty"`
	// Whether to show file type icons
	ShowFileIcons bool `json:"showFileIcons,omitempty"`
	// Documentation text for the component
	Documentation string `json:"documentation,omitempty"`
	// Link to detailed documentation
	DocumentLink string `json:"documentLink,omitempty"`

	// Auto-fill Configuration
	// Simple key-value mapping for auto-fill
	AutoFill map[string]string `json:"autoFill,omitempty"`
	// Whether to trigger auto-fill on initialization
	InitAutoFill bool `json:"initAutoFill,omitempty"`

	// Form Integration Properties
	// Whether to submit form when file changes
	SubmitOnChange bool `json:"submitOnChange,omitempty"`
	// Whether to validate on every change
	ValidateOnChange bool `json:"validateOnChange,omitempty"`
	// Whether to clear value when form item is hidden
	ClearValueOnHidden bool `json:"clearValueOnHidden,omitempty"`
	// Remote validation API
	ValidateAPI any `json:"validateApi,omitempty"`

	// Static Display Properties
	Static bool `json:"static,omitempty"`
	StaticOn string `json:"staticOn,omitempty"`
	StaticPlaceholder string `json:"staticPlaceholder,omitempty"`
	StaticClassName string `json:"staticClassName,omitempty"`
	StaticLabelClassName string `json:"staticLabelClassName,omitempty"`
	StaticInputClassName string `json:"staticInputClassName,omitempty"`
	StaticSchema any `json:"staticSchema,omitempty"`

	// Advanced Properties
	// Description content supporting HTML
	Description string `json:"description,omitempty"`
	// Input hint shown on focus
	Hint string `json:"hint,omitempty"`
	// Custom validation error messages
	ValidationErrors map[string]string `json:"validationErrors,omitempty"`
	// Validation rules configuration
	Validations any `json:"validations,omitempty"`
	// CSS class names for styling
	InputClassName string `json:"inputClassName,omitempty"`
	LabelClassName string `json:"labelClassName,omitempty"`
	DescriptionClassName string `json:"descriptionClassName,omitempty"`

	// Design-time Configuration
	EditorSetting *EditorSetting `json:"editorSetting,omitempty"`
}

// Factory function to create FileControlSchema with sensible defaults
func NewFileControl(name string) *FileControlSchema {
	return &FileControlSchema{
		Type:           "input-file",
		Name:           name,
		BtnLabel:       "Select File",
		AutoUpload:     true,
		Size:           FileControlSizeMD,
		MaxSize:        "10MB",
		FileField:      "file",
		ChunkSize:      5242880, // 5MB default
		UseChunk:       "auto",
		Drag:           true,
		ShowFileIcons:  true,
		Compress:       false,
	}
}

// Factory function to create image file control with cropping
func NewImageFileControl(name string, aspectRatio float64) *FileControlSchema {
	control := NewFileControl(name)
	control.Accept = ".jpg,.jpeg,.png,.gif,image/*"
	control.Crop = &CropConfig{
		AspectRatio:              aspectRatio,
		AutoCrop:                 true,
		AutoCropArea:             0.8,
		Guides:                   true,
		Center:                   true,
		Highlight:                true,
		Background:               true,
		Movable:                  true,
		Rotatable:                true,
		Scalable:                 true,
		Zoomable:                 true,
		ZoomOnWheel:              true,
		WheelZoomRatio:           0.1,
		CropBoxMovable:           true,
		CropBoxResizable:         true,
		ToggleDragModeOnDblclick: true,
		ViewMode:                 CropViewModeRestrict1,
		DragMode:                 CropDragModeCrop,
	}
	control.Compress = true
	control.CompressOptions = &CompressOptions{
		MaxWidth:    1920,
		MaxHeight:   1080,
		Quality:     0.8,
		ConvertSize: 1024 * 1024, // 1MB
		MimeType:    "image/jpeg",
	}
	return control
}

// Factory function to create document file control  
func NewDocumentFileControl(name string) *FileControlSchema {
	control := NewFileControl(name)
	control.Accept = ".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.csv"
	control.Multiple = true
	control.MaxLength = 10
	control.MaxSize = "50MB"
	return control
}

// Validation function for FileControlSchema
func (f *FileControlSchema) Validate() error {
	if f.Type != "input-file" {
		return fmt.Errorf("invalid file control type: %s, must be 'input-file'", f.Type)
	}
	if f.Name == "" {
		return fmt.Errorf("file control name is required")
	}
	
	// Validate enum values
	if f.Size != "" &&
		f.Size != FileControlSizeXS &&
		f.Size != FileControlSizeSM &&
		f.Size != FileControlSizeMD &&
		f.Size != FileControlSizeLG &&
		f.Size != FileControlSizeFull {
		return fmt.Errorf("invalid size: %s", f.Size)
	}
	
	// Validate capture mode if specified
	if f.Capture != "" &&
		f.Capture != FileCaptureUser &&
		f.Capture != FileCaptureEnvironment &&
		f.Capture != FileCaptureCamera &&
		f.Capture != FileCaptureCamcorder &&
		f.Capture != FileCaptureFile {
		return fmt.Errorf("invalid capture mode: %s", f.Capture)
	}
	
	// Validate crop configuration if present
	if f.Crop != nil {
		if f.Crop.ViewMode < CropViewModeRestrict0 || f.Crop.ViewMode > CropViewModeRestrict3 {
			return fmt.Errorf("invalid crop viewMode: %d, must be 0-3", f.Crop.ViewMode)
		}
		if f.Crop.DragMode != "" &&
			f.Crop.DragMode != CropDragModeCrop &&
			f.Crop.DragMode != CropDragModeMove &&
			f.Crop.DragMode != CropDragModeNone {
			return fmt.Errorf("invalid crop dragMode: %s", f.Crop.DragMode)
		}
		if f.Crop.ResizeQuality < 0 || f.Crop.ResizeQuality > 1 {
			return fmt.Errorf("invalid crop resizeQuality: %f, must be between 0 and 1", f.Crop.ResizeQuality)
		}
		if f.Crop.AutoCropArea < 0 || f.Crop.AutoCropArea > 1 {
			return fmt.Errorf("invalid crop autoCropArea: %f, must be between 0 and 1", f.Crop.AutoCropArea)
		}
	}
	
	// Validate compression options if present
	if f.CompressOptions != nil {
		if f.CompressOptions.Quality < 0 || f.CompressOptions.Quality > 1 {
			return fmt.Errorf("invalid compression quality: %f, must be between 0 and 1", f.CompressOptions.Quality)
		}
	}
	
	return nil
}

// ToJSON converts FileControlSchema to JSON string
func (f *FileControlSchema) ToJSON() (string, error) {
	if err := f.Validate(); err != nil {
		return "", err
	}
	
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal FileControlSchema to JSON: %w", err)
	}
	return string(data), nil
}

// FromJSON creates FileControlSchema from JSON string
func FileControlFromJSON(jsonData string) (*FileControlSchema, error) {
	var config FileControlSchema
	err := json.Unmarshal([]byte(jsonData), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to FileControlSchema: %w", err)
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	return &config, nil
}