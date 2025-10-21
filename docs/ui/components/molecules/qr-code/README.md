# QR Code Component

**FILE PURPOSE**: QR code generation and scanning control for data encoding and quick access  
**SCOPE**: QR code generation, data encoding, mobile integration, and quick access functionality  
**TARGET AUDIENCE**: Developers implementing QR code features, mobile integration, and quick access systems

## 📋 Component Overview

QR Code provides comprehensive QR code functionality with generation, customization, scanning integration, and data encoding. Essential for mobile-friendly ERP features and quick access systems.

### Schema Reference
- **Primary Schema**: `QRCodeSchema.json`
- **Related Schemas**: `QRCodeImageSettings.json`, `BaseApiObject.json`
- **Base Interface**: Display and generation component for QR codes

## Basic Usage

```json
{
    "type": "qr-code",
    "name": "asset_qr",
    "value": "ASSET-12345",
    "codeSize": 128
}
```

## Go Type Definition

```go
type QRCodeProps struct {
    Type               string              `json:"type"`
    Name               string              `json:"name"`
    Value              string              `json:"value"`             // Data to encode
    CodeSize           int                 `json:"codeSize"`          // QR code size (px)
    Level              string              `json:"level"`             // Error correction level
    BgColor            string              `json:"bgColor"`           // Background color
    FgColor            string              `json:"fgColor"`           // Foreground color
    ImageSettings      interface{}         `json:"imageSettings"`     // Logo/image overlay
    QrcodeClassName    string              `json:"qrcodeClassName"`   // CSS classes
    Placeholder        string              `json:"placeholder"`       // Loading placeholder
    ShowValue          bool                `json:"showValue"`         // Show encoded value
    ValuePosition      string              `json:"valuePosition"`     // Value display position
    DownloadButton     bool                `json:"downloadButton"`    // Download button
    DownloadName       string              `json:"downloadName"`      // Download filename
    PrintButton        bool                `json:"printButton"`       // Print button
    CopyButton         bool                `json:"copyButton"`        // Copy button
    ShareButton        bool                `json:"shareButton"`       // Share button
    RefreshButton      bool                `json:"refreshButton"`     // Refresh button
    AutoRefresh        int                 `json:"autoRefresh"`       // Auto refresh interval
    ScanIntegration    bool                `json:"scanIntegration"`   // Enable scanning
    ScanCallback       string              `json:"scanCallback"`      // Scan callback
    ApiEndpoint        interface{}         `json:"apiEndpoint"`       // Data source API
    Template           string              `json:"template"`          // Data template
    Responsive         bool                `json:"responsive"`        // Responsive sizing
}
```

## Essential Variants

### Basic QR Code
```json
{
    "type": "qr-code",
    "name": "simple_qr",
    "value": "https://example.com/item/12345",
    "codeSize": 128,
    "level": "M",
    "showValue": true
}
```

### Branded QR Code
```json
{
    "type": "qr-code",
    "name": "branded_qr",
    "value": "${product_url}",
    "codeSize": 200,
    "level": "H",
    "bgColor": "#ffffff",
    "fgColor": "#0066cc",
    "imageSettings": {
        "src": "/images/company-logo.png",
        "width": 40,
        "height": 40,
        "excavate": true
    },
    "downloadButton": true
}
```

### Interactive QR Code
```json
{
    "type": "qr-code",
    "name": "interactive_qr",
    "value": "${dynamic_url}",
    "codeSize": 150,
    "showValue": true,
    "valuePosition": "bottom",
    "downloadButton": true,
    "printButton": true,
    "copyButton": true,
    "refreshButton": true
}
```

### Dynamic QR Code
```json
{
    "type": "qr-code",
    "name": "dynamic_qr",
    "apiEndpoint": "/api/qr-data/${item_id}",
    "codeSize": 160,
    "autoRefresh": 30000,
    "template": "https://app.company.com/item/${id}?token=${access_token}",
    "responsive": true
}
```

## Real-World Use Cases

