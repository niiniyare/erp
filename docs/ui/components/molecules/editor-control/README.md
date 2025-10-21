# Editor Control Component

**FILE PURPOSE**: Advanced rich text editor control for content creation and editing  
**SCOPE**: Rich text editing, HTML content creation, document editing, and content management  
**TARGET AUDIENCE**: Developers implementing content creation features, document editing, and rich text input

## 📋 Component Overview

Editor Control provides comprehensive rich text editing functionality with formatting tools, media insertion, collaborative features, and content management. Essential for content creation and document editing in ERP systems.

### Schema Reference
- **Primary Schema**: `EditorControlSchema.json`
- **Related Schemas**: `BaseApiObject.json`, `FileControlSchema.json`
- **Base Interface**: Form input control for rich text editing

## Basic Usage

```json
{
    "type": "editor",
    "name": "document_content",
    "label": "Document Content",
    "placeholder": "Enter document content..."
}
```

## Go Type Definition

```go
type EditorControlProps struct {
    Type               string              `json:"type"`
    Name               string              `json:"name"`
    Label              interface{}         `json:"label"`
    Placeholder        string              `json:"placeholder"`
    Value              string              `json:"value"`
    Language           string              `json:"language"`          // Content language
    Size               string              `json:"size"`              // Editor size
    Options            interface{}         `json:"options"`           // Editor configuration
    Buttons            interface{}         `json:"buttons"`           // Toolbar buttons
    ToolbarMode        string              `json:"toolbarMode"`       // Toolbar display mode
    AllowImageUpload   bool                `json:"allowImageUpload"`  // Image upload
    AllowFileUpload    bool                `json:"allowFileUpload"`   // File upload
    FileUploadApi      interface{}         `json:"fileUploadApi"`     // Upload endpoint
    MaxFileSize        int                 `json:"maxFileSize"`       // File size limit
    AcceptedFiles      []string            `json:"acceptedFiles"`     // Allowed file types
    AutoSave           bool                `json:"autoSave"`          // Auto-save content
    AutoSaveInterval   int                 `json:"autoSaveInterval"`  // Auto-save frequency
    WordWrap           bool                `json:"wordWrap"`          // Word wrapping
    ShowWordCount      bool                `json:"showWordCount"`     // Word count display
    MaxLength          int                 `json:"maxLength"`         // Character limit
    SpellCheck         bool                `json:"spellCheck"`        // Spell checking
    Collaboration      bool                `json:"collaboration"`     // Collaborative editing
    Templates          interface{}         `json:"templates"`         // Content templates
    CustomCSS          string              `json:"customCSS"`         // Custom styling
    Plugins            []string            `json:"plugins"`           // Editor plugins
}
```

## Essential Variants

### Basic Rich Text Editor
```json
{
    "type": "editor",
    "name": "content_editor",
    "label": "Content",
    "placeholder": "Enter content...",
    "size": "md",
    "buttons": ["bold", "italic", "underline", "link", "list"],
    "showWordCount": true,
    "spellCheck": true
}
```

### Document Editor
```json
{
    "type": "editor",
    "name": "document_editor",
    "label": "Document Content",
    "placeholder": "Create your document...",
    "size": "lg",
    "buttons": ["source", "bold", "italic", "underline", "strikethrough", "link", "image", "table", "list", "outdent", "indent"],
    "allowImageUpload": true,
    "fileUploadApi": "/api/documents/upload",
    "autoSave": true,
    "autoSaveInterval": 30000
}
```

### Email Composer
```json
{
    "type": "editor",
    "name": "email_content",
    "label": "Email Content",
    "placeholder": "Compose your email...",
    "size": "lg",
    "buttons": ["bold", "italic", "underline", "link", "image", "emoji", "template"],
    "templates": "/api/email-templates",
    "allowImageUpload": true,
    "maxLength": 50000
}
```

### Collaborative Editor
```json
{
    "type": "editor",
    "name": "shared_document",
    "label": "Shared Document",
    "placeholder": "Collaborate on document...",
    "size": "xl",
    "collaboration": true,
    "autoSave": true,
    "autoSaveInterval": 10000,
    "showWordCount": true,
    "buttons": ["bold", "italic", "link", "comment", "suggest", "history"]
}
```

## Real-World Use Cases

