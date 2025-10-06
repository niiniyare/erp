package handlers

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"strconv"
// 	"time"
//
// 	"github.com/google/uuid"
// 	"github.com/niiniyare/erp/internal/core/abac"
// 	"github.com/niiniyare/erp/internal/core/audit"
// 	"github.com/niiniyare/erp/internal/core/iam"
// 	"github.com/niiniyare/erp/internal/core/tenant"
// 	"github.com/niiniyare/erp/internal/shared/logger"
// 	"github.com/niiniyare/erp/internal/ui/middleware"
// 	"github.com/niiniyare/erp/internal/ui/types"
// )
//
// // InvoicesHandler handles portal invoice operations for client users
// type InvoicesHandler struct {
// 	iamService    iam.Service
// 	abacService   abac.Service
// 	tenantService tenant.Service
// 	auditService  audit.Service
// 	logger        logger.Logger
// }
//
// // NewInvoicesHandler creates a new portal invoices handler
// func NewInvoicesHandler(
// 	iamService iam.Service,
// 	abacService abac.Service,
// 	tenantService tenant.Service,
// 	auditService audit.Service,
// 	logger logger.Logger,
// ) *InvoicesHandler {
// 	return &InvoicesHandler{
// 		iamService:    iamService,
// 		abacService:   abacService,
// 		tenantService: tenantService,
// 		auditService:  auditService,
// 		logger:        logger,
// 	}
// }
//
// // ListInvoices displays the invoice list page for the current client
// func (h *InvoicesHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Ensure user has client access
// 	if err := middleware.RequireRole(ctx, middleware.UIRoleClientUser); err != nil {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Verify client context
// 	if uiCtx.ClientID == uuid.Nil {
// 		http.Error(w, "No client context", http.StatusForbidden)
// 		return
// 	}
//
// 	// Parse query parameters
// 	filters := h.parseInvoiceFilters(r)
//
// 	// Get invoices for this client
// 	invoices, pagination, err := h.getClientInvoices(ctx, uiCtx.ClientID, filters)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to get client invoices", logger.Fields{
// 			"error":     err.Error(),
// 			"user_id":   uiCtx.UserID,
// 			"client_id": uiCtx.ClientID,
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Prepare page data
// 	pageData := &types.PortalInvoicesPageData{
// 		Title:      "My Invoices",
// 		Invoices:   invoices,
// 		Pagination: pagination,
// 		Filters:    filters,
// 		CSRFToken:  uiCtx.CSRFToken,
// 	}
//
// 	// Render invoice list template
// 	// TODO: Create portal invoices template
// 	h.renderInvoicesListTemp(w, pageData)
//
// 	// Log invoice list access
// 	h.logger.InfoContext(ctx, "Portal invoices list accessed", logger.Fields{
// 		"user_id":   uiCtx.UserID,
// 		"client_id": uiCtx.ClientID,
// 		"filters":   filters,
// 	})
// }
//
// // GetInvoicesData returns invoice list data as JSON (for HTMX requests and DataTable)
// func (h *InvoicesHandler) GetInvoicesData(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Ensure client access
// 	if err := middleware.RequireRole(ctx, middleware.UIRoleClientUser); err != nil {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Parse filters
// 	filters := h.parseInvoiceFilters(r)
//
// 	// Get invoices data
// 	invoices, pagination, err := h.getClientInvoices(ctx, uiCtx.ClientID, filters)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to get invoices data", logger.Fields{
// 			"error":     err.Error(),
// 			"user_id":   uiCtx.UserID,
// 			"client_id": uiCtx.ClientID,
// 		})
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Return JSON response
// 	response := map[string]any{
// 		"invoices":   invoices,
// 		"pagination": pagination,
// 		"success":    true,
// 	}
//
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(response)
// }
//
// // ShowInvoiceDetail displays detailed information about an invoice
// func (h *InvoicesHandler) ShowInvoiceDetail(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Ensure client access
// 	if err := middleware.RequireRole(ctx, middleware.UIRoleClientUser); err != nil {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Extract invoice ID from path
// 	invoiceID := r.URL.Path[len("/portal/invoices/"):]
// 	if invoiceID == "" {
// 		http.Error(w, "Invoice ID required", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Get invoice details (with client ownership verification)
// 	invoice, err := h.getClientInvoice(ctx, uiCtx.ClientID, invoiceID)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to get invoice details", logger.Fields{
// 			"error":      err.Error(),
// 			"user_id":    uiCtx.UserID,
// 			"client_id":  uiCtx.ClientID,
// 			"invoice_id": invoiceID,
// 		})
// 		http.Error(w, "Invoice not found", http.StatusNotFound)
// 		return
// 	}
//
// 	// Prepare page data
// 	pageData := &types.PortalInvoiceDetailPageData{
// 		Title:     fmt.Sprintf("Invoice %s", invoice.Number),
// 		Invoice:   invoice,
// 		CSRFToken: uiCtx.CSRFToken,
// 	}
//
// 	// Render invoice detail template
// 	// TODO: Create portal invoice detail template
// 	h.renderInvoiceDetailTemp(w, pageData)
//
// 	// Log invoice view
// 	h.logger.InfoContext(ctx, "Portal invoice viewed", logger.Fields{
// 		"user_id":    uiCtx.UserID,
// 		"client_id":  uiCtx.ClientID,
// 		"invoice_id": invoiceID,
// 	})
// }
//
// // DownloadInvoice serves the invoice PDF for download
// func (h *InvoicesHandler) DownloadInvoice(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Ensure client access
// 	if err := middleware.RequireRole(ctx, middleware.UIRoleClientUser); err != nil {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Extract invoice ID from path
// 	invoiceID := r.URL.Path[len("/portal/invoices/"):]
// 	invoiceID = invoiceID[:len(invoiceID)-len("/download")] // Remove /download suffix
// 	if invoiceID == "" {
// 		http.Error(w, "Invoice ID required", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Verify client owns this invoice
// 	invoice, err := h.getClientInvoice(ctx, uiCtx.ClientID, invoiceID)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to verify invoice ownership for download", logger.Fields{
// 			"error":      err.Error(),
// 			"user_id":    uiCtx.UserID,
// 			"client_id":  uiCtx.ClientID,
// 			"invoice_id": invoiceID,
// 		})
// 		http.Error(w, "Invoice not found", http.StatusNotFound)
// 		return
// 	}
//
// 	// Generate or retrieve PDF
// 	pdfData, err := h.generateInvoicePDF(ctx, invoice)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to generate invoice PDF", logger.Fields{
// 			"error":      err.Error(),
// 			"user_id":    uiCtx.UserID,
// 			"client_id":  uiCtx.ClientID,
// 			"invoice_id": invoiceID,
// 		})
// 		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Set headers for PDF download
// 	filename := fmt.Sprintf("%s.pdf", invoice.Number)
// 	w.Header().Set("Content-Type", "application/pdf")
// 	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
// 	w.Header().Set("Content-Length", strconv.Itoa(len(pdfData)))
//
// 	// Serve PDF content
// 	w.Write(pdfData)
//
// 	// Log invoice download
// 	h.logger.InfoContext(ctx, "Portal invoice downloaded", logger.Fields{
// 		"user_id":    uiCtx.UserID,
// 		"client_id":  uiCtx.ClientID,
// 		"invoice_id": invoiceID,
// 		"filename":   filename,
// 	})
// }
//
// // PayInvoice handles invoice payment processing
// func (h *InvoicesHandler) PayInvoice(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Ensure client access
// 	if err := middleware.RequireRole(ctx, middleware.UIRoleClientUser); err != nil {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}
//
// 	// Parse form data
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, "Invalid form data", http.StatusBadRequest)
// 		return
// 	}
//
// 	invoiceID := r.FormValue("invoice_id")
// 	paymentMethod := r.FormValue("payment_method")
// 	amount := r.FormValue("amount")
//
// 	if invoiceID == "" {
// 		http.Error(w, "Invoice ID required", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Verify client owns this invoice
// 	invoice, err := h.getClientInvoice(ctx, uiCtx.ClientID, invoiceID)
// 	if err != nil {
// 		http.Error(w, "Invoice not found", http.StatusNotFound)
// 		return
// 	}
//
// 	// Process payment
// 	paymentResult, err := h.processInvoicePayment(ctx, invoice, paymentMethod, amount)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to process invoice payment", logger.Fields{
// 			"error":          err.Error(),
// 			"user_id":        uiCtx.UserID,
// 			"client_id":      uiCtx.ClientID,
// 			"invoice_id":     invoiceID,
// 			"amount":         amount,
// 			"payment_method": paymentMethod,
// 		})
//
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(map[string]any{
// 			"success": false,
// 			"error":   "Payment processing failed. Please try again.",
// 		})
// 		return
// 	}
//
// 	// Log successful payment
// 	h.logger.InfoContext(ctx, "Portal invoice payment processed", logger.Fields{
// 		"user_id":        uiCtx.UserID,
// 		"client_id":      uiCtx.ClientID,
// 		"invoice_id":     invoiceID,
// 		"amount":         amount,
// 		"payment_method": paymentMethod,
// 		"payment_id":     paymentResult.PaymentID,
// 	})
//
// 	// Return success response
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]any{
// 		"success":    true,
// 		"message":    "Payment processed successfully",
// 		"payment_id": paymentResult.PaymentID,
// 		"new_status": paymentResult.NewStatus,
// 	})
// }
//
// // BulkInvoiceActions handles bulk operations on invoices (download, pay, etc.)
// func (h *InvoicesHandler) BulkInvoiceActions(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get UI context
// 	uiCtx, ok := middleware.GetUIContext(ctx)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
//
// 	// Parse request
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, "Invalid form data", http.StatusBadRequest)
// 		return
// 	}
//
// 	action := r.FormValue("action")
// 	invoiceIDs := r.Form["ids[]"]
//
// 	if len(invoiceIDs) == 0 {
// 		http.Error(w, "No invoices selected", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Validate action
// 	switch action {
// 	case "download", "pay", "email":
// 		// Valid actions
// 	default:
// 		http.Error(w, "Invalid action", http.StatusBadRequest)
// 		return
// 	}
//
// 	// Process bulk action
// 	results, err := h.processBulkInvoiceAction(ctx, uiCtx.ClientID, action, invoiceIDs)
// 	if err != nil {
// 		h.logger.ErrorContext(ctx, "Failed to process bulk invoice action", logger.Fields{
// 			"error":     err.Error(),
// 			"user_id":   uiCtx.UserID,
// 			"client_id": uiCtx.ClientID,
// 			"action":    action,
// 			"count":     len(invoiceIDs),
// 		})
// 		http.Error(w, "Failed to process bulk action", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Log bulk action
// 	h.logger.InfoContext(ctx, "Bulk invoice action processed", logger.Fields{
// 		"user_id":       uiCtx.UserID,
// 		"client_id":     uiCtx.ClientID,
// 		"action":        action,
// 		"count":         len(invoiceIDs),
// 		"success_count": results.SuccessCount,
// 		"error_count":   results.ErrorCount,
// 	})
//
// 	// Return JSON response
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]any{
// 		"success": true,
// 		"message": fmt.Sprintf("Bulk action completed: %d successful, %d failed", results.SuccessCount, results.ErrorCount),
// 		"results": results,
// 	})
// }
//
// // Private helper methods
//
// // parseInvoiceFilters extracts invoice filtering parameters from request
// func (h *InvoicesHandler) parseInvoiceFilters(r *http.Request) *types.PortalInvoiceFilters {
// 	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
// 	if page < 1 {
// 		page = 1
// 	}
//
// 	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
// 	if pageSize < 1 || pageSize > 100 {
// 		pageSize = 20
// 	}
//
// 	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
//
// 	return &types.PortalInvoiceFilters{
// 		Page:     page,
// 		PageSize: pageSize,
// 		Status:   r.URL.Query().Get("status"),
// 		Year:     year,
// 	}
// }
//
// // getClientInvoices retrieves invoices for the current client with filtering
// func (h *InvoicesHandler) getClientInvoices(ctx context.Context, clientID uuid.UUID, filters *types.PortalInvoiceFilters) ([]types.PortalInvoice, *types.PaginationMeta, error) {
// 	// TODO: Implement actual invoice querying with client filtering
// 	// For now, return placeholder data
//
// 	invoices := []types.PortalInvoice{}
// 	statuses := []string{"sent", "viewed", "paid", "overdue", "partial"}
//
// 	for i := 1; i <= 25; i++ {
// 		status := statuses[i%len(statuses)]
// 		dueDate := time.Now().Add(time.Duration((i-10)*24) * time.Hour) // Some past, some future
//
// 		invoice := types.PortalInvoice{
// 			ID:          uuid.New().String(),
// 			Number:      fmt.Sprintf("INV-2024-%03d", i),
// 			Date:        time.Now().Add(-time.Duration(i*24) * time.Hour),
// 			DueDate:     dueDate,
// 			Amount:      float64(1000 + i*100),
// 			Status:      status,
// 			Description: fmt.Sprintf("Services for %s 2024", time.Now().Month().String()),
// 			Currency:    "USD",
// 			TaxAmount:   float64((1000 + i*100)) * 0.1, // 10% tax
// 		}
//
// 		// Set paid date for paid invoices
// 		if status == "paid" {
// 			paidDate := dueDate.Add(-time.Duration(i*2) * 24 * time.Hour)
// 			invoice.PaidDate = &paidDate
// 		}
//
// 		invoices = append(invoices, invoice)
// 	}
//
// 	// Apply filters
// 	filteredInvoices := h.applyInvoiceFilters(invoices, filters)
//
// 	// Calculate pagination
// 	total := len(filteredInvoices)
// 	start := (filters.Page - 1) * filters.PageSize
// 	end := start + filters.PageSize
//
// 	if start > total {
// 		start = total
// 	}
// 	if end > total {
// 		end = total
// 	}
//
// 	paginatedInvoices := filteredInvoices[start:end]
//
// 	pagination := &types.PaginationMeta{
// 		Page:       filters.Page,
// 		PageSize:   filters.PageSize,
// 		Total:      total,
// 		TotalPages: (total + filters.PageSize - 1) / filters.PageSize,
// 		HasNext:    filters.Page < (total+filters.PageSize-1)/filters.PageSize,
// 		HasPrev:    filters.Page > 1,
// 	}
//
// 	return paginatedInvoices, pagination, nil
// }
//
// // applyInvoiceFilters applies filtering logic to invoices
// func (h *InvoicesHandler) applyInvoiceFilters(invoices []types.PortalInvoice, filters *types.PortalInvoiceFilters) []types.PortalInvoice {
// 	filtered := make([]types.PortalInvoice, 0)
//
// 	for _, invoice := range invoices {
// 		// Status filter
// 		if filters.Status != "" && filters.Status != "all" && invoice.Status != filters.Status {
// 			continue
// 		}
//
// 		// Year filter
// 		if filters.Year != 0 && invoice.Date.Year() != filters.Year {
// 			continue
// 		}
//
// 		filtered = append(filtered, invoice)
// 	}
//
// 	return filtered
// }
//
// // getClientInvoice retrieves a specific invoice for the client
// func (h *InvoicesHandler) getClientInvoice(ctx context.Context, clientID uuid.UUID, invoiceID string) (*types.PortalInvoice, error) {
// 	// TODO: Implement actual invoice retrieval with client ownership verification
// 	// For now, return placeholder data
// 	return &types.PortalInvoice{
// 		ID:          invoiceID,
// 		Number:      "INV-2024-001",
// 		Date:        time.Now().Add(-30 * 24 * time.Hour),
// 		DueDate:     time.Now().Add(7 * 24 * time.Hour),
// 		Amount:      2450.00,
// 		Status:      "sent",
// 		Description: "Professional Services - March 2024",
// 		Currency:    "USD",
// 		TaxAmount:   245.00,
// 		LineItems: []types.InvoiceLineItem{
// 			{
// 				Description: "Consulting Services",
// 				Quantity:    40,
// 				UnitPrice:   50.00,
// 				Amount:      2000.00,
// 			},
// 			{
// 				Description: "Project Management",
// 				Quantity:    10,
// 				UnitPrice:   45.00,
// 				Amount:      450.00,
// 			},
// 		},
// 		BillingAddress: types.Address{
// 			Street:     "123 Industrial Blvd",
// 			City:       "Manufacturing City",
// 			State:      "MC",
// 			PostalCode: "12345",
// 			Country:    "USA",
// 		},
// 	}, nil
// }
//
// // generateInvoicePDF generates or retrieves the PDF for an invoice
// func (h *InvoicesHandler) generateInvoicePDF(ctx context.Context, invoice *types.PortalInvoice) ([]byte, error) {
// 	// TODO: Implement actual PDF generation
// 	// For now, return placeholder PDF content
// 	pdfContent := fmt.Sprintf(`%%PDF-1.4
// 1 0 obj
// <<
// /Type /Catalog
// /Pages 2 0 R
// >>
// endobj
//
// 2 0 obj
// <<
// /Type /Pages
// /Kids [3 0 R]
// /Count 1
// >>
// endobj
//
// 3 0 obj
// <<
// /Type /Page
// /Parent 2 0 R
// /MediaBox [0 0 612 792]
// /Contents 4 0 R
// >>
// endobj
//
// 4 0 obj
// <<
// /Length 44
// >>
// stream
// BT
// /F1 12 Tf
// 72 720 Td
// (Invoice %s - Amount: $%.2f) Tj
// ET
// endstream
// endobj
//
// xref
// 0 5
// 0000000000 65535 f
// 0000000009 00000 n
// 0000000058 00000 n
// 0000000115 00000 n
// 0000000207 00000 n
// trailer
// <<
// /Size 5
// /Root 1 0 R
// >>
// startxref
// 295
// %%%%EOF`, invoice.Number, invoice.Amount)
//
// 	return []byte(pdfContent), nil
// }
//
// // processInvoicePayment processes a payment for an invoice
// func (h *InvoicesHandler) processInvoicePayment(ctx context.Context, invoice *types.PortalInvoice, paymentMethod, amount string) (*types.PaymentResult, error) {
// 	// TODO: Implement actual payment processing
// 	// For now, return success result
// 	return &types.PaymentResult{
// 		PaymentID: uuid.New().String(),
// 		NewStatus: "paid",
// 		Amount:    invoice.Amount,
// 		Method:    paymentMethod,
// 	}, nil
// }
//
// // processBulkInvoiceAction processes bulk actions on invoices
// func (h *InvoicesHandler) processBulkInvoiceAction(ctx context.Context, clientID uuid.UUID, action string, invoiceIDs []string) (*types.BulkActionResults, error) {
// 	// TODO: Implement actual bulk operations
// 	return &types.BulkActionResults{
// 		SuccessCount: len(invoiceIDs),
// 		ErrorCount:   0,
// 		Errors:       []string{},
// 	}, nil
// }
//
// // Temporary rendering methods (until templates are created)
//
// func (h *InvoicesHandler) renderInvoicesListTemp(w http.ResponseWriter, data *types.PortalInvoicesPageData) {
// 	w.Header().Set("Content-Type", "text/html")
// 	fmt.Fprintf(w, `
// 		<!DOCTYPE html>
// 		<html>
// 		<head>
// 			<title>%s</title>
// 			<script src="/static/hooks/datatable.js"></script>
// 			<style>
// 				body { font-family: Arial, sans-serif; margin: 40px; background: #f8fafc; }
// 				.header { background: linear-gradient(135deg, #8b5cf6, #7c3aed); color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
// 				.actions { margin: 20px 0; }
// 				.btn { background: #8b5cf6; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; margin-right: 10px; }
// 				.table { width: 100%%; border-collapse: collapse; background: white; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
// 				.table th, .table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
// 				.table th { background: #f8fafc; }
// 				.status-sent { color: #2563eb; }
// 				.status-paid { color: #059669; }
// 				.status-overdue { color: #dc2626; }
// 				.status-partial { color: #d97706; }
// 				.status-viewed { color: #7c3aed; }
// 				.amount { font-weight: bold; }
// 			</style>
// 		</head>
// 		<body>
// 			<div class="header">
// 				<h1>%s</h1>
// 				<p>View and manage your invoices</p>
// 			</div>
//
// 			<div class="actions">
// 				<button class="btn" id="bulk-download">Download Selected</button>
// 				<button class="btn" id="bulk-pay">Pay Selected</button>
// 			</div>
//
// 			<div x-data="createPortalDataTable({
// 				rows: %s,
// 				searchColumns: ['number', 'description', 'amount', 'status'],
// 				serviceContext: 'portal'
// 			})">
// 				<div class="filters" style="margin-bottom: 20px; background: white; padding: 15px; border-radius: 8px;">
// 					<select @change="applyFilter" style="padding: 8px; margin-right: 10px;">
// 						<option value="all">All Statuses</option>
// 						<option value="sent">Sent</option>
// 						<option value="viewed">Viewed</option>
// 						<option value="paid">Paid</option>
// 						<option value="overdue">Overdue</option>
// 						<option value="partial">Partial</option>
// 					</select>
// 					<select @change="applyFilter" style="padding: 8px;">
// 						<option value="0">All Years</option>
// 						<option value="2024">2024</option>
// 						<option value="2023">2023</option>
// 						<option value="2022">2022</option>
// 					</select>
// 				</div>
//
// 				<table class="table">
// 					<thead>
// 						<tr>
// 							<th><input type="checkbox" @change="toggleSelectAll"></th>
// 							<th @click="sort('number')">Invoice #</th>
// 							<th @click="sort('date')">Date</th>
// 							<th @click="sort('due_date')">Due Date</th>
// 							<th @click="sort('amount')">Amount</th>
// 							<th @click="sort('status')">Status</th>
// 							<th>Actions</th>
// 						</tr>
// 					</thead>
// 					<tbody>
// 						<template x-for="invoice in paginatedRows" :key="invoice.id">
// 							<tr>
// 								<td><input type="checkbox" :value="invoice.id" @change="toggleRowSelection(invoice.id)"></td>
// 								<td x-text="invoice.number"></td>
// 								<td x-text="new Date(invoice.date).toLocaleDateString()"></td>
// 								<td x-text="new Date(invoice.due_date).toLocaleDateString()"></td>
// 								<td class="amount" x-text="'$' + invoice.amount.toFixed(2)"></td>
// 								<td>
// 									<span :class="'status-' + invoice.status" x-text="invoice.status"></span>
// 								</td>
// 								<td>
// 									<a :href="'/portal/invoices/' + invoice.id" style="margin-right: 10px;">View</a>
// 									<a :href="'/portal/invoices/' + invoice.id + '/download'">Download</a>
// 								</td>
// 							</tr>
// 						</template>
// 					</tbody>
// 				</table>
//
// 				<div class="pagination" style="margin-top: 20px; text-align: center;">
// 					<button @click="prevPage" :disabled="!hasPrevPage" class="btn">Previous</button>
// 					<span x-text="'Page ' + currentPage + ' of ' + totalPages" style="margin: 0 20px;"></span>
// 					<button @click="nextPage" :disabled="!hasNextPage" class="btn">Next</button>
// 				</div>
// 			</div>
//
// 			<script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
// 		</body>
// 		</html>
// 	`, data.Title, data.Title, h.invoicesToJSON(data.Invoices))
// }
//
// func (h *InvoicesHandler) renderInvoiceDetailTemp(w http.ResponseWriter, data *types.PortalInvoiceDetailPageData) {
// 	w.Header().Set("Content-Type", "text/html")
// 	fmt.Fprintf(w, `
// 		<!DOCTYPE html>
// 		<html>
// 		<head>
// 			<title>%s</title>
// 			<style>
// 				body { font-family: Arial, sans-serif; margin: 40px; background: #f8fafc; }
// 				.invoice-card { background: white; border-radius: 8px; padding: 30px; margin-bottom: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
// 				.invoice-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 2px solid #8b5cf6; }
// 				.invoice-number { font-size: 24px; font-weight: bold; color: #8b5cf6; }
// 				.invoice-status { padding: 8px 16px; border-radius: 20px; font-size: 14px; font-weight: bold; text-transform: uppercase; }
// 				.status-sent { background: #dbeafe; color: #2563eb; }
// 				.status-paid { background: #d1fae5; color: #059669; }
// 				.status-overdue { background: #fee2e2; color: #dc2626; }
// 				.invoice-details { display: grid; grid-template-columns: 1fr 1fr; gap: 30px; margin-bottom: 30px; }
// 				.detail-section h3 { margin-bottom: 10px; color: #374151; }
// 				.line-items { margin: 30px 0; }
// 				.line-items table { width: 100%%; border-collapse: collapse; }
// 				.line-items th, .line-items td { padding: 12px; text-align: left; border-bottom: 1px solid #e5e7eb; }
// 				.line-items th { background: #f9fafb; font-weight: bold; }
// 				.total-section { text-align: right; margin-top: 20px; }
// 				.total-row { display: flex; justify-content: space-between; padding: 8px 0; }
// 				.total-final { font-size: 18px; font-weight: bold; border-top: 2px solid #374151; padding-top: 8px; }
// 				.actions { margin-top: 30px; text-align: center; }
// 				.btn { background: #8b5cf6; color: white; padding: 12px 24px; border: none; border-radius: 4px; cursor: pointer; margin: 0 10px; text-decoration: none; display: inline-block; }
// 				.btn-secondary { background: #6b7280; }
// 			</style>
// 		</head>
// 		<body>
// 			<div class="invoice-card">
// 				<div class="invoice-header">
// 					<div>
// 						<div class="invoice-number">%s</div>
// 						<div style="color: #6b7280;">%s</div>
// 					</div>
// 					<div class="invoice-status status-%s">%s</div>
// 				</div>
//
// 				<div class="invoice-details">
// 					<div class="detail-section">
// 						<h3>Invoice Information</h3>
// 						<p><strong>Date:</strong> %s</p>
// 						<p><strong>Due Date:</strong> %s</p>
// 						<p><strong>Currency:</strong> %s</p>
// 					</div>
// 					<div class="detail-section">
// 						<h3>Billing Address</h3>
// 						<p>%s<br>%s, %s %s<br>%s</p>
// 					</div>
// 				</div>
//
// 				<div class="line-items">
// 					<h3>Line Items</h3>
// 					<table>
// 						<thead>
// 							<tr>
// 								<th>Description</th>
// 								<th>Quantity</th>
// 								<th>Unit Price</th>
// 								<th>Amount</th>
// 							</tr>
// 						</thead>
// 						<tbody>
// 							%s
// 						</tbody>
// 					</table>
// 				</div>
//
// 				<div class="total-section">
// 					<div class="total-row">
// 						<span>Subtotal:</span>
// 						<span>$%.2f</span>
// 					</div>
// 					<div class="total-row">
// 						<span>Tax:</span>
// 						<span>$%.2f</span>
// 					</div>
// 					<div class="total-row total-final">
// 						<span>Total:</span>
// 						<span>$%.2f</span>
// 					</div>
// 				</div>
//
// 				<div class="actions">
// 					<a href="/portal/invoices/%s/download" class="btn">Download PDF</a>
// 					%s
// 					<a href="/portal/invoices" class="btn btn-secondary">Back to Invoices</a>
// 				</div>
// 			</div>
// 		</body>
// 		</html>
// 	`,
// 		data.Title,
// 		data.Invoice.Number,
// 		data.Invoice.Description,
// 		data.Invoice.Status,
// 		data.Invoice.Status,
// 		data.Invoice.Date.Format("January 2, 2006"),
// 		data.Invoice.DueDate.Format("January 2, 2006"),
// 		data.Invoice.Currency,
// 		data.Invoice.BillingAddress.Street,
// 		data.Invoice.BillingAddress.City,
// 		data.Invoice.BillingAddress.State,
// 		data.Invoice.BillingAddress.PostalCode,
// 		data.Invoice.BillingAddress.Country,
// 		h.renderLineItems(data.Invoice.LineItems),
// 		data.Invoice.Amount-data.Invoice.TaxAmount,
// 		data.Invoice.TaxAmount,
// 		data.Invoice.Amount,
// 		data.Invoice.ID,
// 		h.renderPayButton(data.Invoice),
// 	)
// }
//
// // Helper methods for temporary rendering
//
// func (h *InvoicesHandler) invoicesToJSON(invoices []types.PortalInvoice) string {
// 	data, _ := json.Marshal(invoices)
// 	return string(data)
// }
//
// func (h *InvoicesHandler) renderLineItems(items []types.InvoiceLineItem) string {
// 	if len(items) == 0 {
// 		return "<tr><td colspan='4'>No line items</td></tr>"
// 	}
//
// 	rows := ""
// 	for _, item := range items {
// 		rows += fmt.Sprintf(`
// 			<tr>
// 				<td>%s</td>
// 				<td>%.0f</td>
// 				<td>$%.2f</td>
// 				<td>$%.2f</td>
// 			</tr>
// 		`, item.Description, item.Quantity, item.UnitPrice, item.Amount)
// 	}
// 	return rows
// }
//
// func (h *InvoicesHandler) renderPayButton(invoice *types.PortalInvoice) string {
// 	if invoice.Status == "paid" {
// 		return ""
// 	}
// 	return fmt.Sprintf(`<button class="btn" onclick="payInvoice('%s')">Pay Now</button>`, invoice.ID)
// }
//
// // SetupRoutes sets up portal invoice routes
// func (h *InvoicesHandler) SetupRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
// 	// Apply authentication middleware to all invoice routes
// 	mux.Handle("/portal/invoices", authMiddleware(http.HandlerFunc(h.ListInvoices)))
// 	mux.Handle("/portal/api/invoices", authMiddleware(http.HandlerFunc(h.GetInvoicesData)))
// 	mux.Handle("/portal/invoices/bulk", authMiddleware(http.HandlerFunc(h.BulkInvoiceActions)))
// 	mux.Handle("/portal/invoices/pay", authMiddleware(http.HandlerFunc(h.PayInvoice)))
// 	// Note: Invoice detail and download routes need path parameter handling in actual router
// }
//
