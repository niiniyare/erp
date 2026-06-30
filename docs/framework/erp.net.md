# ERP.net Core Tables — Plain-English Reference

Source: [docs.erp.net/model/tables](https://docs.erp.net/model/tables/)

This covers the tables that make up the *essentials* of any ERP system — the parts every ERP needs regardless of industry: a shared document model, parties (customers/suppliers/employees), products, the general ledger, inventory, tax, and point-of-sale. Each table is described in plain language with its real columns, so you can see exactly what data it holds and how it connects to other tables.

---

## 1. The Document Backbone

Almost everything in ERP.net — a sales order, an invoice, a GL voucher, a stock movement — is a specialized version of one shared concept: a **document**. Understanding this one relationship explains why so many unrelated-looking tables share fields like `Document_Id`.

### Gen_Documents — the parent of every transaction
Every business document in the system — sales orders, invoices, GL vouchers, stock transactions, etc. — has a row here. Specific document types (like `Acc_Vouchers` or `Inv_Transactions`) just add their own extra columns on top of this shared row. This is what lets the system give every document, no matter its type, the same numbering, status tracking, attachments, comments, and approval history for free.

---

## 2. Accounting (General Ledger)

This is the double-entry bookkeeping core. Money doesn't move in the system without leaving a trail here.

### Acc_Account_Groups — the chart of accounts, as a tree
A hierarchical folder structure for your accounts (e.g. "Assets" → "Current Assets" → "Cash"). Each group can have a parent group, so the whole chart of accounts is one tree.

| Column | What it holds |
|---|---|
| `Account_Group_Id` | Unique ID (primary key) |
| `Account_Group_Name` | Name, multi-language |
| `Account_Group_Number` | The group's number/code |
| `Parent_Account_Group_Id` | Points to the parent group; empty if this is a top-level group |
| `Full_Path` | The full chain of group numbers from root to here |
| `Discontinued` | If true, hide this group from pickers |

### Acc_Accounts — the actual ledger accounts (the leaves of the tree)
The individual accounts you actually post transactions to, like "1010 — Cash" or "4000 — Sales Revenue."

| Column | What it holds |
|---|---|
| `Account_Id` | Unique ID |
| `Account_Group_Id` | Which group (folder) this account belongs to |
| `Account_Number` | Number, unique within its group |
| `Account_Full_Number` | The group number + account number combined, unique system-wide |
| `Account_Name` | Name, multi-language |
| `Currency_Id` | If set, this account only accepts movements in this one currency |
| `Limit_To_Base_Currency` | If on, restricts the account to the company's base/reporting currency only |
| `Currency_Valuation_Method` | How non-base-currency amounts get converted for balance purposes |
| `Discontinued` | Retired accounts that shouldn't appear in pickers |

### Acc_Vouchers — a single journal entry / GL posting
The header of one accounting posting. It's deliberately thin — it inherits its number, date, and status from `Gen_Documents` — and just adds accounting-specific flags.

| Column | What it holds |
|---|---|
| `Voucher_Id` | Unique ID |
| `Document_Id` | Link back to the shared `Gen_Documents` row (where the date, number, status actually live) |
| `Description` | Free-text note on the voucher |
| `Default_Referenced_Document_Id` | Default source document for the voucher's lines |

### Acc_Voucher_Lines — the individual debit/credit rows of a journal entry
Every line of an accounting posting — one row per debit or credit. A balanced voucher has its debits equal its credits within each "correspondance group."

| Column | What it holds |
|---|---|
| `Voucher_Line_Id` | Unique ID |
| `Voucher_Id` | Which voucher this line belongs to |
| `Line_No` | Order of the line within the voucher |
| `Account_Id` | The account being debited or credited |
| `Cost_Center_Id` / `Profit_Center_Id` | Optional cost/profit center tags for management reporting |
| `Debit` / `Credit` | The amount, in the account's own currency |
| `Debit_Base` / `Credit_Base` | The same amount converted to the company's base currency |
| `Debit_Reporting` / `Credit_Reporting` | The same amount converted to a secondary "reporting" currency |
| `Currency_Id` | Currency of this specific line's movement |
| `Rate_Multiplier` / `Rate_Divisor` | The exchange rate used to get from line currency to base currency |
| `Correspondance_No` | Groups lines that must balance against each other within the voucher |
| `Item_Key` | An extra grouping key — `Account_Id + Item_Key` is the smallest unit the system tracks a running balance for |
| `Referenced_Document_Id` | The original business document (e.g. the invoice) that caused this posting |

### Acc_Templates — auto-posting rules
Rather than someone manually creating a voucher every time an invoice is issued, a template watches for a document event and automatically generates the matching GL voucher.

| Column | What it holds |
|---|---|
| `Template_Id` | Unique ID |
| `Template_Name` | Name |
| `Route_Id` | The workflow route that triggers this template |
| `Voucher_Date_Source` | Where the generated voucher's date comes from |
| `Voucher_Description_Mask` | A text template (with placeholders) for the generated voucher's description |

Each template owns a set of `Acc_Template_Lines` describing which accounts to debit/credit and how to calculate the amounts — that's the actual posting logic.

---

## 3. Inventory

Inventory is a perpetual, transaction-based ledger — very similar in spirit to the GL, but tracking quantities and costs of stock instead of money in accounts.

### Inv_Stores — physical or logical warehouses
A store is anywhere stock is held — a warehouse, a shop, a vehicle, a fuel tank. Each store is also registered as a `Gen_Parties` row, so it can be a counterparty on documents (e.g. "transfer from Store A to Store B").

| Column | What it holds |
|---|---|
| `Id` | Unique store ID |
| `Store_Code` / `Store_Name` | Code and name |
| `Party_Id` | Link to the store's `Gen_Parties` record |
| `Store_Group_Id` | Which group (folder) of stores this belongs to |
| `Currency_Id` | Currency used for this store's cost calculations |
| `Responsible_Party_Id` | Who (usually an employee) is accountable for stock here |
| `Default_Store_Bin_Id` | Fallback bin used when no specific bin is given |
| `Default_Supply_Store_Id` | The usual store this one gets restocked from |
| `Number_Of_Dimensions` | How many coordinates a bin location needs (0 = single-bin store) |
| `Tax_Warehouse_Id` | If set, this store is also a licensed excise tax warehouse |
| `Warehouse_Id` | If set, a `Wms_Warehouses` record manages detailed pick/pack operations here |
| `Unmanaged` | If true, the system auto-generates stock transactions instead of requiring manual entry |

### Inv_Current_Balances — the live "how much do we have" view
A real-time snapshot of stock on hand, grouped by store + product + lot + bin + serial number, with its current cost. This is the table you'd query to answer "what's our stock level right now."

| Column | What it holds |
|---|---|
| `Store_Id` / `Product_Id` / `Product_Variant_Id` | What's being counted, and where |
| `Lot_Id` / `Serial_Number_Id` / `Store_Bin_Id` | Optional finer-grained tracking |
| `Quantity_Base` | Quantity on hand, in the product's base unit of measure |
| `Store_Cost` / `Product_Cost` / `Base_Cost` | The value of that stock, expressed in three different currencies (store's, product's, company's base) |

### Inv_Transactions — one stock movement event (receipt or issue)
The header of a single inventory movement — like one delivery received or one shipment issued.

| Column | What it holds |
|---|---|
| `Transaction_Id` | Unique ID |
| `Document_Id` | Link to the shared `Gen_Documents` row |
| `Store_Id` | The store the goods moved into/out of |
| `Movement_Type` | `R` = receipt (stock coming in), `I` = issue (stock going out) |
| `Cost_Source` | Whether cost comes from the store's running average (`S`) or is specified on the document itself (`D`) — receipts usually specify cost, issues usually pull from the store average |
| `Is_Scrap` / `Scrap_Type_Id` | Marks a write-off and its reason |
| `Issuing_Person_Id` / `Receiving_Person_Id` | Who handed off / received the goods |

### Inv_Transaction_Lines — the per-product detail of a movement
One row per product within a transaction.

| Column | What it holds |
|---|---|
| `Transaction_Line_Id` | Unique ID |
| `Transaction_Id` | Which transaction (header) this belongs to |
| `Product_Id` / `Product_Variant_Id` | What item moved |
| `Quantity` / `Quantity_Unit_Id` | How much, in what unit |
| `Quantity_Base` | The same quantity converted to the base unit |
| `Unit_Cost` | Cost per unit |
| `Line_Cost` | Total cost for the line |
| `Line_Base_Cost` / `Line_Product_Cost` / `Line_Store_Cost` | The line's cost expressed in three currencies |
| `Lot_Id` / `Serial_Number_Id` / `Store_Bin_Id` | Which specific lot/serial/bin was affected |
| `Parent_Store_Order_Line_Id` | If this transaction fulfills an order line, which one |
| `Transaction_Timestamp` | The exact moment this line affected the product's running cost — this matters because cost is calculated chronologically |

---

## 4. Point of Sale

The retail/cash-register layer — built for high volume, walk-up transactions rather than the longer-cycle order-to-invoice flow used elsewhere.

### Pos_Sales — one till transaction (receipt)
The header of a single point-of-sale sale.

| Column | What it holds |
|---|---|
| `Pos_Sale_Id` | Unique ID |
| `Document_Number` / `Fiscal_Sales_Number` | The printed receipt number, and the separate fiscal-authority number if your country requires one |
| `Location_Id` / `Terminal_Id` | Which shop/site and which till it happened at |
| `Operator_Id` / `Opened_By_Id` / `Closed_By_Id` | Who's accountable for the sale, who opened it, who finalized it (can be three different people) |
| `Customer_Id` | Set only when the customer is known (e.g. loyalty card); usually empty for anonymous retail sales |
| `Sale_Date` | The business date used for daily reporting and accounting |
| `Opened_At` / `Closed_At` | Timestamps for when the sale started and finished |
| `Sale_Kind` | `SAL` = normal sale, `RET` = return, `INV` = invoice, `CRN` = credit note |
| `Sale_Stage` | `NEW` = still open, `FIN` = finalized (totals on header and lines must match once finalized) |
| `Is_Voided` / `Voided_At` / `Voided_By_Id` | Cancellation tracking |
| `Original_Sale_Id` / `Original_Sale_Number` | For a return/refund, which original sale it's reversing |
| `Payment_Type_Id` | Set only if there was exactly one payment method; multi-payment sales leave this empty and use `Pos_Sale_Payments` instead |
| `Total_Amount` / `Total_Amount_Base` / `Total_Amount_Reporting` | The sale total in sale currency, base currency, and reporting currency |
| `Sale_Currency_Id` | Which currency the sale was recorded in |
| `System_Message` | Logs/errors from processing this sale, useful for troubleshooting failed transactions |

`Pos_Sale_Lines` holds the individual items sold, and `Pos_Sale_Payments` holds the individual tenders (cash, card, etc.) when more than one payment method was used.

---

## 5. Tax (VAT)

A generic tax-determination engine: every taxable transaction gets classified by a "deal type," and the deal type drives both the rate applied and which line of the statutory tax return it ends up on.

### VAT_Deal_Types — the rules for "what kind of taxable event is this"
A classification system. Instead of hardcoding tax logic, every taxable transaction (a sale, a purchase, an import) is tagged with a deal type, and the deal type carries the tax rate category.

| Column | What it holds |
|---|---|
| `Deal_Type_Id` | Unique ID |
| `Deal_Type_Code` / `Deal_Type_Name` | Short code and description |
| `Entry_Type` | `S` = this deal type generates a sales-ledger entry, `P` = purchases-ledger entry |
| `Tax_Code` | Which rate category applies: `STD` (standard), `RED` (reduced), `SPR` (super-reduced), `INT` (intermediate/parking rate), `EXM` (exempt), `NS` (not subject to tax) |
| `Country_Id` | Which country's rules this deal type applies under |
| `Is_System` | True if it's a built-in deal type the user can't edit |

### VAT_Entries — the actual tax ledger
Every taxable transaction generates one row here — this is the data a VAT/sales-tax return is built from.

| Column | What it holds |
|---|---|
| `Entry_Id` | Unique ID |
| `Document_Id` | The source document that caused this entry |
| `Entry_Type` | Sales (`S`) or purchase (`P`) |
| `Deal_Type_Id` | Which deal type classified this transaction |
| `Party_Id` | The counterparty (customer or supplier) involved |
| `Registration_Number` / `Registration_VAT_Number` | The counterparty's national ID and tax registration number at the time of the transaction (stored as a snapshot, since the counterparty's own record could change later) |
| `Apply_Date` | The reporting period this entry counts toward (usually but not always the document date) |
| `Amount_Base` | The taxable amount, excluding tax, in base currency |
| `VAT_Amount_Base` | The tax amount itself, in base currency |
| `Cash_Reporting_Mode` | Whether this entry uses cash-basis VAT reporting instead of accrual-basis |
| `Referenced_Document_No` / `Referenced_Document_Type_Id` | The number and type of the original document, kept here even if the document itself is later deleted |

---

## 6. Excise (for excisable goods like fuel, alcohol, tobacco)

A specialized layer on top of inventory and tax for goods where a per-unit duty applies and movements must be tracked under government oversight — directly relevant to anything involving fuel, since petroleum products are excisable almost everywhere.

### Exc_Tax_Warehouses — a licensed bonded warehouse
A store that's been registered with customs/excise authorities as a place where excisable goods can be held, produced, or moved under duty suspension (i.e., before the duty has actually been paid).

| Column | What it holds |
|---|---|
| `Tax_Warehouse_Id` | Unique ID |
| `Name` | Multi-language name |
| `Enterprise_Company_Id` | Which legal entity owns this warehouse |
| `Trader_Excise_Number` | The excise ID of the warehouse's owner |
| `Tax_Warehouse_Excise_Number` | The excise ID of the warehouse itself |
| `Customs_Office` | Which customs office the warehouse reports to |

### Exc_Excise_Products — the standardized excise product codes
The official classification codes that tax authorities use for excisable goods (the EU uses 4-character codes like `T200` for certain tobacco products, `S200` for certain spirits).

| Column | What it holds |
|---|---|
| `Excise_Product_Id` | Unique ID |
| `Code` | The official classification code |
| `Name` | Multi-language description |
| `Excise_Product_Category_Id` | Broader category this code belongs to (fuels, alcohol, tobacco, etc.) |

### Exc_Excise_Duty_Rates — the actual duty rate per product
The per-unit tax rate that applies to a given excise product for a given purpose, with an effective date range.

| Column | What it holds |
|---|---|
| `Excise_Duty_Rate_Id` | Unique ID |
| `Excise_Product_Id` | Which product this rate applies to |
| `Excise_Purpose_Code_Id` | The purpose this rate applies to (rates can differ by intended use) |
| `Excise_Duty_Rate` | The actual rate |
| `Excise_Measurement_Unit_Id` | The unit the rate is expressed per (e.g. per litre) |
| `Valid_From` / `Valid_To` | Effective date range; a new rate with a later start date automatically supersedes the old one |

### Exc_Excise_Stamp_Lots — a batch of physical excise stamps received
Many countries require a physical tax stamp on excisable products (cigarette packs, spirit bottles); this table tracks a numbered batch of stamps as they're received from the authorities.

| Column | What it holds |
|---|---|
| `Excise_Stamp_Lot_Id` | Unique ID |
| `Excise_Product_Type_Id` | What type of product these stamps are for |
| `Batch_Number` / `Prefix` | Production batch identifiers |
| `Start_Number` / `End_Number` | The numeric range of stamps in this lot |
| `Quantity` | How many stamps are in the lot |
| `Purchase_Lot_Number` | The customs document reference under which the stamps were received |
| `Is_Active` | Whether this lot is still usable |

### Exc_Excise_Stamp_Operations — what happened to a batch of stamps
Records an event involving excise stamps (e.g. stamps applied to product, stamps damaged/destroyed, stamps returned) at a specific tax warehouse.

| Column | What it holds |
|---|---|
| `Excise_Stamp_Operation_Id` | Unique ID |
| `Document_Id` | Link to the shared document record |
| `Excise_Stamp_Operation_Type_Id` | What kind of stamp event this was |
| `Tax_Warehouse_Id` | Where it happened |

---

## 7. Asset & Maintenance Tracking

For physical equipment that needs scheduled servicing and whose usage needs to be measured over time — directly applicable to anything with meters, odometers, or running-hour counters.

### Eam_Managed_Assets — a piece of equipment under maintenance management
Any asset whose upkeep is being tracked: a vehicle, a pump, a generator, a piece of machinery.

| Column | What it holds |
|---|---|
| `Managed_Asset_Id` | Unique ID |
| `Managed_Asset_Code` / `Managed_Asset_Name` | Code and name |
| `Managed_Asset_Type_Id` | The type, which determines what parameters get tracked and what maintenance types apply |
| `Managed_Asset_Group_Id` | Organizational grouping |
| `Fixed_Asset_Id` | Link to the matching row in the fixed-asset register, if this is also a depreciable asset |
| `Registration_Number` | Official registration number (e.g. vehicle plate), if applicable |
| `Is_Active` | Whether it's currently in service |

### Eam_Tracked_Parameters — the kind of thing you measure on an asset
A reusable definition of a measurable parameter — like "odometer reading" or "engine hours" or "fuel level."

| Column | What it holds |
|---|---|
| `Tracked_Parameter_Id` | Unique ID |
| `Tracked_Parameter_Code` / `Tracked_Parameter_Name` | Code and name |
| `Is_Active` | Whether it's still in use |

### Eam_Managed_Asset_Parameter_Values — the actual readings over time
Every individual measurement taken for a given asset's tracked parameter — often fed automatically by IoT sensors.

| Column | What it holds |
|---|---|
| `Managed_Asset_Parameter_Value_Id` | Unique ID |
| `Managed_Asset_Id` | Which asset this reading is for |
| `Tracked_Parameter_Id` | Which parameter was measured |
| `Time_Utc` | When the reading was taken |
| `Value` | The reading itself |

### Eam_Managed_Asset_Maintenance_Schedules — when maintenance is due
Defines the recurrence rule for a specific asset's maintenance: how often it should happen, expressed as days, months, or a change in a tracked parameter (whichever comes first, depending on how it's configured).