### Policy Document Editor
```json
{
    "type": "editor",
    "name": "policy_content",
    "label": "Policy Content",
    "placeholder": "Enter policy content...",
    "size": "lg",
    "language": "en",
    "buttons": [
        "source", "bold", "italic", "underline", "strikethrough",
        "link", "unlink", "image", "table", "list", "outdent", "indent",
        "align", "hr", "superscript", "subscript", "print"
    ],
    "toolbarMode": "floating",
    "allowImageUpload": true,
    "allowFileUpload": true,
    "fileUploadApi": "/api/policies/attachments/upload",
    "maxFileSize": 10485760,
    "acceptedFiles": [".jpg", ".png", ".gif", ".pdf", ".doc", ".docx"],
    "autoSave": true,
    "autoSaveInterval": 60000,
    "showWordCount": true,
    "maxLength": 500000,
    "spellCheck": true,
    "wordWrap": true,
    "templates": [
        {"name": "Company Policy Template", "content": "<h1>Policy Title</h1><h2>Purpose</h2><p>...</p>"},
        {"name": "Procedure Template", "content": "<h1>Procedure Name</h1><h2>Steps</h2><ol><li>...</li></ol>"}
    ],
    "customCSS": ".company-policy { font-family: 'Times New Roman'; line-height: 1.6; }",
    "plugins": ["table", "link", "image", "wordcount", "autosave"],
    "required": true
}
```

### Product Description Editor
```json
{
    "type": "editor",
    "name": "product_description",
    "label": "Product Description",
    "placeholder": "Describe your product features and benefits...",
    "size": "md",
    "buttons": [
        "bold", "italic", "underline", "link", "unlink",
        "image", "list", "table", "emoji", "color"
    ],
    "allowImageUpload": true,
    "fileUploadApi": "/api/products/images/upload",
    "maxFileSize": 5242880,
    "acceptedFiles": [".jpg", ".jpeg", ".png", ".webp"],
    "showWordCount": true,
    "maxLength": 10000,
    "spellCheck": true,
    "templates": [
        {"name": "Feature List", "content": "<h3>Key Features</h3><ul><li>Feature 1</li><li>Feature 2</li></ul>"},
        {"name": "Benefits", "content": "<h3>Benefits</h3><p>This product provides...</p>"}
    ],
    "autoSave": true,
    "autoSaveInterval": 30000,
    "hint": "Create compelling product descriptions with rich formatting"
}
```

### Training Material Editor
```json
{
    "type": "editor",
    "name": "training_content",
    "label": "Training Content",
    "placeholder": "Create training materials...",
    "size": "lg",
    "buttons": [
        "source", "bold", "italic", "underline", "link", "image",
        "video", "table", "list", "outdent", "indent", "quote",
        "code", "highlight", "align"
    ],
    "allowImageUpload": true,
    "allowFileUpload": true,
    "fileUploadApi": "/api/training/media/upload",
    "maxFileSize": 52428800,
    "acceptedFiles": [".jpg", ".png", ".gif", ".mp4", ".pdf", ".ppt", ".pptx"],
    "autoSave": true,
    "autoSaveInterval": 45000,
    "showWordCount": true,
    "spellCheck": true,
    "collaboration": true,
    "templates": [
        {"name": "Lesson Plan", "content": "<h1>Lesson Title</h1><h2>Objectives</h2><p>...</p><h2>Content</h2><p>...</p>"},
        {"name": "Assessment", "content": "<h1>Assessment</h1><h2>Questions</h2><ol><li>Question 1</li></ol>"}
    ],
    "plugins": ["table", "link", "image", "video", "collaboration", "wordcount"]
}
```

### Internal Communication Editor
```json
{
    "type": "editor",
    "name": "announcement_content",
    "label": "Announcement Content",
    "placeholder": "Write your announcement...",
    "size": "md",
    "buttons": [
        "bold", "italic", "underline", "link", "image",
        "list", "emoji", "mention", "highlight", "align"
    ],
    "allowImageUpload": true,
    "fileUploadApi": "/api/announcements/images/upload",
    "maxFileSize": 2097152,
    "acceptedFiles": [".jpg", ".png", ".gif"],
    "showWordCount": true,
    "maxLength": 5000,
    "spellCheck": true,
    "templates": [
        {"name": "Company Update", "content": "<h2>Company Update</h2><p>We're excited to announce...</p>"},
        {"name": "Policy Change", "content": "<h2>Policy Update</h2><p>Effective [date], the following changes...</p>"}
    ],
    "autoSave": true,
    "autoSaveInterval": 20000
}
```

