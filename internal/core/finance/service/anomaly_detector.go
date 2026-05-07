package service

// anomaly_detector.go — Lightweight heuristic + statistical anomaly detection
// for the finance module.
//
// AnomalyDetector observes finance mutation events and emits structured metrics
// when suspicious patterns are detected. Detection is intentionally non-blocking:
// SafetyEnforcer handles hard limits; this layer handles statistical anomalies.
//
// ── Design philosophy ────────────────────────────────────────────────────────
//
// This detector never returns errors and never rejects a mutation. Its only job
// is to increment Prometheus counters so that the alerting layer
// (Prometheus rules → Grafana → PagerDuty) can decide what a suspicious volume
// looks like at runtime, without a code change. Thresholds and alert sensitivity
// belong in the alerting config, not here.
//
// All methods are nil-safe so the detector can be omitted in lightweight test
// setups without requiring a mock.
//
// ── Backward compatibility ───────────────────────────────────────────────────
//
// NewAnomalyDetector(m) is unchanged. All detection algorithms are opt-in:
// they are disabled by default and activated only when the corresponding
// With* option is passed to NewAnomalyDetectorWithOptions. No existing call
// site needs to be modified.
//
// ── Detection algorithms ─────────────────────────────────────────────────────
//
//	Existing
//	  LARGE_TRANSACTION     Single posting exceeds largeTxnThreshold.
//
//	Original opt-in (unchanged)
//	  HIGH_VELOCITY         Posting rate per tenant exceeds N events in a window.
//	  REVERSAL_BURST        Reversal rate per user exceeds N events in a window.
//	  ROUND_AMOUNT          Posting divisible by a configurable quantum.
//	  OFF_HOURS_POSTING     Posting falls outside configured business hours.
//	  DUPLICATE_AMOUNT      Same amount posted N+ times by one tenant in a window.
//
//	New opt-in (this revision)
//	  STATISTICAL_OUTLIER   Posting amount deviates > N σ from the tenant's own
//	                        historical mean (Welford online algorithm).
//	  VELOCITY_ACCELERATION Short-window posting rate is > N× the long-window
//	                        baseline — catches slow-start → sudden-spike attacks.
//	  CROSS_TENANT_SPREAD   A single user posts to > N distinct tenants in a window,
//	                        indicating possible credential compromise or insider
//	                        lateral movement.
//	  AMOUNT_SEQUENCE       Recent posting amounts form an arithmetic progression
//	                        (constant delta across K consecutive postings) — the
//	                        hallmark of structured layering / smurfing.

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// ═══════════════════════════════════════════════════════════════════════════════
// ── Internal helpers — original ──────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// velocityTracker counts events in a per-key rolling time window.
// The key is an opaque string (tenant UUID, user UUID, …).
// All methods are safe for concurrent use.
//
// Memory note: timestamps for a key are pruned on every record() call. In
// production the number of distinct tenants / users is bounded and small, so
// unbounded map growth is not a practical concern. If cardinality ever grows
// large, replace the map with an LRU cache.
type velocityTracker struct {
	mu         sync.Mutex
	window     time.Duration
	max        int
	timestamps map[string][]time.Time
}

func newVelocityTracker(window time.Duration, max int) *velocityTracker {
	return &velocityTracker{
		window:     window,
		max:        max,
		timestamps: make(map[string][]time.Time),
	}
}

// record appends now for key, evicts timestamps older than the window, and
// returns the resulting in-window count. The caller decides whether to fire
// an anomaly based on the returned count vs. vt.max.
func (vt *velocityTracker) record(key string, now time.Time) (count int, exceeded bool) {
	cutoff := now.Add(-vt.window)

	vt.mu.Lock()
	defer vt.mu.Unlock()

	// Compact in-place: keep only timestamps within the window.
	existing := vt.timestamps[key]
	fresh := existing[:0] // reuse backing array — avoids allocation on the hot path
	for _, t := range existing {
		if t.After(cutoff) {
			fresh = append(fresh, t)
		}
	}
	fresh = append(fresh, now)
	vt.timestamps[key] = fresh

	count = len(fresh)
	exceeded = count > vt.max
	return count, exceeded
}

// duplicateTracker detects repeated identical amounts per tenant in a rolling
// window. It is a two-level velocityTracker: tenant → amount → []time.Time.
type duplicateTracker struct {
	mu      sync.Mutex
	window  time.Duration
	max     int
	history map[string]map[string][]time.Time // tenantID → amountKey → []time.Time
}

func newDuplicateTracker(window time.Duration, max int) *duplicateTracker {
	return &duplicateTracker{
		window:  window,
		max:     max,
		history: make(map[string]map[string][]time.Time),
	}
}

