# Input Signature Component

**FILE PURPOSE**: Digital signature capture and management control  
**SCOPE**: Signature capture, digital signing, document approval, and identity verification  
**TARGET AUDIENCE**: Developers implementing digital signatures, document approval workflows, and identity verification features

## 📋 Component Overview

Input Signature provides comprehensive digital signature functionality with drawing pad, image upload, text signatures, and verification features. Essential for document signing and approval workflows in ERP systems.

### Schema Reference
- **Primary Schema**: `InputSignatureSchema.json`
- **Related Schemas**: `BaseApiObject.json`, `ImageControlSchema.json`
- **Base Interface**: Form input control for digital signature capture

## Basic Usage

```json
{
    "type": "input-signature",
    "name": "approval_signature",
    "label": "Approval Signature",
    "placeholder": "Sign here..."
}
```

## Go Type Definition

```go
type InputSignatureProps struct {
    Type               string              `json:"type"`
    Name               string              `json:"name"`
    Label              interface{}         `json:"label"`
    Placeholder        string              `json:"placeholder"`
    Value              interface{}         `json:"value"`
    Width              int                 `json:"width"`             // Canvas width
    Height             int                 `json:"height"`            // Canvas height
    PenColor           string              `json:"penColor"`          // Stroke color
    PenWidth           int                 `json:"penWidth"`          // Stroke width
    BackgroundColor    string              `json:"backgroundColor"`   // Canvas background
    ShowBorder         bool                `json:"showBorder"`        // Canvas border
    ClearButton        bool                `json:"clearButton"`       // Clear button
    UndoButton         bool                `json:"undoButton"`        // Undo button
    SaveFormat         string              `json:"saveFormat"`        // Output format
    Quality            float64             `json:"quality"`           // Image quality
    Embed              bool                `json:"embed"`             // Embedded mode
    AllowUpload        bool                `json:"allowUpload"`       // Upload signature image
    AllowText          bool                `json:"allowText"`         // Text signature
    FontFamily         string              `json:"fontFamily"`        // Text font
    FontSize           int                 `json:"fontSize"`          // Text size
    RequireSignature   bool                `json:"requireSignature"`  // Signature required
    VerifySignature    bool                `json:"verifySignature"`   // Verify against stored
    SignerInfo         interface{}         `json:"signerInfo"`        // Signer details
    Timestamp          bool                `json:"timestamp"`         // Add timestamp
    FileApi            interface{}         `json:"fileApi"`           // Upload endpoint
    AutoUpload         bool                `json:"autoUpload"`        // Auto upload
    MaxFileSize        int                 `json:"maxFileSize"`       // File size limit
}
```

## Essential Variants

### Basic Signature Pad
```json
{
    "type": "input-signature",
    "name": "document_signature",
    "label": "Document Signature",
    "placeholder": "Please sign here...",
    "width": 400,
    "height": 200,
    "penColor": "#000000",
    "penWidth": 2,
    "clearButton": true,
    "saveFormat": "image/png"
}
```

### Contract Approval Signature
```json
{
    "type": "input-signature",
    "name": "contract_approval",
    "label": "Contract Approval",
    "placeholder": "Sign to approve contract...",
    "width": 500,
    "height": 250,
    "penColor": "#0066cc",
    "penWidth": 3,
    "showBorder": true,
    "clearButton": true,
    "undoButton": true,
    "timestamp": true,
    "requireSignature": true
}
```

### Multi-Method Signature
```json
{
    "type": "input-signature",
    "name": "flexible_signature",
    "label": "Digital Signature",
    "placeholder": "Draw, type, or upload signature...",
    "allowUpload": true,
    "allowText": true,
    "fontFamily": "Brush Script MT",
    "fontSize": 24,
    "clearButton": true,
    "autoUpload": true,
    "fileApi": "/api/signatures/upload"
}
```

### Verification Signature
```json
{
    "type": "input-signature",
    "name": "verification_signature",
    "label": "Verification Signature",
    "placeholder": "Sign to verify identity...",
    "verifySignature": true,
    "signerInfo": {
        "name": "${signer_name}",
        "employee_id": "${employee_id}"
    },
    "timestamp": true,
    "requireSignature": true
}
```

## Real-World Use Cases

