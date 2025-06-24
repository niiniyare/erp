package inventory


import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// Transaction represents a database transaction record
type Transaction struct {
	ID                int     `json:"id"`
	OrderNumber       string  `json:"ordnumber"`
	TransDate         string  `json:"transdate"`
	ReqDate           string  `json:"reqdate"`
	Amount            float64 `json:"amount"`
	Name              string  `json:"name"`
	CustomerNumber    string  `json:"customernumber"`
	VendorNumber      string  `json:"vendornumber"`
	NetAmount         float64 `json:"netamount"`
	CustomerID        int     `json:"customer_id"`
	VendorID          int     `json:"vendor_id"`
	ExchangeRate      float64 `json:"exchangerate"`
	Closed            bool    `json:"closed"`
	QuoNumber         string  `json:"quonumber"`
	ShippingPoint     string  `json:"shippingpoint"`
	ShipVia           string  `json:"shipvia"`
	Waybill           string  `json:"waybill"`
	Employee          string  `json:"employee"`
	Manager           string  `json:"manager"`
	Currency          string  `json:"curr"`
	PONumber          string  `json:"ponumber"`
	Notes             string  `json:"notes"`
	IntNotes          string  `json:"intnotes"`
	Warehouse         string  `json:"warehouse"`
	Description       string  `json:"description"`
	Memo              string  `json:"memo"`
	SellPrice         float64 `json:"sellprice"`
	Qty               float64 `json:"qty"`
	OrderItemsID      int     `json:"orderitemsid"`
}

// FormData represents the input parameters for the query
type FormData struct {
	Type            string
	VC              string // 'customer' or 'vendor'
	Detail          bool
	LMemo           string
	LEmployee       bool
	LManager        bool
	Year            int
	Month           int
	Interval        string
	TransDateFrom   string
	TransDateTo     string
	OrderNumber     string
	QuoNumber       string
	CustomerID      int
	VendorID        int
	Customer        string
	Vendor          string
	CustomerNumber  string
	VendorNumber    string
	PONumber        string
	Open            bool
	Closed          bool
	ShipVia         string
	Waybill         string
	Notes           string
	Description     string
	Memo            string
	Department      string
	Warehouse       string
	Employee        string
}

// MyConfig represents database configuration
type MyConfig struct {
	DSN string
}

// TransactionService handles transaction operations
type TransactionService struct {
	db *sql.DB
}

// NewTransactionService creates a new transaction service
func NewTransactionService(db *sql.DB) *TransactionService {
	return &TransactionService{db: db}
}

// Transactions retrieves transactions based on form criteria
func (ts *TransactionService) Transactions(config *MyConfig, form *FormData) ([]*Transaction, error) {
	// Remove locks (implementation would depend on your locking mechanism)
	if err := ts.removeLocks("oe"); err != nil {
		return nil, fmt.Errorf("failed to remove locks: %w", err)
	}

	// Set default values
	form = ts.setDefaults(form)

	var (
		orderNumber            = "ordnumber"
		quotation              = "0"
		whereClause            strings.Builder
		orderItemsDescription  string
		orderItemsJoin         string
	)

	// Handle ship/receive orders or memo requests
	if strings.Contains(form.Type, "ship_order") || strings.Contains(form.Type, "receive_order") || form.LMemo == "Y" {
		orderItemsDescription = ", oi.description AS memo"
		orderItemsJoin = "JOIN orderitems oi ON (oi.trans_id = o.id)"
	}

	// Handle detail view
	if form.Detail {
		orderItemsDescription += ", oi.description AS memo, oi.sellprice, oi.qty, oi.id AS orderitemsid"
		if orderItemsJoin == "" {
			orderItemsJoin = "JOIN orderitems oi ON (oi.trans_id = o.id)"
		}
	}

	// Ensure VC is either 'customer' or 'vendor' for SQL injection protection
	if form.VC != "customer" {
		form.VC = "vendor"
	}

	rate := "sell"
	if form.VC == "customer" {
		rate = "buy"
	}

	// Handle date range
	if form.Year > 0 && form.Month > 0 {
		form.TransDateFrom, form.TransDateTo = ts.calculateDateRange(form.Year, form.Month, form.Interval)
	}

	// Handle quotations
	if strings.HasSuffix(form.Type, "_quotation") {
		quotation = "1"
		orderNumber = "quonumber"
	}

	// Build WHERE clause for department, warehouse, employee
	ts.buildDepartmentWhere(form, &whereClause)

	// Build main query
	query := fmt.Sprintf(`
		SELECT o.id, o.ordnumber, o.transdate, o.reqdate,
		       o.amount, ct.name, ct.%snumber, o.netamount,
		       o.%s_id,
		       ex.%s AS exchangerate,
		       o.closed, o.quonumber, o.shippingpoint, o.shipvia, o.waybill,
		       e.name AS employee, m.name AS manager, o.curr, o.ponumber,
		       o.notes, o.intnotes, w.description AS warehouse, o.description
		       %s
		FROM oe o
		JOIN %s ct ON (o.%s_id = ct.id)
		%s
		LEFT JOIN employee e ON (o.employee_id = e.id)
		LEFT JOIN employee m ON (e.managerid = m.id)
		LEFT JOIN warehouse w ON (o.warehouse_id = w.id)
		LEFT JOIN exchangerate ex ON (ex.curr = o.curr AND ex.transdate = o.transdate)
		WHERE o.quotation = '%s'
		%s`,
		form.VC, form.VC, rate, orderItemsDescription, form.VC, form.VC,
		orderItemsJoin, quotation, whereClause.String())

	// Add additional WHERE conditions
	query = ts.buildAdditionalWhere(query, form, orderNumber)

	// Add ORDER BY clause
	query += ts.buildOrderBy(form, orderNumber)

	// Execute query
	rows, err := ts.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	return ts.processResults(rows, form)
}