// record returns the number of times amountKey appeared for tenantKey within
// the window (including now) and whether that count exceeds max.
func (dt *duplicateTracker) record(tenantKey, amountKey string, now time.Time) (count int, exceeded bool) {
	cutoff := now.Add(-dt.window)

	dt.mu.Lock()
	defer dt.mu.Unlock()

	if dt.history[tenantKey] == nil {
		dt.history[tenantKey] = make(map[string][]time.Time)
	}

	existing := dt.history[tenantKey][amountKey]
	fresh := existing[:0]
	for _, t := range existing {
		if t.After(cutoff) {
			fresh = append(fresh, t)
		}
	}
	fresh = append(fresh, now)
	dt.history[tenantKey][amountKey] = fresh

	count = len(fresh)
	exceeded = count > dt.max
	return count, exceeded
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── Internal helpers — NEW ───────────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// ── statisticsTracker (STATISTICAL_OUTLIER) ──────────────────────────────────

// statisticsTracker maintains a per-key online mean and variance using
// Welford's numerically stable single-pass algorithm. Unlike a windowed
// tracker it accumulates an unbounded historical baseline, which is exactly
// what is needed here: a posting that deviates from a tenant's long-run norm
// is suspicious even if the rolling window shows no velocity spike.
//
// The z-score is computed against the baseline BEFORE incorporating the new
// observation, so each posting is judged against existing history rather than
// a retroactively adjusted mean. A cold-start guard (minSample) suppresses
// flags until there is enough data for reliable z-score estimation.
//
// Reference: Welford, B.P. (1962). "Note on a method for calculating corrected
// sums of squares and products." Technometrics 4(3): 419–420.
type statisticsTracker struct {
	mu        sync.Mutex
	threshold float64 // fire when |z-score| exceeds this (e.g. 3.0 for 3σ)
	minSample int64   // observations required before flagging (cold-start guard)
	state     map[string]*welfordState
}

// welfordState holds the three accumulators required by Welford's algorithm.
// n is the sample count; mean is the running mean; m2 is the sum of squared
// deviations from the running mean (from which variance = m2/(n-1)).
type welfordState struct {
	n    int64
	mean float64
	m2   float64
}

func newStatisticsTracker(zScoreThreshold float64, minSample int64) *statisticsTracker {
	return &statisticsTracker{
		threshold: zScoreThreshold,
		minSample: minSample,
		state:     make(map[string]*welfordState),
	}
}

// observe scores x against the existing baseline for key, then incorporates x
// into the baseline.  Returns the absolute z-score and whether it exceeds the
// threshold.  Returns (0, false) during the warm-up period or when all
// historical values are identical (zero variance).
func (st *statisticsTracker) observe(key string, x float64) (zScore float64, outlier bool) {
	st.mu.Lock()
	defer st.mu.Unlock()

	s, ok := st.state[key]
	if !ok {
		s = &welfordState{}
		st.state[key] = s
	}

	// Score BEFORE updating — the new value is judged against the existing
	// baseline, not one that has already absorbed the current observation.
	if s.n >= st.minSample && s.m2 > 0 {
		variance := s.m2 / float64(s.n-1) // sample variance (Bessel-corrected)
		stddev := math.Sqrt(variance)
		zScore = math.Abs((x - s.mean) / stddev)
		outlier = zScore > st.threshold
	}

	// Welford update — order matters: delta uses the OLD mean, then the mean
	// is advanced, then m2 accumulates the product of the two deltas.
	s.n++
	delta := x - s.mean
	s.mean += delta / float64(s.n)
	s.m2 += delta * (x - s.mean) // (x - newMean) after the mean has been updated

	return zScore, outlier
}

// ── dualWindowTracker (VELOCITY_ACCELERATION) ────────────────────────────────

// dualWindowTracker detects a sudden acceleration in event rate by comparing a
// short-term window against a long-term baseline window. The ratio
//
//	shortRate (events/s) / longRate (events/s)
//
// spikes when a tenant that has been posting at a normal background rate
// suddenly bursts — a pattern that HIGH_VELOCITY alone would miss if the
// absolute count in the short window has not yet crossed its threshold.
//
// Both rates are normalised to events-per-second so that windows of different
// lengths are directly comparable.
type dualWindowTracker struct {
	mu             sync.Mutex
	shortWindow    time.Duration
	longWindow     time.Duration
	minRatio       float64 // fire when shortRate/longRate exceeds this value
	minEventsShort int     // minimum short-window count before evaluating ratio
	timestamps     map[string][]time.Time
}

func newDualWindowTracker(
	shortWindow, longWindow time.Duration,
	minRatio float64,
	minEventsShort int,
) *dualWindowTracker {
	return &dualWindowTracker{
		shortWindow:    shortWindow,
		longWindow:     longWindow,
		minRatio:       minRatio,
		minEventsShort: minEventsShort,
		timestamps:     make(map[string][]time.Time),
	}
}

// record registers an event for key at now, evicts entries older than
// longWindow, and returns the current short/long rate ratio and whether it
// exceeded minRatio.
func (dw *dualWindowTracker) record(key string, now time.Time) (ratio float64, exceeded bool) {
	shortCutoff := now.Add(-dw.shortWindow)
	longCutoff := now.Add(-dw.longWindow)

	dw.mu.Lock()
	defer dw.mu.Unlock()

	existing := dw.timestamps[key]
	fresh := existing[:0]
	for _, t := range existing {
		if t.After(longCutoff) {
			fresh = append(fresh, t)
		}
	}
	fresh = append(fresh, now)
	dw.timestamps[key] = fresh

	// Count events in each window in a single pass.
	var shortCount, longCount int
	for _, t := range fresh {
		longCount++
		if t.After(shortCutoff) {
			shortCount++
		}
	}

	// Guard: require a minimum short-window count to avoid false positives from
	// single-event bursts in otherwise quiescent tenants.
	if shortCount < dw.minEventsShort {
		return 0, false
	}

	// Normalise both windows to events-per-second before computing the ratio.
	shortRate := float64(shortCount) / dw.shortWindow.Seconds()
	longRate := float64(longCount) / dw.longWindow.Seconds()
	if longRate == 0 {
		return 0, false
	}

	ratio = shortRate / longRate
	exceeded = ratio > dw.minRatio
	return ratio, exceeded
}

// ── spreadTracker (CROSS_TENANT_SPREAD) ─────────────────────────────────────

// spreadTracker tracks how many distinct "targets" (e.g. tenant IDs) a single
// "actor" (e.g. user ID) has touched within a rolling window.
//
// A user who posts across an unusual number of tenants in a short period
// deviates from normal operator behaviour. It may indicate:
//   - Credential compromise where the attacker probes all tenants attached
//     to the stolen session.
//   - A misconfigured automation that is writing to the wrong tenants.
//   - An insider threat conducting reconnaissance across subsidiaries.
//
// Memory layout: actor → []spreadEntry (value + timestamp), compacted on
// each record() call. Distinct values are counted with a local map[string]struct{}
// which is cheap because max is small (typically 3–5).
type spreadTracker struct {
	mu      sync.Mutex
	window  time.Duration
	max     int
	history map[string][]spreadEntry // actor → [(target, time)]
}

type spreadEntry struct {
	value string
	at    time.Time
}

func newSpreadTracker(window time.Duration, max int) *spreadTracker {
	return &spreadTracker{
		window:  window,
		max:     max,
		history: make(map[string][]spreadEntry),
	}
}

// record logs that actor touched target at now and returns the number of
// distinct targets seen in the window (including this call) and whether it
// exceeds max.
func (st *spreadTracker) record(actor, target string, now time.Time) (distinctCount int, exceeded bool) {
	cutoff := now.Add(-st.window)

	st.mu.Lock()
	defer st.mu.Unlock()

	entries := st.history[actor]
	fresh := entries[:0]
	for _, e := range entries {
		if e.at.After(cutoff) {
			fresh = append(fresh, e)
		}
	}
	fresh = append(fresh, spreadEntry{value: target, at: now})
	st.history[actor] = fresh

	// Count distinct targets using an inline set.
	seen := make(map[string]struct{}, len(fresh))
	for _, e := range fresh {
		seen[e.value] = struct{}{}
	}
	distinctCount = len(seen)
	exceeded = distinctCount > st.max
	return distinctCount, exceeded
}

// ── sequenceTracker (AMOUNT_SEQUENCE) ───────────────────────────────────────

// sequenceTracker detects arithmetic progressions in recent posting amounts
// for a tenant within a rolling window.
//
// Structured layering (a.k.a. smurfing) is a common technique where a
// perpetrator splits a large sum into a series of incrementally increasing
// amounts to stay below approval or reporting thresholds. Classic examples:
//
//	9 000 → 9 500 → 10 000 → 10 500   (delta = 500 each step)
//	5 000 → 10 000 → 15 000 → 20 000  (delta = 5 000 each step)
//
// The detector counts the length of the longest arithmetic run at the tail of
// the per-tenant history (i.e. the most recent consecutive pairs that share a
// constant delta). A run of minLength equal-delta pairs (minLength+1 amounts)
// triggers the flag.
//
// The delta is derived from the final two amounts in the window, so the check
// is directionally-aware and non-parametric: no delta magnitude threshold is
// required. A delta of zero is intentionally excluded because it is captured
// by DUPLICATE_AMOUNT.
type sequenceTracker struct {
	mu        sync.Mutex
	window    time.Duration
	minLength int // minimum consecutive equal-delta pairs required to flag
	entries   map[string][]sequenceEntry
}

type sequenceEntry struct {
	amount decimal.Decimal
	at     time.Time
}

func newSequenceTracker(window time.Duration, minLength int) *sequenceTracker {
	return &sequenceTracker{
		window:    window,
		minLength: minLength,
		entries:   make(map[string][]sequenceEntry),
	}
}

// record appends amount for key and returns the length of the longest
// consecutive arithmetic run at the tail of the window-compacted history, plus
// whether that run meets or exceeds minLength.
func (st *sequenceTracker) record(key string, amount decimal.Decimal, now time.Time) (runLength int, detected bool) {
	cutoff := now.Add(-st.window)

	st.mu.Lock()
	defer st.mu.Unlock()

	existing := st.entries[key]
	fresh := existing[:0]
	for _, e := range existing {
		if e.at.After(cutoff) {
			fresh = append(fresh, e)
		}
	}
	fresh = append(fresh, sequenceEntry{amount: amount, at: now})
	st.entries[key] = fresh

	n := len(fresh)
	if n < 2 {
		return 0, false
	}

	// Derive the reference delta from the most recent consecutive pair.
	// We scan backward through the slice (chronological order) counting how
	// many pairs share the same delta. We stop at the first mismatch because
	// we are looking for a contiguous tail run, not the best subsequence.
	delta := fresh[n-1].amount.Sub(fresh[n-2].amount)

	// A zero delta means identical amounts — that is DUPLICATE_AMOUNT territory.
	if delta.IsZero() {
		return 0, false
	}

	run := 1 // we already have one pair (the last two entries)
	for i := n - 2; i >= 1; i-- {
		d := fresh[i].amount.Sub(fresh[i-1].amount)
		if d.Equal(delta) {
			run++
		} else {
			break // progression broken — no need to look further back
		}
	}

	return run, run >= st.minLength
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── AnomalyDetector ──────────────────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// AnomalyDetector observes finance mutations and detects suspicious behavioral
// patterns.
//
// It is purely metrics-driven: each observation increments a Prometheus counter
// and the alerting layer decides when that volume warrants a page. The detector
// itself never blocks, panics, or returns an error — it is always a side-effect,
// never a gate.
//
// Two families of counter are emitted:
//
//   - Event-specific counters (finance_anomaly_postings_total,
//     _reversals_total, …) give per-signal granularity for dashboards.
//
//   - finance_anomaly_events_total with a "kind" label is a roll-up used by a
//     single alerting rule so the on-call engineer only needs to watch one
//     counter for all anomaly classes.
//
// All detection algorithms are opt-in via With* constructor options. When an
// option has not been applied, the corresponding field is nil / zero and the
// check short-circuits without any allocation or lock contention.
type AnomalyDetector struct {
	// ── core fields (unchanged) ───────────────────────────────────────────

	// largeTxnThreshold is the debit amount above which a posting is flagged.
	// This is a soft limit — the posting is not blocked; a counter is
	// incremented and a warning is logged. SafetyEnforcer enforces the hard
	// cap (currently 10 M); this threshold sits intentionally below it at 1 M
	// to give finance ops an early-warning buffer before the hard wall.
	largeTxnThreshold decimal.Decimal

	// metrics is the telemetry sink.  Production code uses a Prometheus adapter;
	// tests can substitute a no-op or a recording double.
	metrics metrics.MetricsProvider

	// ── original opt-in fields ────────────────────────────────────────────

	// now is an injectable clock used by all time-sensitive checks.  Production
	// code uses time.Now; tests can substitute a deterministic function.
	now func() time.Time

	// postingVelocity detects HIGH_VELOCITY.  nil when WithPostingVelocity has
	// not been applied.
	postingVelocity *velocityTracker

	// reversalVelocity detects REVERSAL_BURST.  nil when WithReversalVelocity
	// has not been applied.
	reversalVelocity *velocityTracker

	// roundAmountQuantum detects ROUND_AMOUNT.  IsZero() when
	// WithRoundAmountDetection has not been applied.
	//
	// A quantum of 10 000 flags every multiple of KES 10 000.  Rationale:
	// perpetrators often use psychologically salient round numbers to stay
	// below approval thresholds; clustering at round amounts is a known fraud
	// marker in East African markets.
	roundAmountQuantum decimal.Decimal

	// businessTimezone, businessHoursStart, businessHoursEnd define the window
	// within which postings are expected.  A nil timezone disables the check.
	// businessHoursStart is inclusive, businessHoursEnd is exclusive (24 h).
	//
	// Example: start=8, end=18, tz=Africa/Nairobi → 08:00–17:59 EAT.
	businessTimezone   *time.Location
	businessHoursStart int
	businessHoursEnd   int

	// duplicateAmounts detects DUPLICATE_AMOUNT.  nil when
	// WithDuplicateAmountDetection has not been applied.
	duplicateAmounts *duplicateTracker

	// ── NEW opt-in fields ─────────────────────────────────────────────────

	// outlierStats detects STATISTICAL_OUTLIER via Welford's online algorithm.
	// nil when WithStatisticalOutlierDetection has not been applied.
	outlierStats *statisticsTracker

	// velocityAccel detects VELOCITY_ACCELERATION via a dual-window ratio.
	// nil when WithVelocityAcceleration has not been applied.
	velocityAccel *dualWindowTracker

	// crossTenantSpread detects CROSS_TENANT_SPREAD: a user posting to an
	// unusual number of tenants in a rolling window.
	// nil when WithCrossTenantSpread has not been applied.
	crossTenantSpread *spreadTracker

	// amountSequence detects AMOUNT_SEQUENCE: an arithmetic progression in
	// recent posting amounts (structuring / smurfing marker).
	// nil when WithAmountSequenceDetection has not been applied.
	amountSequence *sequenceTracker
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── Functional options ───────────────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// Option is a functional modifier applied to AnomalyDetector during
// construction. All With* functions return an Option. Callers combine options
// freely; ordering does not matter unless two options configure the same field
// (last one wins).
type Option func(*AnomalyDetector)

// ── original options (signatures unchanged) ───────────────────────────────────

// WithLargeTxnThreshold overrides the default 1 M large-transaction threshold.
// Pass this to align with your chart-of-accounts materiality level.
func WithLargeTxnThreshold(threshold decimal.Decimal) Option {
	return func(d *AnomalyDetector) {
		d.largeTxnThreshold = threshold
	}
}

// WithPostingVelocity enables HIGH_VELOCITY detection: if more than maxPostings
// occur for a single tenant within window, a HIGH_VELOCITY anomaly counter is
// incremented. Typical production values: window=1*time.Minute, maxPostings=100.
func WithPostingVelocity(window time.Duration, maxPostings int) Option {
	return func(d *AnomalyDetector) {
		d.postingVelocity = newVelocityTracker(window, maxPostings)
	}
}

// WithReversalVelocity enables REVERSAL_BURST detection: if a single user
// issues more than maxReversals reversals within window, a REVERSAL_BURST
// anomaly counter is incremented.
// Typical values: window=5*time.Minute, maxReversals=5.
func WithReversalVelocity(window time.Duration, maxReversals int) Option {
	return func(d *AnomalyDetector) {
		d.reversalVelocity = newVelocityTracker(window, maxReversals)
	}
}

// WithRoundAmountDetection enables ROUND_AMOUNT detection: any posting whose
// debit amount is divisible by quantum (and positive) is flagged.
// A quantum of 10 000 flags every multiple of KES 10 000.
func WithRoundAmountDetection(quantum decimal.Decimal) Option {
	return func(d *AnomalyDetector) {
		d.roundAmountQuantum = quantum
	}
}

// WithOffHoursDetection enables OFF_HOURS_POSTING detection.  Postings whose
// wall-clock time (in tz) falls outside [startHour, endHour) are flagged.
// startHour and endHour are 24-hour integers (0–23).  If tz is nil, time.UTC
// is used.
//
// Example:
//
//	WithOffHoursDetection(8, 18, mustLoadLocation("Africa/Nairobi"))
//	// flags any posting before 08:00 or from 18:00 onward EAT
func WithOffHoursDetection(startHour, endHour int, tz *time.Location) Option {
	return func(d *AnomalyDetector) {
		if tz == nil {
			tz = time.UTC
		}
		d.businessTimezone = tz
		d.businessHoursStart = startHour
		d.businessHoursEnd = endHour
	}
}

// WithDuplicateAmountDetection enables DUPLICATE_AMOUNT detection: if the same
// debit amount appears more than maxOccurrences times for one tenant within
// window, an anomaly counter is incremented.  This catches duplicate CSV
// imports and naive replay attacks.
// Typical values: window=10*time.Minute, maxOccurrences=3.
func WithDuplicateAmountDetection(window time.Duration, maxOccurrences int) Option {
	return func(d *AnomalyDetector) {
		d.duplicateAmounts = newDuplicateTracker(window, maxOccurrences)
	}
}

// WithClock overrides the wall clock used by all time-sensitive checks.  Use
// this in tests to drive time deterministically without real sleeps.
//
//	detector := NewAnomalyDetectorWithOptions(m,
//	    WithOffHoursDetection(8, 18, time.UTC),
//	    WithClock(func() time.Time { return fixedTime }),
//	)
func WithClock(fn func() time.Time) Option {
	return func(d *AnomalyDetector) {
		d.now = fn
	}
}

// ── NEW options ───────────────────────────────────────────────────────────────

// WithStatisticalOutlierDetection enables STATISTICAL_OUTLIER detection using
// Welford's online algorithm to maintain a per-tenant running mean and sample
// variance.
//
// A posting whose amount deviates more than zScoreThreshold standard deviations
// from the tenant's own historical mean triggers the anomaly counter.  The
// baseline is built from all postings observed since the process started —
// not from a rolling window — so it reflects a tenant's true long-run norm
// rather than a short-term snapshot.
//
// Parameters:
//   - zScoreThreshold: number of standard deviations above which an amount is
//     considered an outlier. 3.0 is a reasonable starting point (flags the
//     ~0.3 % of a normal distribution that falls outside 3σ). More conservative
//     deployments may prefer 4.0 or 5.0 to reduce alert volume.
//   - minSampleSize: minimum number of postings for a tenant before any
//     outlier flag is emitted.  This prevents false positives during the warm-up
//     period when the mean and variance estimates are unreliable.  30 is a
//     commonly used minimum for sample-based statistics.
//
// Example:
//
//	WithStatisticalOutlierDetection(3.0, 30)
func WithStatisticalOutlierDetection(zScoreThreshold float64, minSampleSize int64) Option {
	return func(d *AnomalyDetector) {
		d.outlierStats = newStatisticsTracker(zScoreThreshold, minSampleSize)
	}
}

// WithVelocityAcceleration enables VELOCITY_ACCELERATION detection by comparing
// a short-term event rate against a longer-term baseline rate.
//
// Motivation: HIGH_VELOCITY catches sustained high throughput. VELOCITY_ACCELERATION
// catches the "slow start → sudden spike" pattern — a tenant that normally posts
// 10 times per minute and suddenly posts 200 times in 10 seconds without exceeding
// the per-minute limit. This pattern is characteristic of automated injection attacks
// and batch-import bugs that start gradually before ramping up.
//
// Parameters:
//   - shortWindow: the recent window whose rate is compared against the baseline.
//     Typical value: 30*time.Second.
//   - longWindow: the baseline window.  Must be longer than shortWindow.
//     Typical value: 10*time.Minute.
//   - minRatio: fire when shortRate/longRate exceeds this value.
//     A ratio of 5 means "the tenant is posting 5× faster than its recent norm".
//   - minEventsShort: minimum events in shortWindow before evaluating the ratio.
//     Prevents false positives when a single event in a quiet window produces an
//     arbitrarily large ratio. Typical value: 5.
//
// Example:
//
//	WithVelocityAcceleration(30*time.Second, 10*time.Minute, 5.0, 5)
func WithVelocityAcceleration(shortWindow, longWindow time.Duration, minRatio float64, minEventsShort int) Option {
	return func(d *AnomalyDetector) {
		d.velocityAccel = newDualWindowTracker(shortWindow, longWindow, minRatio, minEventsShort)
	}
}

// WithCrossTenantSpread enables CROSS_TENANT_SPREAD detection.
//
// A user whose session or credentials are legitimately configured will only
// ever post within a small, stable set of tenants. A user who touches more
// than maxTenants distinct tenants within window is behaving anomalously —
// either through credential compromise, misconfigured automation, or insider
// reconnaissance across subsidiaries.
//
// This check extracts the user ID from the request context (shared.GetUserID).
// Requests arriving without a user ID in context are keyed to the zero UUID,
// which is itself a detectable signal (unauthenticated postings should not exist).
//
// Parameters:
//   - window: rolling observation window. Typical value: 1*time.Hour.
//   - maxTenants: maximum distinct tenants before flagging. Typical value: 3
//     (most operators work within a single tenant; system accounts may span 2–3).
//
// Example:
//
//	WithCrossTenantSpread(time.Hour, 3)
func WithCrossTenantSpread(window time.Duration, maxTenants int) Option {
	return func(d *AnomalyDetector) {
		d.crossTenantSpread = newSpreadTracker(window, maxTenants)
	}
}

// WithAmountSequenceDetection enables AMOUNT_SEQUENCE detection.
//
// Structured layering (smurfing) involves splitting a large payment into a
// series of smaller amounts that step by a constant delta to stay below
// approval or reporting thresholds. For example:
//
//	9 000 → 9 500 → 10 000 → 10 500   (delta = KES 500)
//
// This detector maintains a per-tenant history of posting amounts within the
// window. It scans backward from the most recent posting and counts consecutive
// pairs that share the same arithmetic delta. When that run meets or exceeds
// minSequenceLength, AMOUNT_SEQUENCE is fired.
//
// The detector is directionally-aware and non-parametric: no delta magnitude
// threshold is required, and both increasing and decreasing sequences are
// caught. A zero delta (identical amounts) is excluded because that is the
// domain of DUPLICATE_AMOUNT.
//
// Parameters:
//   - window: rolling history window for a tenant's amounts. Typical: 30*time.Minute.
//   - minSequenceLength: consecutive equal-delta pairs required to fire.
//     minSequenceLength=3 means four amounts with the same step between each
//     (e.g. 9 000, 9 500, 10 000, 10 500). Typical value: 3.
//
// Example:
//
//	WithAmountSequenceDetection(30*time.Minute, 3)
func WithAmountSequenceDetection(window time.Duration, minSequenceLength int) Option {
	return func(d *AnomalyDetector) {
		d.amountSequence = newSequenceTracker(window, minSequenceLength)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── Constructors ─────────────────────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// NewAnomalyDetector creates a detector with production defaults.
// Signature is unchanged from the original; all existing call sites compile
// without modification. To activate the new detection algorithms, use
// NewAnomalyDetectorWithOptions.
func NewAnomalyDetector(m metrics.MetricsProvider) *AnomalyDetector {
	return NewAnomalyDetectorWithOptions(m)
}

// NewAnomalyDetectorWithOptions creates a detector with production defaults and
// applies each option in sequence. Only the algorithms whose With* option is
// provided are active; all others are disabled at zero cost (nil pointer checks
// on the hot path, no goroutines, no allocations).
//
// Example wiring (production, all algorithms enabled):
//
//	nairobi, _ := time.LoadLocation("Africa/Nairobi")
//	detector := NewAnomalyDetectorWithOptions(
//	    metricsProvider,
//	    // ── original ────────────────────────────────────────────────────
//	    WithPostingVelocity(time.Minute, 200),
//	    WithReversalVelocity(5*time.Minute, 10),
//	    WithRoundAmountDetection(decimal.NewFromInt(10_000)),
//	    WithOffHoursDetection(8, 18, nairobi),
//	    WithDuplicateAmountDetection(10*time.Minute, 3),
//	    // ── new ─────────────────────────────────────────────────────────
//	    WithStatisticalOutlierDetection(3.0, 30),
//	    WithVelocityAcceleration(30*time.Second, 10*time.Minute, 5.0, 5),
//	    WithCrossTenantSpread(time.Hour, 3),
//	    WithAmountSequenceDetection(30*time.Minute, 3),
//	)
func NewAnomalyDetectorWithOptions(m metrics.MetricsProvider, opts ...Option) *AnomalyDetector {
	d := &AnomalyDetector{
		// 1 M is suspicious but not necessarily invalid.  SafetyEnforcer blocks at
		// 10 M; the 9 M gap lets finance ops review outlier postings before they
		// hit the hard wall.
		largeTxnThreshold: decimal.NewFromInt(1_000_000),
		metrics:           m,
		now:               time.Now,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── Public Observe* methods (all signatures unchanged) ───────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// ObservePosting records every posting event.  In addition to the original
// LARGE_TRANSACTION check it runs up to eight opt-in checks:
//
//   - HIGH_VELOCITY           (WithPostingVelocity)
//   - ROUND_AMOUNT            (WithRoundAmountDetection)
//   - OFF_HOURS_POSTING       (WithOffHoursDetection)
//   - DUPLICATE_AMOUNT        (WithDuplicateAmountDetection)
//   - STATISTICAL_OUTLIER     (WithStatisticalOutlierDetection)
//   - VELOCITY_ACCELERATION   (WithVelocityAcceleration)
//   - CROSS_TENANT_SPREAD     (WithCrossTenantSpread)
//   - AMOUNT_SEQUENCE         (WithAmountSequenceDetection)
//
// Each check is a no-op when its option has not been applied.
func (d *AnomalyDetector) ObservePosting(ctx context.Context, totalDebit decimal.Decimal) {
	if d == nil {
		return
	}

	// tenantID may be zero-value if the context is unauthenticated; the counter
	// label will be the zero UUID in that case, which is itself a detectable anomaly.
	tenantID, _ := shared.GetTenantID(ctx)

	d.metrics.IncrementCounter("finance_anomaly_postings_total", metrics.Fields{
		"tenant_id": tenantID.String(),
	})

	// ── existing: large-transaction check ────────────────────────────────────
	if totalDebit.GreaterThan(d.largeTxnThreshold) {
		d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
			"kind":      "LARGE_TRANSACTION",
			"tenant_id": tenantID.String(),
		})
		logger.WarnContext(ctx, "finance anomaly: large transaction posted — review recommended", logger.Fields{
			"amount":    totalDebit.StringFixed(2),
			"threshold": d.largeTxnThreshold.StringFixed(2),
		})
	}

	// ── original opt-in checks ────────────────────────────────────────────────
	d.checkPostingVelocity(ctx, tenantID)
	d.checkRoundAmount(ctx, tenantID, totalDebit)
	d.checkOffHours(ctx, tenantID)
	d.checkDuplicateAmount(ctx, tenantID, totalDebit)

	// ── new opt-in checks ─────────────────────────────────────────────────────
	d.checkStatisticalOutlier(ctx, tenantID, totalDebit)
	d.checkVelocityAcceleration(ctx, tenantID)
	d.checkCrossTenantSpread(ctx, tenantID)
	d.checkAmountSequence(ctx, tenantID, totalDebit)
}

// ObserveReversal records a reversal event for churn detection.
// With WithReversalVelocity applied it additionally fires REVERSAL_BURST when
// a single user's reversal rate exceeds the configured threshold.
func (d *AnomalyDetector) ObserveReversal(ctx context.Context, byUserID uuid.UUID) {
	if d == nil {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_reversals_total", metrics.Fields{
		"user_id": byUserID.String(),
	})
	d.checkReversalBurst(ctx, byUserID)
}

// ObserveApproval records a successful approval event.
//
// This counter exists primarily to give the approval failure rate a meaningful
// denominator: alert rules can express "failure rate > N%" rather than relying
// on an absolute failure count, which would fire noisily on high-volume tenants.
func (d *AnomalyDetector) ObserveApproval(ctx context.Context, byUserID uuid.UUID) {
	if d == nil {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_approvals_total", metrics.Fields{
		"user_id": byUserID.String(),
	})
}

// ObserveApprovalFailure records a failed or rejected approval attempt and fires
// the APPROVAL_FAILURE roll-up anomaly.
//
// Two counters are incremented:
//   - finance_anomaly_approval_failures_total{user_id, tenant_id} — per-user
//     failure detail; used for targeted investigation.
//   - finance_anomaly_events_total{kind=APPROVAL_FAILURE, tenant_id} — roll-up;
//     used by the shared anomaly alert rule alongside all other anomaly kinds.
//
// Repeated failures from the same user suggest either an incorrectly configured
// approval threshold, an account with insufficient signing authority, or
// deliberate probing of the approval chain.
func (d *AnomalyDetector) ObserveApprovalFailure(ctx context.Context, byUserID uuid.UUID) {
	if d == nil {
		return
	}
	tenantID, _ := shared.GetTenantID(ctx)

	d.metrics.IncrementCounter("finance_anomaly_approval_failures_total", metrics.Fields{
		"user_id":   byUserID.String(),
		"tenant_id": tenantID.String(),
	})
	d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
		"kind":      "APPROVAL_FAILURE",
		"tenant_id": tenantID.String(),
	})
}

// ObserveReconciliationUnmatch records a line-level unmatch during reconciliation.
//
// A single unmatch is normal; a sustained spike indicates systematic data quality
// issues — misrouted postings, duplicate imports, or a bug in the matching logic.
// The counter is tenant-scoped because unmatch rates are meaningful relative to a
// tenant's own transaction volume, not across tenants.
func (d *AnomalyDetector) ObserveReconciliationUnmatch(ctx context.Context) {
	if d == nil {
		return
	}
	tenantID, _ := shared.GetTenantID(ctx)
	d.metrics.IncrementCounter("finance_anomaly_reconciliation_unmatches_total", metrics.Fields{
		"tenant_id": tenantID.String(),
	})
}

// ObserveRejection records a transaction rejection within the approval chain.
//
// Unlike ObserveApprovalFailure (which captures system-level or permission
// failures), this counter represents an explicit human decision to reject a
// submitted transaction. It is labelled by the rejecting user so that
// approval-chain load and patterns are visible per reviewer.
func (d *AnomalyDetector) ObserveRejection(ctx context.Context, byUserID uuid.UUID) {
	if d == nil {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_rejections_total", metrics.Fields{
		"user_id": byUserID.String(),
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── Private check helpers — original ────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════
//
// Contract shared by all helpers:
//   - Return immediately when the backing field is nil / zero (option not applied).
//   - Emit finance_anomaly_events_total{kind=KIND} on the roll-up counter.
//   - Emit a structured WarnContext log for log-pipeline traceability.
//   - Never return an error or block the caller.

// checkPostingVelocity fires HIGH_VELOCITY when the per-tenant posting rate
// exceeds postingVelocity.max within the configured rolling window.
func (d *AnomalyDetector) checkPostingVelocity(ctx context.Context, tenantID uuid.UUID) {
	if d.postingVelocity == nil {
		return
	}
	count, exceeded := d.postingVelocity.record(tenantID.String(), d.now())
	if !exceeded {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
		"kind":      "HIGH_VELOCITY",
		"tenant_id": tenantID.String(),
	})
	logger.WarnContext(ctx, "finance anomaly: high posting velocity for tenant", logger.Fields{
		"postings_in_window": count,
		"window":             d.postingVelocity.window.String(),
		"threshold":          d.postingVelocity.max,
		"tenant_id":          tenantID.String(),
	})
}

// checkReversalBurst fires REVERSAL_BURST when a single user's reversal rate
// exceeds reversalVelocity.max within the configured rolling window.
func (d *AnomalyDetector) checkReversalBurst(ctx context.Context, byUserID uuid.UUID) {
	if d.reversalVelocity == nil {
		return
	}
	tenantID, _ := shared.GetTenantID(ctx)
	count, exceeded := d.reversalVelocity.record(byUserID.String(), d.now())
	if !exceeded {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
		"kind":      "REVERSAL_BURST",
		"tenant_id": tenantID.String(),
	})
	logger.WarnContext(ctx, "finance anomaly: reversal burst detected for user", logger.Fields{
		"reversals_in_window": count,
		"window":              d.reversalVelocity.window.String(),
		"threshold":           d.reversalVelocity.max,
		"user_id":             byUserID.String(),
	})
}

// checkRoundAmount fires ROUND_AMOUNT when the debit amount is a positive
// multiple of roundAmountQuantum.  Zero and negative amounts are excluded
// because they indicate corrections, not structured transactions.
func (d *AnomalyDetector) checkRoundAmount(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal) {
	if d.roundAmountQuantum.IsZero() {
		return
	}
	if amount.IsZero() || amount.IsNegative() {
		return
	}
	if !amount.Mod(d.roundAmountQuantum).IsZero() {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
		"kind":      "ROUND_AMOUNT",
		"tenant_id": tenantID.String(),
	})
	logger.WarnContext(ctx, "finance anomaly: suspiciously round transaction amount", logger.Fields{
		"amount":    amount.StringFixed(2),
		"quantum":   d.roundAmountQuantum.StringFixed(2),
		"tenant_id": tenantID.String(),
	})
}