| Column | What it holds |
|---|---|
| `Managed_Asset_Maintenance_Schedule_Id` | Unique ID |
| `Managed_Asset_Id` | Which asset |
| `Maintenance_Type_Id` | What kind of maintenance |
| `Schedule_Days` | Recur every N days, if set |
| `Schedule_Months` | Recur every N months, if set |
| `Parameter_Change_Delta` | Recur every time the tracked parameter changes by this much (e.g. every 5,000 km), if set |

---

## How these pieces fit together

A simple example walks through most of the tables above: a fuel delivery arrives at a depot.

1. The delivery is received into a **store** (`Inv_Stores`) that's also registered as a **tax warehouse** (`Exc_Tax_Warehouses`), because fuel is excisable.
2. Receiving the fuel creates an **inventory transaction** (`Inv_Transactions` + `Inv_Transaction_Lines`), which updates the store's **current balance** (`Inv_Current_Balances`).
3. Because the product is excisable, the duty owed is calculated using the matching **excise duty rate** (`Exc_Excise_Duty_Rates`) for that **excise product** (`Exc_Excise_Products`).
4. The purchase invoice for the delivery is tagged with a **VAT deal type** (`VAT_Deal_Types`), which generates a **VAT entry** (`VAT_Entries`) for the tax return.
5. An **accounting template** (`Acc_Templates`) automatically turns the purchase invoice into a **GL voucher** (`Acc_Vouchers` + `Acc_Voucher_Lines`), debiting inventory and crediting accounts payable.
6. Later, fuel sold at the pump becomes a **POS sale** (`Pos_Sales` + lines + payments), which in turn generates its own inventory issue transaction and its own VAT and GL entries.
7. The pump itself, as a **managed asset** (`Eam_Managed_Assets`), has its totalizer readings logged over time (`Eam_Managed_Asset_Parameter_Values`) and a **maintenance schedule** (`Eam_Managed_Asset_Maintenance_Schedules`) that flags when it's due for calibration.


Every one of those documents — the receipt, the invoice, the voucher, the POS sale — is, underneath, a row in the same shared `Gen_Documents` table, which is why the system can give all of them consistent numbering, status tracking, and audit history without each module having to reinvent it.


[UI desing](https://docs.erp.net/webclient/layouts-and-views/index.html) 
