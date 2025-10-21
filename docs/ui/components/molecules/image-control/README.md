# Image Control Component

**FILE PURPOSE**: Image upload and management control for media handling  
**SCOPE**: Image upload, preview, editing, gallery management, and media asset handling  
**TARGET AUDIENCE**: Developers implementing image upload features, media management, and visual content handling

## 📋 Component Overview

Image Control provides comprehensive image handling functionality with upload, preview, cropping, resizing, and gallery features. Essential for managing visual content and media assets in ERP systems.

### Schema Reference
- **Primary Schema**: `ImageControlSchema.json`
- **Related Schemas**: `BaseApiObject.json`, `ImageCrop.json`
- **Base Interface**: Form input control for image management

## Basic Usage

```json
{
    "type": "input-image",
    "name": "profile_image",
    "label": "Profile Image",
    "placeholder": "Upload image..."
}
```

## Go Type Definition

```go
type ImageControlProps struct {
    Type               string              `json:"type"`
    Name               string              `json:"name"`
    Label              interface{}         `json:"label"`
    Placeholder        string              `json:"placeholder"`
    Value              interface{}         `json:"value"`
    Src                string              `json:"src"`               // Image source URL
    Alt                string              `json:"alt"`               // Alt text
    DefaultImage       string              `json:"defaultImage"`      // Fallback image
    ImageClassName     string              `json:"imageClassName"`    // Image CSS classes
    ImageMode          string              `json:"imageMode"`         // Display mode
    ThumbMode          string              `json:"thumbMode"`         // Thumbnail mode
    ThumbRatio         string              `json:"thumbRatio"`        // Aspect ratio
    Crop               interface{}         `json:"crop"`              // Crop settings
    CropFormat         string              `json:"cropFormat"`        // Output format
    CropQuality        float64             `json:"cropQuality"`       // Compression quality
    Multiple           bool                `json:"multiple"`          // Multiple images
    MaxLength          int                 `json:"maxLength"`         // Max image count
    MaxSize            int                 `json:"maxSize"`           // File size limit
    Accept             string              `json:"accept"`            // File types
    HideUploadButton   bool                `json:"hideUploadButton"`  // Hide upload UI
    DragTip            string              `json:"dragTip"`           // Drag & drop text
    DropTip            string              `json:"dropTip"`           // Drop zone text
    FileApi            interface{}         `json:"fileApi"`           // Upload endpoint
    UploadType         string              `json:"uploadType"`        // Upload method
    AutoUpload         bool                `json:"autoUpload"`        // Auto upload
    ShowRemoveButton   bool                `json:"showRemoveButton"`  // Remove button
    ShowDownloadButton bool                `json:"showDownloadButton"` // Download button
    Delimiter          string              `json:"delimiter"`         // Value separator
    JoinValues         bool                `json:"joinValues"`        // Join as string
    ExtractValue       bool                `json:"extractValue"`      // Extract value only
}
```

## Essential Variants

### Basic Image Upload
```json
{
    "type": "input-image",
    "name": "product_image",
    "label": "Product Image",
    "placeholder": "Upload product image...",
    "maxSize": 5242880,
    "accept": ".jpg,.jpeg,.png,.webp",
    "autoUpload": true,
    "fileApi": "/api/images/upload"
}
```

### Profile Image with Crop
```json
{
    "type": "input-image",
    "name": "avatar",
    "label": "Profile Picture",
    "placeholder": "Upload your profile picture...",
    "crop": {
        "aspectRatio": 1,
        "rotatable": true,
        "scalable": true,
        "cropFormat": "image/jpeg",
        "cropQuality": 0.9
    },
    "thumbMode": "cover",
    "thumbRatio": "1:1",
    "maxSize": 2097152,
    "autoUpload": true
}
```

### Multiple Image Gallery
```json
{
    "type": "input-image",
    "name": "gallery_images",
    "label": "Image Gallery",
    "placeholder": "Upload multiple images...",
    "multiple": true,
    "maxLength": 10,
    "maxSize": 10485760,
    "accept": ".jpg,.jpeg,.png,.webp",
    "showRemoveButton": true,
    "showDownloadButton": true,
    "dragTip": "Drag images here or click to upload",
    "autoUpload": true
}
```