### Contract Clause Editor
```json
{
    "type": "editor",
    "name": "contract_clauses",
    "label": "Contract Clauses",
    "placeholder": "Enter contract clauses and terms...",
    "size": "lg",
    "buttons": [
        "source", "bold", "italic", "underline", "link",
        "table", "list", "outdent", "indent", "align",
        "superscript", "subscript", "strikethrough"
    ],
    "showWordCount": true,
    "spellCheck": true,
    "wordWrap": true,
    "autoSave": true,
    "autoSaveInterval": 120000,
    "templates": [
        {"name": "Payment Terms", "content": "<h3>Payment Terms</h3><p>Payment shall be due...</p>"},
        {"name": "Termination Clause", "content": "<h3>Termination</h3><p>Either party may terminate...</p>"},
        {"name": "Liability Limitation", "content": "<h3>Limitation of Liability</h3><p>In no event shall...</p>"}
    ],
    "customCSS": ".contract-text { font-family: 'Times New Roman'; font-size: 12pt; line-height: 1.5; }",
    "collaboration": true,
    "required": true
}
```

### Knowledge Base Article Editor
```json
{
    "type": "editor",
    "name": "kb_article_content",
    "label": "Article Content",
    "placeholder": "Write your knowledge base article...",
    "size": "lg",
    "buttons": [
        "source", "bold", "italic", "underline", "link", "unlink",
        "image", "table", "list", "outdent", "indent", "quote",
        "code", "codeblock", "hr", "align", "toc"
    ],
    "allowImageUpload": true,
    "allowFileUpload": true,
    "fileUploadApi": "/api/knowledge-base/files/upload",
    "maxFileSize": 10485760,
    "acceptedFiles": [".jpg", ".png", ".gif", ".pdf", ".zip"],
    "autoSave": true,
    "autoSaveInterval": 30000,
    "showWordCount": true,
    "spellCheck": true,
    "templates": [
        {"name": "How-to Guide", "content": "<h1>How to [Task]</h1><h2>Prerequisites</h2><p>...</p><h2>Steps</h2><ol><li>Step 1</li></ol>"},
        {"name": "Troubleshooting", "content": "<h1>Troubleshooting [Issue]</h1><h2>Problem</h2><p>...</p><h2>Solution</h2><p>...</p>"},
        {"name": "FAQ", "content": "<h1>Frequently Asked Questions</h1><h3>Question 1</h3><p>Answer...</p>"}
    ],
    "plugins": ["table", "link", "image", "code", "toc", "autosave"]
}
```

### Quality Control Report Editor
```json
{
    "type": "editor",
    "name": "qc_report_content",
    "label": "QC Report Content",
    "placeholder": "Enter quality control findings...",
    "size": "lg",
    "buttons": [
        "bold", "italic", "underline", "link", "image",
        "table", "list", "align", "highlight", "strikethrough"
    ],
    "allowImageUpload": true,
    "fileUploadApi": "/api/quality/reports/images/upload",
    "maxFileSize": 15728640,
    "acceptedFiles": [".jpg", ".png", ".pdf"],
    "showWordCount": true,
    "spellCheck": true,
    "autoSave": true,
    "autoSaveInterval": 60000,
    "templates": [
        {"name": "Inspection Report", "content": "<h1>Inspection Report</h1><h2>Summary</h2><p>...</p><h2>Findings</h2><table><tr><th>Item</th><th>Status</th><th>Notes</th></tr></table>"},
        {"name": "Non-Conformance", "content": "<h1>Non-Conformance Report</h1><h2>Description</h2><p>...</p><h2>Corrective Action</h2><p>...</p>"}
    ],
    "required": true
}
```

### Project Proposal Editor
```json
{
    "type": "editor",
    "name": "proposal_content",
    "label": "Proposal Content",
    "placeholder": "Write your project proposal...",
    "size": "xl",
    "buttons": [
        "source", "bold", "italic", "underline", "link", "image",
        "table", "list", "outdent", "indent", "align", "quote",
        "hr", "pagebreak", "toc", "footnote"
    ],
    "allowImageUpload": true,
    "allowFileUpload": true,
    "fileUploadApi": "/api/proposals/attachments/upload",
    "maxFileSize": 20971520,
    "acceptedFiles": [".jpg", ".png", ".pdf", ".xls", ".xlsx", ".doc", ".docx"],
    "autoSave": true,
    "autoSaveInterval": 90000,
    "showWordCount": true,
    "spellCheck": true,
    "collaboration": true,
    "templates": [
        {"name": "Project Proposal", "content": "<h1>Project Title</h1><h2>Executive Summary</h2><p>...</p><h2>Objectives</h2><p>...</p><h2>Timeline</h2><p>...</p><h2>Budget</h2><p>...</p>"},
        {"name": "Technical Proposal", "content": "<h1>Technical Proposal</h1><h2>Technical Approach</h2><p>...</p><h2>Architecture</h2><p>...</p>"}
    ],
    "wordWrap": true,
    "required": true
}
```

This component provides essential rich text editing functionality for ERP systems requiring content creation, document editing, and collaborative writing capabilities.