// checkOffHours fires OFF_HOURS_POSTING when the current wall-clock hour
// (in businessTimezone) falls outside [businessHoursStart, businessHoursEnd).
// A nil businessTimezone means WithOffHoursDetection was never called; the
// check is skipped entirely.
func (d *AnomalyDetector) checkOffHours(ctx context.Context, tenantID uuid.UUID) {
	if d.businessTimezone == nil {
		return
	}
	now := d.now().In(d.businessTimezone)
	hour := now.Hour()
	if hour >= d.businessHoursStart && hour < d.businessHoursEnd {
		return // within business hours — no anomaly
	}
	d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
		"kind":      "OFF_HOURS_POSTING",
		"tenant_id": tenantID.String(),
	})
	logger.WarnContext(ctx, "finance anomaly: posting outside business hours", logger.Fields{
		"hour":      hour,
		"start":     d.businessHoursStart,
		"end":       d.businessHoursEnd,
		"timezone":  d.businessTimezone.String(),
		"tenant_id": tenantID.String(),
	})
}

// checkDuplicateAmount fires DUPLICATE_AMOUNT when the same debit amount
// appears more than duplicateAmounts.max times for one tenant within the
// configured window. The amount is keyed as a two-decimal string so that
// 1000.00 and 1000.000 are treated identically regardless of Decimal scale.
func (d *AnomalyDetector) checkDuplicateAmount(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal) {
	if d.duplicateAmounts == nil {
		return
	}
	amountKey := amount.StringFixed(2)
	count, exceeded := d.duplicateAmounts.record(tenantID.String(), amountKey, d.now())
	if !exceeded {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
		"kind":      "DUPLICATE_AMOUNT",
		"tenant_id": tenantID.String(),
	})
	logger.WarnContext(ctx, "finance anomaly: duplicate posting amount detected", logger.Fields{
		"amount":      amountKey,
		"occurrences": count,
		"window":      d.duplicateAmounts.window.String(),
		"threshold":   d.duplicateAmounts.max,
		"tenant_id":   tenantID.String(),
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── Private check helpers — NEW ──────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// checkStatisticalOutlier fires STATISTICAL_OUTLIER when the posting amount
// deviates more than the configured z-score threshold from the tenant's
// historical mean (Welford online algorithm).
//
// The z-score is computed BEFORE incorporating the new observation, so each
// posting is judged against the existing baseline — not one that has already
// absorbed the current value. This makes the check sensitive to genuine outliers
// rather than to gradual drift in the mean.
//
// Example: a tenant whose postings average KES 5 000 (σ = 2 000) would trigger
// at 3σ on any posting above ~11 000 or below ~0 (though negative postings
// are rare and caught by other guards).
func (d *AnomalyDetector) checkStatisticalOutlier(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal) {
	if d.outlierStats == nil {
		return
	}
	// Convert to float64 for arithmetic.  Precision loss above ~10^15 is not
	// a concern given that SafetyEnforcer hard-caps postings at 10 M.
	amountF, _ := amount.Float64()
	zScore, outlier := d.outlierStats.observe(tenantID.String(), amountF)
	if !outlier {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
		"kind":      "STATISTICAL_OUTLIER",
		"tenant_id": tenantID.String(),
	})
	logger.WarnContext(ctx, "finance anomaly: posting amount is a statistical outlier for this tenant", logger.Fields{
		"amount":    amount.StringFixed(2),
		"z_score":   math.Round(zScore*100) / 100, // round to 2 d.p. for log readability
		"threshold": d.outlierStats.threshold,
		"tenant_id": tenantID.String(),
	})
}

// checkVelocityAcceleration fires VELOCITY_ACCELERATION when the short-window
// posting rate exceeds minRatio × the long-window baseline rate.
//
// This catches the "slow start → sudden spike" pattern: a tenant that normally
// posts at a low cadence and suddenly floods the service within a short burst —
// a pattern characteristic of automated injection attacks, runaway batch jobs,
// or replay attacks — even when the absolute count in either window has not yet
// crossed the HIGH_VELOCITY threshold.
func (d *AnomalyDetector) checkVelocityAcceleration(ctx context.Context, tenantID uuid.UUID) {
	if d.velocityAccel == nil {
		return
	}
	ratio, exceeded := d.velocityAccel.record(tenantID.String(), d.now())
	if !exceeded {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
		"kind":      "VELOCITY_ACCELERATION",
		"tenant_id": tenantID.String(),
	})
	logger.WarnContext(ctx, "finance anomaly: sudden acceleration in posting rate detected", logger.Fields{
		"short_window": d.velocityAccel.shortWindow.String(),
		"long_window":  d.velocityAccel.longWindow.String(),
		"rate_ratio":   math.Round(ratio*100) / 100,
		"min_ratio":    d.velocityAccel.minRatio,
		"tenant_id":    tenantID.String(),
	})
}