// setDefaults sets default values for the form
func (ts *TransactionService) setDefaults(form *FormData) *FormData {
	// Implementation would fetch defaults from database
	// For now, just return the form as-is
	return form
}

// removeLocks removes database locks
func (ts *TransactionService) removeLocks(table string) error {
	// Implementation would depend on your locking mechanism
	// This is a placeholder
	return nil
}

// calculateDateRange calculates date range based on year, month, and interval
func (ts *TransactionService) calculateDateRange(year, month int, interval string) (string, string) {
	// Implementation would calculate the date range
	// This is a simplified version
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := fmt.Sprintf("%04d-%02d-31", year, month)
	return start, end
}

// buildDepartmentWhere builds WHERE clause for department, warehouse, employee
func (ts *TransactionService) buildDepartmentWhere(form *FormData, whereClause *strings.Builder) {
	fields := map[string]string{
		"department": form.Department,
		"warehouse":  form.Warehouse,
		"employee":   form.Employee,
	}

	for field, value := range fields {
		if value != "" {
			// Clean the value and extract ID
			cleaned := ts.dbClean(value)
			if parts := strings.Split(cleaned, "--"); len(parts) > 1 {
				if id, err := strconv.Atoi(parts[1]); err == nil {
					whereClause.WriteString(fmt.Sprintf(" AND o.%s_id = %d", field, id))
				}
			}
		}
	}
}

// buildAdditionalWhere adds additional WHERE conditions
func (ts *TransactionService) buildAdditionalWhere(query string, form *FormData, orderNumber string) string {
	var conditions []string

	// Handle ship/receive orders
	if strings.Contains(form.Type, "ship_order") || strings.Contains(form.Type, "receive_order") {
		conditions = append(conditions,
			"o.quotation = '0'",
			"oi.qty <> oi.ship",
			"o.id NOT IN (SELECT id FROM semaphore)")
	}

	// Customer/Vendor ID filter
	if form.VC == "customer" && form.CustomerID > 0 {
		conditions = append(conditions, fmt.Sprintf("o.customer_id = %d", form.CustomerID))
	} else if form.VC == "vendor" && form.VendorID > 0 {
		conditions = append(conditions, fmt.Sprintf("o.vendor_id = %d", form.VendorID))
	} else {
		// Handle name search
		if form.VC == "customer" && form.Customer != "" {
			name := ts.like(strings.ToLower(form.Customer))
			conditions = append(conditions, fmt.Sprintf("lower(ct.name) LIKE '%s'", name))
		} else if form.VC == "vendor" && form.Vendor != "" {
			name := ts.like(strings.ToLower(form.Vendor))
			conditions = append(conditions, fmt.Sprintf("lower(ct.name) LIKE '%s'", name))
		}

		// Handle number search
		if form.VC == "customer" && form.CustomerNumber != "" {
			number := ts.like(strings.ToLower(form.CustomerNumber))
			conditions = append(conditions, fmt.Sprintf("lower(ct.customernumber) LIKE '%s'", number))
		} else if form.VC == "vendor" && form.VendorNumber != "" {
			number := ts.like(strings.ToLower(form.VendorNumber))
			conditions = append(conditions, fmt.Sprintf("lower(ct.vendornumber) LIKE '%s'", number))
		}
	}

	// Order number filter
	if (orderNumber == "ordnumber" && form.OrderNumber != "") ||
		(orderNumber == "quonumber" && form.QuoNumber != "") {
		var searchValue string
		if orderNumber == "ordnumber" {
			searchValue = form.OrderNumber
		} else {
			searchValue = form.QuoNumber
		}
		number := ts.like(strings.ToLower(searchValue))
		conditions = append(conditions, fmt.Sprintf("lower(%s) LIKE '%s'", orderNumber, number))
	}

	// PO Number filter
	if form.PONumber != "" {
		ponumber := ts.like(strings.ToLower(form.PONumber))
		conditions = append(conditions, fmt.Sprintf("lower(ponumber) LIKE '%s'", ponumber))
	}

	// Open/Closed filter
	if !form.Open && !form.Closed {
		conditions = append(conditions, "o.id = 0")
	} else if form.Open && !form.Closed {
		conditions = append(conditions, "o.closed = '0'")
	} else if !form.Open && form.Closed {
		conditions = append(conditions, "o.closed = '1'")
	}

	// Other text filters
	textFilters := map[string]string{
		"o.shipvia":     form.ShipVia,
		"o.waybill":     form.Waybill,
		"o.notes":       form.Notes,
		"o.description": form.Description,
	}

	for field, value := range textFilters {
		if value != "" {
			likeValue := ts.like(strings.ToLower(value))
			conditions = append(conditions, fmt.Sprintf("lower(%s) LIKE '%s'", field, likeValue))
		}
	}

	// Memo filter (subquery)
	if form.Memo != "" {
		memo := ts.like(strings.ToLower(form.Memo))
		conditions = append(conditions,
			fmt.Sprintf(`o.id IN (SELECT DISTINCT trans_id 
						FROM orderitems 
						WHERE lower(description) LIKE '%s')`, memo))
	}

	// Date range filters
	if form.TransDateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("o.transdate >= '%s'", ts.dbClean(form.TransDateFrom)))
	}
	if form.TransDateTo != "" {
		conditions = append(conditions, fmt.Sprintf("o.transdate <= '%s'", ts.dbClean(form.TransDateTo)))
	}

	// Add conditions to query
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	return query
}