### Document Approval Workflow
```json
{
    "type": "input-signature",
    "name": "approval_signature",
    "label": "Approval Signature",
    "placeholder": "Sign to approve this document...",
    "width": 450,
    "height": 200,
    "penColor": "#0066cc",
    "penWidth": 2,
    "backgroundColor": "#f8f9fa",
    "showBorder": true,
    "clearButton": true,
    "undoButton": true,
    "saveFormat": "image/png",
    "quality": 0.9,
    "timestamp": true,
    "requireSignature": true,
    "signerInfo": {
        "approver_name": "${approver_name}",
        "approver_title": "${approver_title}",
        "employee_id": "${employee_id}",
        "department": "${department}"
    },
    "fileApi": "/api/documents/signatures/upload",
    "autoUpload": true,
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "Approval signature is required to proceed"
    },
    "hint": "Your signature confirms approval of this document"
}
```

### Purchase Order Authorization
```json
{
    "type": "input-signature",
    "name": "po_authorization_signature",
    "label": "Purchase Order Authorization",
    "placeholder": "Authorize this purchase order...",
    "width": 500,
    "height": 250,
    "penColor": "#dc3545",
    "penWidth": 3,
    "showBorder": true,
    "clearButton": true,
    "undoButton": true,
    "timestamp": true,
    "requireSignature": true,
    "allowText": true,
    "fontFamily": "Times New Roman",
    "fontSize": 20,
    "signerInfo": {
        "authorizer_name": "${authorizer_name}",
        "authorization_level": "${auth_level}",
        "budget_limit": "${budget_limit}",
        "timestamp": "${current_timestamp}"
    },
    "autoUpload": true,
    "fileApi": "/api/purchase-orders/signatures/upload",
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "Authorization signature is required for purchase orders"
    }
}
```

### Employee Onboarding Signature
```json
{
    "type": "input-signature",
    "name": "onboarding_signature",
    "label": "Employee Signature",
    "placeholder": "Sign to confirm receipt of handbook...",
    "width": 400,
    "height": 180,
    "penColor": "#28a745",
    "penWidth": 2,
    "clearButton": true,
    "timestamp": true,
    "requireSignature": true,
    "allowUpload": true,
    "maxFileSize": 1048576,
    "signerInfo": {
        "employee_name": "${employee_name}",
        "employee_id": "${employee_id}",
        "start_date": "${start_date}",
        "department": "${department}",
        "position": "${position}"
    },
    "autoUpload": true,
    "fileApi": "/api/hr/onboarding/signatures",
    "hint": "Sign to acknowledge receipt of employee handbook and policies"
}
```

### Quality Control Sign-off
```json
{
    "type": "input-signature",
    "name": "qc_signature",
    "label": "Quality Control Sign-off",
    "placeholder": "Sign to certify quality inspection...",
    "width": 400,
    "height": 200,
    "penColor": "#fd7e14",
    "penWidth": 2,
    "showBorder": true,
    "clearButton": true,
    "undoButton": true,
    "timestamp": true,
    "requireSignature": true,
    "signerInfo": {
        "inspector_name": "${inspector_name}",
        "inspector_id": "${inspector_id}",
        "certification_level": "${cert_level}",
        "inspection_date": "${inspection_date}",
        "batch_number": "${batch_number}"
    },
    "verifySignature": true,
    "autoUpload": true,
    "fileApi": "/api/quality/signatures/upload"
}
```

### Contract Signing
```json
{
    "type": "input-signature",
    "name": "contract_signature",
    "label": "Contract Signature",
    "placeholder": "Sign to execute contract...",
    "width": 500,
    "height": 250,
    "penColor": "#6f42c1",
    "penWidth": 3,
    "backgroundColor": "#ffffff",
    "showBorder": true,
    "clearButton": true,
    "undoButton": true,
    "timestamp": true,
    "requireSignature": true,
    "allowText": true,
    "allowUpload": true,
    "fontFamily": "Brush Script MT",
    "fontSize": 22,
    "signerInfo": {
        "signatory_name": "${signatory_name}",
        "signatory_title": "${signatory_title}",
        "company_name": "${company_name}",
        "contract_number": "${contract_number}",
        "contract_value": "${contract_value}"
    },
    "autoUpload": true,
    "fileApi": "/api/contracts/signatures/upload",
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "Contract signature is required to execute agreement"
    }
}
```