// checkCrossTenantSpread fires CROSS_TENANT_SPREAD when the posting's user has
// touched more than crossTenantSpread.max distinct tenants within the configured
// rolling window.
//
// A user legitimately configured for multiple tenants will touch a stable, small
// set. An expanding set — especially one that grows rapidly — is a strong signal
// of either credential compromise (attacker enumerating tenants) or a system
// account misconfigured to write to every tenant it can reach.
//
// The user ID is extracted from the request context.  A missing user ID produces
// the zero UUID as the actor key; a zero-UUID actor posting to multiple tenants
// is itself a noteworthy signal (unauthenticated postings should not reach this
// code path).
func (d *AnomalyDetector) checkCrossTenantSpread(ctx context.Context, tenantID uuid.UUID) {
	if d.crossTenantSpread == nil {
		return
	}
	// shared.GetUserID returns the zero UUID and false when no user is in context.
	userID, _ := shared.GetUserID(ctx)
	distinctCount, exceeded := d.crossTenantSpread.record(userID.String(), tenantID.String(), d.now())
	if !exceeded {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
		"kind":      "CROSS_TENANT_SPREAD",
		"tenant_id": tenantID.String(),
	})
	logger.WarnContext(ctx, "finance anomaly: user posting across an unusual number of tenants", logger.Fields{
		"user_id":          userID.String(),
		"distinct_tenants": distinctCount,
		"window":           d.crossTenantSpread.window.String(),
		"threshold":        d.crossTenantSpread.max,
		"tenant_id":        tenantID.String(),
	})
}