### Image with Preview Modes
```json
{
    "type": "input-image",
    "name": "banner_image",
    "label": "Banner Image",
    "placeholder": "Upload banner image...",
    "imageMode": "responsive",
    "thumbMode": "contain",
    "thumbRatio": "16:9",
    "crop": {
        "aspectRatio": 16/9,
        "viewMode": 1
    },
    "maxSize": 15728640,
    "autoUpload": true
}
```

## Real-World Use Cases

### Product Image Management
```json
{
    "type": "input-image",
    "name": "product_images",
    "label": "Product Images",
    "placeholder": "Upload product images...",
    "multiple": true,
    "maxLength": 8,
    "maxSize": 5242880,
    "accept": ".jpg,.jpeg,.png,.webp",
    "fileApi": "/api/products/images/upload",
    "autoUpload": true,
    "showRemoveButton": true,
    "showDownloadButton": true,
    "imageMode": "responsive",
    "thumbMode": "cover",
    "thumbRatio": "1:1",
    "crop": {
        "aspectRatio": 1,
        "viewMode": 1,
        "dragMode": "crop",
        "cropFormat": "image/jpeg",
        "cropQuality": 0.85
    },
    "dragTip": "Drag product images here or click to upload",
    "dropTip": "Drop images to upload",
    "hint": "Upload high-quality product images (max 8 images, 5MB each)",
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "At least one product image is required"
    }
}
```

### Employee Profile Photo
```json
{
    "type": "input-image",
    "name": "employee_photo",
    "label": "Employee Photo",
    "placeholder": "Upload employee photo...",
    "maxSize": 2097152,
    "accept": ".jpg,.jpeg,.png",
    "fileApi": "/api/employees/photos/upload",
    "autoUpload": true,
    "defaultImage": "/images/default-avatar.png",
    "imageMode": "responsive",
    "thumbMode": "cover",
    "thumbRatio": "1:1",
    "crop": {
        "aspectRatio": 1,
        "viewMode": 1,
        "dragMode": "crop",
        "autoCropArea": 0.8,
        "restore": false,
        "guides": true,
        "center": true,
        "highlight": false,
        "cropBoxMovable": true,
        "cropBoxResizable": true,
        "toggleDragModeOnDblclick": false
    },
    "cropFormat": "image/jpeg",
    "cropQuality": 0.9,
    "showRemoveButton": true,
    "hint": "Upload a professional headshot (max 2MB)"
}
```

### Company Logo Upload
```json
{
    "type": "input-image",
    "name": "company_logo",
    "label": "Company Logo",
    "placeholder": "Upload company logo...",
    "maxSize": 1048576,
    "accept": ".svg,.png,.jpg,.jpeg",
    "fileApi": "/api/company/logo/upload",
    "autoUpload": true,
    "imageMode": "responsive",
    "thumbMode": "contain",
    "thumbRatio": "2:1",
    "showRemoveButton": true,
    "showDownloadButton": true,
    "hint": "Upload your company logo (SVG or PNG preferred, max 1MB)",
    "autoFill": {
        "api": "/api/company/logo/metadata/${value}",
        "fillMapping": {
            "logo_width": "width",
            "logo_height": "height",
            "logo_format": "format",
            "logo_size": "file_size"
        }
    }
}
```

### Document Attachment Images
```json
{
    "type": "input-image",
    "name": "document_images",
    "label": "Document Images",
    "placeholder": "Upload document images or scans...",
    "multiple": true,
    "maxLength": 20,
    "maxSize": 10485760,
    "accept": ".jpg,.jpeg,.png,.tiff,.pdf",
    "fileApi": "/api/documents/images/upload",
    "autoUpload": true,
    "showRemoveButton": true,
    "showDownloadButton": true,
    "imageMode": "responsive",
    "thumbMode": "contain",
    "thumbRatio": "4:3",
    "dragTip": "Drag document images here",
    "dropTip": "Drop to upload document images",
    "hint": "Upload scanned documents or photos (max 20 files, 10MB each)"
}
```

### Marketing Banner Images
```json
{
    "type": "input-image",
    "name": "marketing_banners",
    "label": "Marketing Banners",
    "placeholder": "Upload banner images...",
    "multiple": true,
    "maxLength": 5,
    "maxSize": 15728640,
    "accept": ".jpg,.jpeg,.png,.webp",
    "fileApi": "/api/marketing/banners/upload",
    "autoUpload": true,
    "imageMode": "responsive",
    "thumbMode": "cover",
    "thumbRatio": "16:9",
    "crop": {
        "aspectRatio": 16/9,
        "viewMode": 1,
        "dragMode": "crop",
        "autoCropArea": 1.0,
        "cropFormat": "image/webp",
        "cropQuality": 0.8
    },
    "showRemoveButton": true,
    "showDownloadButton": true,
    "hint": "Upload banner images in 16:9 aspect ratio (max 5 images, 15MB each)"
}
```