### Asset Tracking QR Code
```json
{
    "type": "qr-code",
    "name": "asset_tracking_qr",
    "value": "https://erp.company.com/assets/${asset_id}?action=view&token=${access_token}",
    "codeSize": 150,
    "level": "H",
    "bgColor": "#ffffff",
    "fgColor": "#000000",
    "imageSettings": {
        "src": "/images/asset-icon.png",
        "width": 30,
        "height": 30,
        "excavate": true
    },
    "showValue": false,
    "downloadButton": true,
    "downloadName": "asset-${asset_tag}-qr",
    "printButton": true,
    "copyButton": true,
    "qrcodeClassName": "asset-qr-code",
    "placeholder": "Generating asset QR code...",
    "responsive": true,
    "hint": "Scan to view asset details and maintenance history"
}
```

### Product Information QR Code
```json
{
    "type": "qr-code",
    "name": "product_info_qr",
    "template": "https://products.company.com/${product_sku}?lang=${user_language}&ref=qr",
    "value": "${generated_product_url}",
    "codeSize": 200,
    "level": "H",
    "bgColor": "#f8f9fa",
    "fgColor": "#0066cc",
    "imageSettings": {
        "src": "${product_thumbnail}",
        "width": 50,
        "height": 50,
        "excavate": true
    },
    "showValue": true,
    "valuePosition": "bottom",
    "downloadButton": true,
    "downloadName": "product-${product_sku}-qr",
    "printButton": true,
    "shareButton": true,
    "apiEndpoint": "/api/products/${product_id}/qr-data",
    "autoRefresh": 60000,
    "responsive": true
}
```

### Employee ID QR Code
```json
{
    "type": "qr-code",
    "name": "employee_id_qr",
    "value": "EMP:${employee_id}|NAME:${employee_name}|DEPT:${department}|EXP:${expiry_date}",
    "codeSize": 180,
    "level": "H",
    "bgColor": "#ffffff",
    "fgColor": "#2c3e50",
    "imageSettings": {
        "src": "${employee_photo}",
        "width": 40,
        "height": 40,
        "excavate": true,
        "rounded": true
    },
    "showValue": false,
    "downloadButton": true,
    "downloadName": "employee-${employee_id}-badge",
    "printButton": true,
    "qrcodeClassName": "employee-badge-qr",
    "scanIntegration": true,
    "scanCallback": "handleEmployeeScan"
}
```

### Inventory Item QR Code
```json
{
    "type": "qr-code",
    "name": "inventory_qr",
    "value": "INV:${item_code}|LOC:${location_code}|QTY:${quantity}|UPD:${last_updated}",
    "codeSize": 120,
    "level": "M",
    "bgColor": "#ffffff",
    "fgColor": "#28a745",
    "showValue": true,
    "valuePosition": "bottom",
    "downloadButton": true,
    "printButton": true,
    "refreshButton": true,
    "apiEndpoint": "/api/inventory/${item_id}/qr-data",
    "autoRefresh": 300000,
    "responsive": true,
    "hint": "Scan for real-time inventory information"
}
```

### Order Tracking QR Code
```json
{
    "type": "qr-code",
    "name": "order_tracking_qr",
    "value": "https://tracking.company.com/order/${order_number}?key=${tracking_key}",
    "codeSize": 160,
    "level": "H",
    "bgColor": "#ffffff",
    "fgColor": "#dc3545",
    "imageSettings": {
        "src": "/images/shipping-icon.png",
        "width": 35,
        "height": 35,
        "excavate": true
    },
    "showValue": true,
    "valuePosition": "bottom",
    "downloadButton": true,
    "downloadName": "order-${order_number}-tracking",
    "shareButton": true,
    "copyButton": true,
    "apiEndpoint": "/api/orders/${order_id}/tracking-qr",
    "responsive": true
}
```