// checkAmountSequence fires AMOUNT_SEQUENCE when the tenant's recent posting
// amounts form an arithmetic progression of at least minLength consecutive
// equal-delta pairs.
//
// A progression implies intent: random legitimate postings do not step by a
// constant delta. This pattern is a hallmark of structured layering where the
// attacker or a misconfigured automation increases each amount by a fixed step
// to probe approval thresholds or stay just below a reporting limit.
//
// Both increasing (9 000 → 9 500 → 10 000) and decreasing (10 000 → 9 500 →
// 9 000) sequences are detected because either direction can represent probing.
// A zero delta between consecutive amounts is excluded (see DUPLICATE_AMOUNT).
func (d *AnomalyDetector) checkAmountSequence(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal) {
	if d.amountSequence == nil {
		return
	}
	runLength, detected := d.amountSequence.record(tenantID.String(), amount, d.now())
	if !detected {
		return
	}
	d.metrics.IncrementCounter("finance_anomaly_events_total", metrics.Fields{
		"kind":      "AMOUNT_SEQUENCE",
		"tenant_id": tenantID.String(),
	})
	logger.WarnContext(ctx, "finance anomaly: arithmetic progression detected in posting amounts — possible structuring", logger.Fields{
		"amount":         amount.StringFixed(2),
		"run_length":     runLength,
		"min_run_length": d.amountSequence.minLength,
		"window":         d.amountSequence.window.String(),
		"tenant_id":      tenantID.String(),
	})
}