### Asset/Equipment Photos
```json
{
    "type": "input-image",
    "name": "asset_photos",
    "label": "Asset Photos",
    "placeholder": "Upload asset photos...",
    "multiple": true,
    "maxLength": 6,
    "maxSize": 8388608,
    "accept": ".jpg,.jpeg,.png",
    "fileApi": "/api/assets/photos/upload",
    "autoUpload": true,
    "showRemoveButton": true,
    "imageMode": "responsive",
    "thumbMode": "cover",
    "thumbRatio": "4:3",
    "dragTip": "Drag asset photos here",
    "hint": "Upload clear photos of the asset from different angles (max 6 photos, 8MB each)",
    "autoFill": {
        "api": "/api/assets/photos/analyze/${value}",
        "fillMapping": {
            "photo_metadata": "exif_data",
            "photo_location": "gps_coordinates",
            "photo_timestamp": "date_taken"
        }
    }
}
```

### Signature Upload
```json
{
    "type": "input-image",
    "name": "digital_signature",
    "label": "Digital Signature",
    "placeholder": "Upload signature image...",
    "maxSize": 512000,
    "accept": ".png,.jpg,.jpeg",
    "fileApi": "/api/signatures/upload",
    "autoUpload": true,
    "imageMode": "responsive",
    "thumbMode": "contain",
    "thumbRatio": "3:1",
    "crop": {
        "aspectRatio": 3,
        "viewMode": 1,
        "dragMode": "crop",
        "autoCropArea": 0.9,
        "cropFormat": "image/png",
        "cropQuality": 1.0
    },
    "showRemoveButton": true,
    "hint": "Upload a clear signature on white background (max 500KB)"
}
```

### Floor Plan/Layout Images
```json
{
    "type": "input-image",
    "name": "floor_plans",
    "label": "Floor Plans",
    "placeholder": "Upload floor plan images...",
    "multiple": true,
    "maxLength": 3,
    "maxSize": 20971520,
    "accept": ".jpg,.jpeg,.png,.pdf,.dwg",
    "fileApi": "/api/facilities/floor-plans/upload",
    "autoUpload": true,
    "showRemoveButton": true,
    "showDownloadButton": true,
    "imageMode": "responsive",
    "thumbMode": "contain",
    "dragTip": "Drag floor plan images here",
    "hint": "Upload architectural drawings or floor plans (max 3 files, 20MB each)"
}
```

### Quality Control Images
```json
{
    "type": "input-image",
    "name": "qc_images",
    "label": "Quality Control Images",
    "placeholder": "Upload QC inspection images...",
    "multiple": true,
    "maxLength": 12,
    "maxSize": 5242880,
    "accept": ".jpg,.jpeg,.png",
    "fileApi": "/api/quality/images/upload",
    "autoUpload": true,
    "showRemoveButton": true,
    "imageMode": "responsive",
    "thumbMode": "cover",
    "thumbRatio": "4:3",
    "dragTip": "Drag inspection photos here",
    "hint": "Upload detailed photos of inspection points (max 12 images, 5MB each)",
    "autoFill": {
        "api": "/api/quality/images/metadata/${value}",
        "fillMapping": {
            "inspection_timestamp": "date_taken",
            "inspector_device": "device_info",
            "image_resolution": "dimensions"
        }
    }
}
```

### Training Material Images
```json
{
    "type": "input-image",
    "name": "training_images",
    "label": "Training Images",
    "placeholder": "Upload training material images...",
    "multiple": true,
    "maxLength": 15,
    "maxSize": 10485760,
    "accept": ".jpg,.jpeg,.png,.gif,.webp",
    "fileApi": "/api/training/images/upload",
    "autoUpload": true,
    "showRemoveButton": true,
    "showDownloadButton": true,
    "imageMode": "responsive",
    "thumbMode": "contain",
    "thumbRatio": "16:10",
    "dragTip": "Drag training images here",
    "hint": "Upload images for training materials (max 15 images, 10MB each)"
}
```

This component provides essential image management functionality for ERP systems requiring comprehensive media handling, visual content management, and document imaging capabilities.