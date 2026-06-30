package def

// NamingSeriesDef configures automatic document-number generation for an entity.
//
// The naming package increments a per-entity, per-tenant counter stored in
// naming_sequences and stamps the generated value onto Field during Create.
//
// # Prefix tokens
//
// The Prefix string may contain date placeholder tokens that are replaced at
// generation time. Supported tokens:
//
//	{YYYY}   — 4-digit year  (e.g. "2025")
//	{YY}     — 2-digit year  (e.g. "25")
//	{MM}     — 2-digit month (e.g. "06")
//	{DD}     — 2-digit day   (e.g. "30")
//
// When date tokens are present, consider setting ResetPeriod to reset the
// counter each year or month so sequence numbers stay short.
//
// Example:
//
//	NamingSeries: &def.NamingSeriesDef{
//	    Field:       "name",
//	    Prefix:      "INV-{YYYY}-{MM}-",
//	    Padding:     4,
//	    ResetPeriod: ResetMonthly,
//	}
//
// produces "INV-2025-06-0001", "INV-2025-06-0002", …
// resetting to "INV-2025-07-0001" the next month.
type NamingSeriesDef struct {
	// Field is the entity field that receives the generated value.
	// The field must exist in EntityDefinition.Fields.
	Field string

	// Prefix is prepended to the zero-padded sequence number.
	// May include date tokens: {YYYY}, {YY}, {MM}, {DD}.
	Prefix string

	// Padding is the minimum width of the numeric part, zero-padded.
	// Defaults to 5 when zero.
	Padding int

	// ResetPeriod controls when the sequence counter resets to 1.
	// Defaults to ResetNever (monotonically increasing).
	ResetPeriod ResetPeriod
}

// ResetPeriod controls the counter-reset frequency for a NamingSeriesDef.
type ResetPeriod string

const (
	// ResetNever means the counter increments indefinitely (default).
	ResetNever ResetPeriod = ""

	// ResetYearly resets the counter to 1 on the first use in each calendar year.
	ResetYearly ResetPeriod = "yearly"

	// ResetMonthly resets the counter to 1 on the first use in each calendar month.
	ResetMonthly ResetPeriod = "monthly"
)