// buildOrderBy builds the ORDER BY clause
func (ts *TransactionService) buildOrderBy(form *FormData, orderNumber string) string {
	ordinal := map[string]int{
		"id":          1,
		"ordnumber":   2,
		"transdate":   3,
		"reqdate":     4,
		"name":        6,
		"quonumber":   12,
		"shipvia":     14,
		"waybill":     15,
		"employee":    16,
		"manager":     17,
		"curr":        18,
		"ponumber":    19,
		"warehouse":   21,
		"description": 22,
	}

	sortFields := []string{"transdate", orderNumber, "name"}
	if form.LEmployee {
		sortFields = append(sortFields, "employee")
	}
	if form.LManager && !strings.Contains(form.Type, "ship_order") && !strings.Contains(form.Type, "receive_order") {
		sortFields = append(sortFields, "manager")
	}

	// Convert to ORDER BY clause (simplified)
	return " ORDER BY " + strings.Join(sortFields, ", ")
}

// processResults processes the query results
func (ts *TransactionService) processResults(rows *sql.Rows, form *FormData) ([]*Transaction, error) {
	var transactions []*Transaction
	objID := make(map[string]map[string]map[int]int)
	objID["vc"] = make(map[string]map[int]int)
	objID["id"] = make(map[string]map[int]int)

	i := -1
	seenIDs := make(map[int]bool)

	for rows.Next() {
		var t Transaction
		var exchangeRate sql.NullFloat64

		// Scan row (this would need to match your exact query columns)
		err := rows.Scan(
			&t.ID, &t.OrderNumber, &t.TransDate, &t.ReqDate,
			&t.Amount, &t.Name, &t.CustomerNumber, &t.NetAmount,
			&t.CustomerID, &exchangeRate, &t.Closed, &t.QuoNumber,
			&t.ShippingPoint, &t.ShipVia, &t.Waybill, &t.Employee,
			&t.Manager, &t.Currency, &t.PONumber, &t.Notes,
			&t.IntNotes, &t.Warehouse, &t.Description,
			// Additional fields for detail mode would be scanned here
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Set default exchange rate
		if exchangeRate.Valid {
			t.ExchangeRate = exchangeRate.Float64
		} else {
			t.ExchangeRate = 1.0
		}

		if form.Detail {
			i++
			ml := 1.0
			if t.NetAmount != 0 {
				ml = t.Amount / t.NetAmount
			}
			t.NetAmount = t.SellPrice * t.Qty
			t.Amount = t.NetAmount * ml
			transactions = append(transactions, &t)
		} else {
			if !seenIDs[t.ID] {
				i++
				transactions = append(transactions, &t)
				seenIDs[t.ID] = true
			} else {
				// Append memo to existing transaction
				if t.Memo != "" {
					transactions[i].Memo += "\n" + t.Memo
				}
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Handle consolidation
	if strings.HasPrefix(form.Type, "consolidate_") {
		transactions = ts.consolidateTransactions(transactions, form)
	}

	return transactions, nil
}

// consolidateTransactions consolidates transactions based on criteria
func (ts *TransactionService) consolidateTransactions(transactions []*Transaction, form *FormData) []*Transaction {
	// Count occurrences by currency and customer/vendor
	counts := make(map[string]map[int]int)
	
	for _, t := range transactions {
		if counts[t.Currency] == nil {
			counts[t.Currency] = make(map[int]int)
		}
		if form.VC == "customer" {
			counts[t.Currency][t.CustomerID]++
		} else {
			counts[t.Currency][t.VendorID]++
		}
	}

	// Filter to only include transactions with count > 1
	var consolidated []*Transaction
	for _, t := range transactions {
		var id int
		if form.VC == "customer" {
			id = t.CustomerID
		} else {
			id = t.VendorID
		}
		
		if counts[t.Currency][id] > 1 {
			consolidated = append(consolidated, t)
		}
	}

	return consolidated
}

// Helper functions
func (ts *TransactionService) like(s string) string {
	return "%" + strings.ReplaceAll(s, "'", "''") + "%"
}

func (ts *TransactionService) dbClean(s string) string {
	// Implementation for cleaning database input
	return strings.ReplaceAll(s, "'", "''")
}

// Example usage
func Example() {
	// This would be your actual database connection
	db, err := sql.Open("postgres", "your-connection-string")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	service := NewTransactionService(db)
	
	form := &FormData{
		Type:   "sales_order",
		VC:     "customer",
		Detail: false,
		Open:   true,
		Closed: false,
	}

	config := &MyConfig{
		DSN: "your-connection-string",
	}

	transactions, err := service.Transactions(config, form)
	if err != nil {
		log.Printf("Error fetching transactions: %v", err)
		return
	}

	fmt.Printf("Found %d transactions\n", len(transactions))
}


)



//========================================================
//      SaveOrder
//========================================================

// OrderType represents the type of order/quotation
type OrderType string

const (
	SalesOrder      OrderType = "sales_order"
	PurchaseOrder   OrderType = "purchase_order"
	SalesQuotation  OrderType = "sales_quotation"
	PurchaseQuote   OrderType = "purchase_quote"
)

// OrderService handles order/quotation operations
type OrderService struct {
	db *sql.DB
}

// NewOrderService creates a new order service instance
func NewOrderService(db *sql.DB) *OrderService {
	return &OrderService{db: db}
}

// OrderForm represents the main order data structure
type OrderForm struct {
	ID               int64             `json:"id"`
	Type             OrderType         `json:"type"`
	OrderNumber      string            `json:"order_number"`
	QuoteNumber      string            `json:"quote_number"`
	Description      string            `json:"description"`
	TransactionDate  time.Time         `json:"transaction_date"`
	RequiredDate     *time.Time        `json:"required_date"`
	VendorID         *int64            `json:"vendor_id"`
	CustomerID       *int64            `json:"customer_id"`
	EmployeeID       int64             `json:"employee_id"`
	DepartmentID     *int64            `json:"department_id"`
	WarehouseID      *int64            `json:"warehouse_id"`
	Currency         string            `json:"currency"`
	DefaultCurrency  string            `json:"default_currency"`
	ExchangeRate     float64           `json:"exchange_rate"`
	TaxIncluded      bool              `json:"tax_included"`
	ShippingPoint    string            `json:"shipping_point"`
	ShipVia          string            `json:"ship_via"`
	Waybill          string            `json:"waybill"`
	Notes            string            `json:"notes"`
	InternalNotes    string            `json:"internal_notes"`
	LanguageCode     string            `json:"language_code"`
	PONumber         string            `json:"po_number"`
	Terms            int               `json:"terms"`
	Closed           bool              `json:"closed"`
	AddShipping      bool              `json:"add_shipping"`
	Precision        int               `json:"precision"`
	Items            []OrderItem       `json:"items"`
	TaxRates         map[string]float64 `json:"tax_rates"`
}

// OrderItem represents a line item in an order
type OrderItem struct {
	ID              int64    `json:"id"`
	ProductID       int64    `json:"product_id"`
	Description     string   `json:"description"`
	Quantity        float64  `json:"quantity"`
	SellPrice       float64  `json:"sell_price"`
	Discount        float64  `json:"discount"`
	Unit            string   `json:"unit"`
	RequiredDate    *time.Time `json:"required_date"`
	ProjectID       *int64   `json:"project_id"`
	ShipQuantity    float64  `json:"ship_quantity"`
	SerialNumber    string   `json:"serial_number"`
	OrderNumber     string   `json:"order_number"`
	PONumber        string   `json:"po_number"`
	ItemNotes       string   `json:"item_notes"`
	TaxAccounts     string   `json:"tax_accounts"`
	Package         string   `json:"package"`
	NetWeight       float64  `json:"net_weight"`
	GrossWeight     float64  `json:"gross_weight"`
	Volume          float64  `json:"volume"`
	LineItemDetail  bool     `json:"line_item_detail"`
}

// ProductInfo represents product details from database
type ProductInfo struct {
	Assembly  bool   `json:"assembly"`
	ProjectID *int64 `json:"project_id"`
}

// TaxCalculation represents tax calculation results
type TaxCalculation struct {
	TaxAccounts map[string]float64 `json:"tax_accounts"`
	TaxBases    map[string]float64 `json:"tax_bases"`
	TotalTax    float64            `json:"total_tax"`
}

// OrderTotals represents calculated order totals
type OrderTotals struct {
	NetAmount   float64 `json:"net_amount"`
	TaxAmount   float64 `json:"tax_amount"`
	TotalAmount float64 `json:"total_amount"`
}

// Save processes and saves an order/quotation to the database
func (os *OrderService) Save(form *OrderForm) error {
	// Start transaction
	tx, err := os.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get system defaults
	if err := os.getSystemDefaults(tx, form); err != nil {
		return fmt.Errorf("failed to get system defaults: %w", err)
	}

	// Handle employee ID
	if err := os.handleEmployeeID(tx, form); err != nil {
		return fmt.Errorf("failed to handle employee ID: %w", err)
	}

	// Handle existing vs new order
	if err := os.handleOrderRecord(tx, form); err != nil {
		return fmt.Errorf("failed to handle order record: %w", err)
	}

	// Process line items and calculate totals
	totals, err := os.processLineItems(tx, form)
	if err != nil {
		return fmt.Errorf("failed to process line items: %w", err)
	}

	// Handle currency and exchange rate
	if err := os.handleCurrency(tx, form); err != nil {
		return fmt.Errorf("failed to handle currency: %w", err)
	}

	// Generate document number if needed
	if err := os.generateDocumentNumber(tx, form); err != nil {
		return fmt.Errorf("failed to generate document number: %w", err)
	}

	// Save main order record
	if err := os.saveMainOrder(tx, form, totals); err != nil {
		return fmt.Errorf("failed to save main order: %w", err)
	}

	// Handle additional operations
	if err := os.handleAdditionalOperations(tx, form); err != nil {
		return fmt.Errorf("failed to handle additional operations: %w", err)
	}

	// Create audit trail
	if err := os.createAuditTrail(tx, form); err != nil {
		return fmt.Errorf("failed to create audit trail: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// getSystemDefaults retrieves system configuration defaults
func (os *OrderService) getSystemDefaults(tx *sql.Tx, form *OrderForm) error {
	query := `SELECT precision FROM defaults WHERE 1=1 LIMIT 1`
	var precision sql.NullInt64
	err := tx.QueryRow(query).Scan(&precision)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if precision.Valid {
		form.Precision = int(precision.Int64)
	} else {
		form.Precision = 2 // default precision
	}
	return nil
}

// handleEmployeeID processes employee information
func (os *OrderService) handleEmployeeID(tx *sql.Tx, form *OrderForm) error {
	if form.EmployeeID == 0 {
		// Get current employee (simplified)
		query := `SELECT id FROM employee ORDER BY id LIMIT 1`
		err := tx.QueryRow(query).Scan(&form.EmployeeID)
		if err != nil {
			return fmt.Errorf("failed to get employee: %w", err)
		}
	}
	return nil
}

// handleOrderRecord manages existing vs new order records
func (os *OrderService) handleOrderRecord(tx *sql.Tx, form *OrderForm) error {
	if form.ID > 0 {
		// Check if record exists
		var existingID int64
		var aaID sql.NullInt64
		query := `SELECT id, aa_id FROM oe WHERE id = $1`
		err := tx.QueryRow(query, form.ID).Scan(&existingID, &aaID)
		
		if err == nil {
			// Record exists, adjust inventory if needed
			if os.isOrder(form.Type) && !aaID.Valid {
				if err := os.adjustInventoryOnHand(tx, form, os.getMultiplier(form.Type)); err != nil {
					return err
				}
			}
			
			// Clean up old data
			if err := os.cleanupOldData(tx, form.ID); err != nil {
				return err
			}
		} else if err == sql.ErrNoRows {
			// Create placeholder record
			query := `INSERT INTO oe (id) VALUES ($1)`
			if _, err := tx.Exec(query, form.ID); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Create new record
		uid := os.generateUID()
		query := `INSERT INTO oe (ordnumber, employee_id) VALUES ($1, $2) RETURNING id`
		err := tx.QueryRow(query, uid, form.EmployeeID).Scan(&form.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// processLineItems processes all line items and calculates totals
func (os *OrderService) processLineItems(tx *sql.Tx, form *OrderForm) (*OrderTotals, error) {
	totals := &OrderTotals{}
	taxCalc := &TaxCalculation{
		TaxAccounts: make(map[string]float64),
		TaxBases:    make(map[string]float64),
	}

	uid := os.generateUID()
	multiplier := os.getMultiplier(form.Type)

	for i := range form.Items {
		item := &form.Items[i]
		
		if item.Quantity == 0 {
			continue
		}

		// Get product information
		productInfo, err := os.getProductInfo(tx, item.ProductID)
		if err != nil {
			return nil, err
		}

		// Calculate line totals
		discountAmount := item.SellPrice * (item.Discount / 100)
		finalPrice := item.SellPrice - discountAmount
		lineTotal := os.roundAmount(finalPrice * item.Quantity, form.Precision)

		// Calculate taxes for this line
		if err := os.calculateLineTaxes(item, lineTotal, taxCalc, form); err != nil {
			return nil, err
		}

		totals.NetAmount += finalPrice * item.Quantity

		// Save line item
		if err := os.saveOrderItem(tx, form, item, uid, i, productInfo); err != nil {
			return nil, err
		}

		// Handle shipping/inventory
		if form.AddShipping {
			if err := os.addInventoryTransaction(tx, form, item, multiplier); err != nil {
				return nil, err
			}
		}

		// Handle cargo/package info
		if os.hasPackageInfo(item) {
			if err := os.saveCargoInfo(tx, form.ID, item); err != nil {
				return nil, err
			}
		}

		// Update product weight if applicable
		if os.isOrder(form.Type) && item.NetWeight > 0 {
			if err := os.updateProductWeight(tx, item); err != nil {
				return nil, err
			}
		}
	}

	// Calculate final totals
	totals.TaxAmount = taxCalc.TotalTax
	if form.TaxIncluded {
		totals.NetAmount -= totals.TaxAmount
	}
	totals.TotalAmount = os.roundAmount(totals.NetAmount + totals.TaxAmount, form.Precision)
	totals.NetAmount = os.roundAmount(totals.NetAmount, form.Precision)

	return totals, nil
}

// saveOrderItem saves a single order item to the database
func (os *OrderService) saveOrderItem(tx *sql.Tx, form *OrderForm, item *OrderItem, uid string, index int, productInfo *ProductInfo) error {
	// Insert initial record with UID
	query := `INSERT INTO orderitems (description, trans_id, parts_id) VALUES ($1, $2, $3) RETURNING id`
	err := tx.QueryRow(query, uid, form.ID, item.ProductID).Scan(&item.ID)
	if err != nil {
		return err
	}

	// Project ID handling
	projectID := item.ProjectID
	if projectID == nil && productInfo.ProjectID != nil {
		projectID = productInfo.ProjectID
	}

	// Update with actual data
	updateQuery := `
		UPDATE orderitems SET
			description = $1,
			qty = $2,
			sellprice = $3,
			discount = $4,
			unit = $5,
			reqdate = $6,
			project_id = $7,
			ship = $8,
			serialnumber = $9,
			ordernumber = $10,
			ponumber = $11,
			itemnotes = $12,
			lineitemdetail = $13
		WHERE id = $14 AND trans_id = $15`

	_, err = tx.Exec(updateQuery,
		item.Description,
		item.Quantity,
		item.SellPrice,
		item.Discount/100, // Convert back to decimal
		item.Unit,
		item.RequiredDate,
		projectID,
		item.ShipQuantity,
		item.SerialNumber,
		item.OrderNumber,
		item.PONumber,
		item.ItemNotes,
		item.LineItemDetail,
		item.ID,
		form.ID,
	)

	return err
}

// calculateLineTaxes calculates taxes for a line item
func (os *OrderService) calculateLineTaxes(item *OrderItem, lineTotal float64, taxCalc *TaxCalculation, form *OrderForm) error {
	if item.TaxAccounts == "" {
		return nil
	}

	taxAccounts := strings.Fields(item.TaxAccounts)
	
	for _, account := range taxAccounts {
		rate, exists := form.TaxRates[account]
		if !exists {
			continue
		}

		if form.TaxIncluded {
			// Tax is included in the price
			taxAmount := lineTotal * rate / (1 + rate)
			taxBase := lineTotal - taxAmount
			
			taxCalc.TaxAccounts[account] += taxAmount
			taxCalc.TaxBases[account] += taxBase
		} else {
			// Tax is added to the price
			taxAmount := lineTotal * rate
			
			taxCalc.TaxAccounts[account] += taxAmount
			taxCalc.TaxBases[account] += lineTotal
		}
	}

	// Calculate total tax for this line
	var lineTax float64
	for _, account := range taxAccounts {
		if rate, exists := form.TaxRates[account]; exists {
			if form.TaxIncluded {
				lineTax += lineTotal * rate / (1 + rate)
			} else {
				lineTax += lineTotal * rate
			}
		}
	}
	
	taxCalc.TotalTax += lineTax
	return nil
}

// addInventoryTransaction adds inventory movement record
func (os *OrderService) addInventoryTransaction(tx *sql.Tx, form *OrderForm, item *OrderItem, multiplier int) error {
	query := `
		INSERT INTO inventory (parts_id, warehouse_id, department_id, qty, trans_id, orderitems_id, shippingdate, employee_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := tx.Exec(query,
		item.ProductID,
		form.WarehouseID,
		form.DepartmentID,
		item.ShipQuantity * float64(multiplier) * -1,
		form.ID,
		item.ID,
		form.TransactionDate,
		form.EmployeeID,
	)
	
	return err
}

// saveCargoInfo saves package/cargo information
func (os *OrderService) saveCargoInfo(tx *sql.Tx, orderID int64, item *OrderItem) error {
	query := `
		INSERT INTO cargo (id, trans_id, package, netweight, grossweight, volume)
		VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err := tx.Exec(query,
		item.ID,
		orderID,
		item.Package,
		item.NetWeight,
		item.GrossWeight,
		item.Volume,
	)
	
	return err
}

// handleCurrency processes currency and exchange rate
func (os *OrderService) handleCurrency(tx *sql.Tx, form *OrderForm) error {
	if form.Currency == form.DefaultCurrency {
		form.ExchangeRate = 1.0
		return nil
	}

	// Check for existing exchange rate
	query := `SELECT rate FROM exchangerate WHERE curr = $1 AND transdate = $2`
	var existingRate sql.NullFloat64
	err := tx.QueryRow(query, form.Currency, form.TransactionDate).Scan(&existingRate)
	
	if err == nil && existingRate.Valid {
		form.ExchangeRate = existingRate.Float64
	} else if form.ExchangeRate == 0 {
		form.ExchangeRate = 1.0 // Default fallback
	}

	return nil
}

// generateDocumentNumber generates order/quote numbers
func (os *OrderService) generateDocumentNumber(tx *sql.Tx, form *OrderForm) error {
	if form.OrderNumber != "" {
		return nil
	}

	var numberField string
	switch form.Type {
	case SalesOrder:
		numberField = "sonumber"
	case PurchaseOrder:
		numberField = "ponumber"
	case SalesQuotation:
		numberField = "sqnumber"
	case PurchaseQuote:
		numberField = "rfqnumber"
	}

	// Get next number (simplified)
	query := `SELECT COALESCE(MAX(CAST(SUBSTRING(ordnumber FROM '[0-9]+') AS INTEGER)), 0) + 1 FROM oe WHERE ordnumber ~ '^[0-9]+$'`
	var nextNum int
	err := tx.QueryRow(query).Scan(&nextNum)
	if err != nil {
		nextNum = 1
	}

	form.OrderNumber = fmt.Sprintf("%06d", nextNum)
	return nil
}

// saveMainOrder saves the main order record
func (os *OrderService) saveMainOrder(tx *sql.Tx, form *OrderForm, totals *OrderTotals) error {
	quotation := !os.isOrder(form.Type)
	
	query := `
		UPDATE oe SET
			ordnumber = $1,
			quonumber = $2,
			description = $3,
			transdate = $4,
			vendor_id = $5,
			customer_id = $6,
			amount = $7,
			netamount = $8,
			reqdate = $9,
			taxincluded = $10,
			shippingpoint = $11,
			shipvia = $12,
			waybill = $13,
			notes = $14,
			intnotes = $15,
			curr = $16,
			closed = $17,
			quotation = $18,
			department_id = $19,
			employee_id = $20,
			language_code = $21,
			ponumber = $22,
			terms = $23,
			warehouse_id = $24,
			exchangerate = $25
		WHERE id = $26`

	_, err := tx.Exec(query,
		form.OrderNumber,
		form.QuoteNumber,
		form.Description,
		form.TransactionDate,
		form.VendorID,
		form.CustomerID,
		totals.TotalAmount,
		totals.NetAmount,
		form.RequiredDate,
		form.TaxIncluded,
		form.ShippingPoint,
		form.ShipVia,
		form.Waybill,
		form.Notes,
		form.InternalNotes,
		form.Currency,
		form.Closed,
		quotation,
		form.DepartmentID,
		form.EmployeeID,
		form.LanguageCode,
		form.PONumber,
		form.Terms,
		form.WarehouseID,
		form.ExchangeRate,
		form.ID,
	)

	return err
}

// Helper functions

func (os *OrderService) getProductInfo(tx *sql.Tx, productID int64) (*ProductInfo, error) {
	query := `SELECT assembly, project_id FROM parts WHERE id = $1`
	info := &ProductInfo{}
	var projectID sql.NullInt64
	
	err := tx.QueryRow(query, productID).Scan(&info.Assembly, &projectID)
	if err != nil {
		return nil, err
	}
	
	if projectID.Valid {
		info.ProjectID = &projectID.Int64
	}
	
	return info, nil
}

func (os *OrderService) cleanupOldData(tx *sql.Tx, orderID int64) error {
	tables := []string{"dpt_trans", "orderitems", "shipto", "cargo"}
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE trans_id = $1", table)
		if _, err := tx.Exec(query, orderID); err != nil {
			return err
		}
	}
	return nil
}

func (os *OrderService) adjustInventoryOnHand(tx *sql.Tx, form *OrderForm, multiplier int) error {
	// Simplified inventory adjustment
	// In a real implementation, this would handle complex inventory logic
	log.Printf("Adjusting inventory for order %d with multiplier %d", form.ID, multiplier)
	return nil
}

func (os *OrderService) updateProductWeight(tx *sql.Tx, item *OrderItem) error {
	if item.Quantity == 0 {
		return nil
	}
	
	weight := math.Abs(item.NetWeight / item.Quantity)
	query := `UPDATE parts SET weight = $1 WHERE id = $2`
	_, err := tx.Exec(query, weight, item.ProductID)
	return err
}

func (os *OrderService) handleAdditionalOperations(tx *sql.Tx, form *OrderForm) error {
	// Add shipping address
	if err := os.addShippingAddress(tx, form); err != nil {
		return err
	}

	// Save document status
	if err := os.saveDocumentStatus(tx, form); err != nil {
		return err
	}

	// Update exchange rates if needed
	if form.Currency != form.DefaultCurrency {
		if err := os.updateExchangeRates(tx, form); err != nil {
			return err
		}
	}

	// Link to department
	if form.DepartmentID != nil && *form.DepartmentID > 0 {
		query := `INSERT INTO dpt_trans (trans_id, department_id) VALUES ($1, $2)`
		if _, err := tx.Exec(query, form.ID, *form.DepartmentID); err != nil {
			return err
		}
	}

	// Final inventory adjustments for orders
	if os.isOrder(form.Type) {
		multiplier := os.getMultiplier(form.Type)
		if err := os.adjustInventoryOnHand(tx, form, multiplier * -1); err != nil {
			return err
		}
	}

	return nil
}

func (os *OrderService) addShippingAddress(tx *sql.Tx, form *OrderForm) error {
	// Simplified shipping address handling
	log.Printf("Adding shipping address for order %d", form.ID)
	return nil
}

func (os *OrderService) saveDocumentStatus(tx *sql.Tx, form *OrderForm) error {
	// Save print/email status
	log.Printf("Saving document status for order %d", form.ID)
	return nil
}

func (os *OrderService) updateExchangeRates(tx *sql.Tx, form *OrderForm) error {
	// Update exchange rate table
	query := `
		INSERT INTO exchangerate (curr, transdate, buy, sell) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (curr, transdate) 
		DO UPDATE SET buy = $3, sell = $4`
	
	_, err := tx.Exec(query, form.Currency, form.TransactionDate, form.ExchangeRate, form.ExchangeRate)
	return err
}

func (os *OrderService) createAuditTrail(tx *sql.Tx, form *OrderForm) error {
	reference := form.OrderNumber
	if !os.isOrder(form.Type) {
		reference = form.QuoteNumber
	}

	query := `
		INSERT INTO audittrail (tablename, reference, formname, action, trans_id, employee_id, transdate)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err := tx.Exec(query,
		"oe",
		reference,
		string(form.Type),
		"saved",
		form.ID,
		form.EmployeeID,
		time.Now(),
	)
	
	return err
}

// Utility functions

func (os *OrderService) generateUID() string {
	return fmt.Sprintf("%d_%d", time.Now().Unix(), time.Now().Nanosecond())
}

func (os *OrderService) getMultiplier(orderType OrderType) int {
	if orderType == SalesOrder || orderType == SalesQuotation {
		return 1
	}
	return -1
}

func (os *OrderService) isOrder(orderType OrderType) bool {
	return orderType == SalesOrder || orderType == PurchaseOrder
}

func (os *OrderService) roundAmount(amount float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(amount * multiplier) / multiplier
}

func (os *OrderService) hasPackageInfo(item *OrderItem) bool {
	return item.Package != "" || item.NetWeight > 0 || item.GrossWeight > 0 || item.Volume > 0
}

// Example usage and HTTP handlers

import (
	"encoding/json"
	"net/http"
)

// OrderHandler handles HTTP requests for order operations
type OrderHandler struct {
	service *OrderService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service *OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// SaveOrder handles POST requests to save orders
func (oh *OrderHandler) SaveOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var form OrderForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err := oh.service.Save(&form); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save order: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id": form.ID,
		"message": "Order saved successfully",
	})
}

// Main function to demonstrate usage
func main() {
	// Database connection (replace with your actual connection string)
	db, err := sql.Open("postgres", "postgresql://user:password@localhost/dbname?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create service and handler
	orderService := NewOrderService(db)
	orderHandler := NewOrderHandler(orderService)

	// Setup HTTP routes
	http.HandleFunc("/api/orders", orderHandler.SaveOrder)
	
	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