### Timesheet Approval
```json
{
    "type": "input-signature",
    "name": "timesheet_approval_signature",
    "label": "Supervisor Approval",
    "placeholder": "Sign to approve timesheet...",
    "width": 350,
    "height": 150,
    "penColor": "#20c997",
    "penWidth": 2,
    "clearButton": true,
    "timestamp": true,
    "requireSignature": true,
    "signerInfo": {
        "supervisor_name": "${supervisor_name}",
        "supervisor_id": "${supervisor_id}",
        "employee_name": "${employee_name}",
        "pay_period": "${pay_period}",
        "total_hours": "${total_hours}"
    },
    "autoUpload": true,
    "fileApi": "/api/timesheets/approvals/signatures"
}
```

### Delivery Confirmation
```json
{
    "type": "input-signature",
    "name": "delivery_signature",
    "label": "Delivery Confirmation",
    "placeholder": "Sign to confirm receipt...",
    "width": 400,
    "height": 180,
    "penColor": "#17a2b8",
    "penWidth": 2,
    "clearButton": true,
    "timestamp": true,
    "requireSignature": true,
    "allowText": true,
    "fontFamily": "Arial",
    "fontSize": 18,
    "signerInfo": {
        "recipient_name": "${recipient_name}",
        "delivery_address": "${delivery_address}",
        "order_number": "${order_number}",
        "delivery_date": "${delivery_date}",
        "items_count": "${items_count}"
    },
    "autoUpload": true,
    "fileApi": "/api/deliveries/signatures/upload",
    "hint": "Sign to confirm receipt of delivered items"
}
```

### Maintenance Work Order
```json
{
    "type": "input-signature",
    "name": "maintenance_signature",
    "label": "Maintenance Completion",
    "placeholder": "Sign to confirm work completion...",
    "width": 450,
    "height": 200,
    "penColor": "#e83e8c",
    "penWidth": 2,
    "showBorder": true,
    "clearButton": true,
    "undoButton": true,
    "timestamp": true,
    "requireSignature": true,
    "signerInfo": {
        "technician_name": "${technician_name}",
        "technician_id": "${technician_id}",
        "work_order_number": "${work_order_number}",
        "asset_tag": "${asset_tag}",
        "completion_date": "${completion_date}",
        "work_performed": "${work_description}"
    },
    "autoUpload": true,
    "fileApi": "/api/maintenance/signatures/upload"
}
```

### Financial Transaction Authorization
```json
{
    "type": "input-signature",
    "name": "financial_authorization",
    "label": "Financial Authorization",
    "placeholder": "Authorize this transaction...",
    "width": 500,
    "height": 250,
    "penColor": "#dc3545",
    "penWidth": 3,
    "backgroundColor": "#fff3cd",
    "showBorder": true,
    "clearButton": true,
    "undoButton": true,
    "timestamp": true,
    "requireSignature": true,
    "verifySignature": true,
    "signerInfo": {
        "authorizer_name": "${authorizer_name}",
        "authorization_level": "${auth_level}",
        "transaction_amount": "${transaction_amount}",
        "transaction_type": "${transaction_type}",
        "account_number": "${account_number}"
    },
    "autoUpload": true,
    "fileApi": "/api/financial/signatures/upload",
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "Financial authorization signature is required"
    }
}
```

### Training Completion Certificate
```json
{
    "type": "input-signature",
    "name": "training_completion_signature",
    "label": "Training Completion",
    "placeholder": "Sign to confirm training completion...",
    "width": 400,
    "height": 180,
    "penColor": "#6610f2",
    "penWidth": 2,
    "clearButton": true,
    "timestamp": true,
    "requireSignature": true,
    "allowText": true,
    "fontFamily": "Georgia",
    "fontSize": 20,
    "signerInfo": {
        "trainee_name": "${trainee_name}",
        "employee_id": "${employee_id}",
        "training_course": "${course_name}",
        "completion_date": "${completion_date}",
        "instructor_name": "${instructor_name}",
        "certificate_number": "${certificate_number}"
    },
    "autoUpload": true,
    "fileApi": "/api/training/certificates/signatures"
}
```

This component provides essential digital signature functionality for ERP systems requiring document approval workflows, identity verification, and secure transaction authorization.