### Meeting Room QR Code
```json
{
    "type": "qr-code",
    "name": "meeting_room_qr",
    "value": "https://booking.company.com/rooms/${room_id}?action=book",
    "codeSize": 180,
    "level": "H",
    "bgColor": "#f0f8ff",
    "fgColor": "#0066cc",
    "imageSettings": {
        "src": "/images/meeting-room-icon.png",
        "width": 40,
        "height": 40,
        "excavate": true
    },
    "showValue": true,
    "valuePosition": "bottom",
    "downloadButton": true,
    "printButton": true,
    "qrcodeClassName": "meeting-room-qr",
    "apiEndpoint": "/api/rooms/${room_id}/booking-qr",
    "autoRefresh": 600000,
    "responsive": true,
    "hint": "Scan to check availability and book this meeting room"
}
```

### Document Access QR Code
```json
{
    "type": "qr-code",
    "name": "document_access_qr",
    "template": "https://docs.company.com/${document_id}?token=${temp_token}&exp=${expiry}",
    "value": "${secure_document_url}",
    "codeSize": 150,
    "level": "H",
    "bgColor": "#ffffff",
    "fgColor": "#6f42c1",
    "imageSettings": {
        "src": "/images/document-icon.png",
        "width": 30,
        "height": 30,
        "excavate": true
    },
    "showValue": false,
    "downloadButton": true,
    "downloadName": "document-${document_id}-access",
    "copyButton": true,
    "refreshButton": true,
    "apiEndpoint": "/api/documents/${document_id}/access-qr",
    "autoRefresh": 900000,
    "responsive": true
}
```

### Event Check-in QR Code
```json
{
    "type": "qr-code",
    "name": "event_checkin_qr",
    "value": "EVENT:${event_id}|TICKET:${ticket_number}|ATTENDEE:${attendee_id}|VALID:${valid_until}",
    "codeSize": 200,
    "level": "H",
    "bgColor": "#ffffff",
    "fgColor": "#fd7e14",
    "imageSettings": {
        "src": "${event_logo}",
        "width": 50,
        "height": 50,
        "excavate": true
    },
    "showValue": true,
    "valuePosition": "bottom",
    "downloadButton": true,
    "downloadName": "event-${event_id}-ticket",
    "printButton": true,
    "shareButton": true,
    "qrcodeClassName": "event-ticket-qr",
    "scanIntegration": true,
    "scanCallback": "handleEventCheckin",
    "responsive": true
}
```

### Equipment Manual QR Code
```json
{
    "type": "qr-code",
    "name": "equipment_manual_qr",
    "value": "https://manuals.company.com/${equipment_model}?serial=${serial_number}&lang=${language}",
    "codeSize": 120,
    "level": "M",
    "bgColor": "#ffffff",
    "fgColor": "#17a2b8",
    "imageSettings": {
        "src": "/images/manual-icon.png",
        "width": 25,
        "height": 25,
        "excavate": true
    },
    "showValue": false,
    "downloadButton": true,
    "printButton": true,
    "qrcodeClassName": "equipment-manual-qr",
    "responsive": true,
    "hint": "Scan for equipment manual and troubleshooting guides"
}
```

### Payment Link QR Code
```json
{
    "type": "qr-code",
    "name": "payment_qr",
    "template": "https://pay.company.com/invoice/${invoice_id}?amount=${amount}&currency=${currency}&exp=${expiry}",
    "value": "${secure_payment_url}",
    "codeSize": 180,
    "level": "H",
    "bgColor": "#ffffff",
    "fgColor": "#28a745",
    "imageSettings": {
        "src": "/images/payment-icon.png",
        "width": 40,
        "height": 40,
        "excavate": true
    },
    "showValue": true,
    "valuePosition": "bottom",
    "downloadButton": true,
    "downloadName": "payment-${invoice_number}",
    "shareButton": true,
    "copyButton": true,
    "apiEndpoint": "/api/invoices/${invoice_id}/payment-qr",
    "autoRefresh": 1800000,
    "responsive": true
}
```

This component provides essential QR code functionality for ERP systems requiring mobile integration, quick access features, and efficient data encoding for various business processes.