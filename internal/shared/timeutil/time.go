// Package timeutil package provides comprehensive tools for parsing and formatting time, including support for multiple parsing strategies (Unix timestamps, float timestamps, common formats) and robust error handling. It also offers powerful time calculation features such as age calculation from birthdate, determining days and business days between dates, and quarter start/end calculations. Business logic functions enable weekend/business day detection, adding or subtracting business days, and finding the next or previous business day.
//
// Additionally, timeutil excels in timezone operations with enhanced conversion capabilities and functions to get the current time, today, tomorrow, or yesterday in any timezone. The package includes utilities for human-readable duration formatting and parsing, as well as calendar functions to determine the start or end of a week, month, quarter, or year, detect leap years, and calculate days in a month. With convenience features like the TimeRange struct, truncate and round functions, and Unix timestamp conversions, timeutil is designed to handle a wide range of time-related tasks in business applications.
package timeutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Common time formats
const (
	// ISO and RFC Standards
	ISO8601         = "2006-01-02T15:04:05Z07:00"
	ISO8601Basic    = "20060102T150405Z"
	ISO8601Date     = "2006-01-02"
	ISO8601Time     = "15:04:05"
	ISO8601DateTime = "2006-01-02T15:04:05"
	ISO8601Nano     = "2006-01-02T15:04:05.000000000Z07:00"
	ISO8601Milli    = "2006-01-02T15:04:05.000Z07:00"
	ISO8601Micro    = "2006-01-02T15:04:05.000000Z07:00"
	RFC822Short     = "02 Jan 06 15:04 MST"
	RFC850Long      = "Monday, 02-Jan-06 15:04:05 MST"

	// Date Formats
	DateTime        = "2006-01-02 15:04:05"
	DateTimeLocal   = "2006-01-02 15:04:05 MST"
	DateTimeMilli   = "2006-01-02 15:04:05.000"
	DateTimeMicro   = "2006-01-02 15:04:05.000000"
	DateTimeNano    = "2006-01-02 15:04:05.000000000"
	DateTimeCompact = "20060102150405"
	DateCompact     = "20060102"

	// Regional Date Formats
	USDate          = "01/02/2006"
	USDateTime      = "01/02/2006 15:04:05"
	USDateTime12    = "01/02/2006 3:04:05 PM"
	USDateShort     = "1/2/2006"
	USDateTimeShort = "1/2/2006 3:04 PM"

	EUDate          = "02/01/2006"
	EUDateTime      = "02/01/2006 15:04:05"
	EUDateTime12    = "02/01/2006 3:04:05 PM"
	EUDateShort     = "2/1/2006"
	EUDateTimeShort = "2/1/2006 3:04 PM"

	UKDate     = "02/01/2006"
	UKDateTime = "02/01/2006 15:04:05"

	GermanDate      = "02.01.2006"
	GermanDateTime  = "02.01.2006 15:04:05"
	GermanDateShort = "2.1.2006"

	// Human Readable Formats
	HumanDate          = "January 2, 2006"
	HumanDateTime      = "January 2, 2006 at 3:04 PM"
	HumanDateTimeLong  = "Monday, January 2, 2006 at 3:04:05 PM"
	HumanDateShort     = "Jan 2, 2006"
	HumanDateTimeShort = "Jan 2, 2006 3:04 PM"
	HumanWeekday       = "Monday, January 2"
	HumanWeekdayShort  = "Mon, Jan 2"

	// Time Only Formats
	TimeOnly        = "15:04:05"
	TimeShort       = "15:04"
	Time12Hour      = "3:04 PM"
	Time12HourSec   = "3:04:05 PM"
	Time12HourMilli = "3:04:05.000 PM"
	TimeWithZone    = "15:04:05 MST"
	Time12WithZone  = "3:04:05 PM MST"
	TimePrecise     = "15:04:05.000000000"

	// Log and System Formats
	LogFormat      = "2006-01-02T15:04:05.000Z07:00"
	LogFormatLocal = "2006-01-02 15:04:05.000"
	LogFormatShort = "01-02 15:04:05"
	SyslogFormat   = "Jan _2 15:04:05"
	ApacheLog      = "02/Jan/2006:15:04:05 -0700"
	NginxLog       = "02/Jan/2006:15:04:05 +0000"

	// Database Formats
	MySQLDateTime  = "2006-01-02 15:04:05"
	MySQLDate      = "2006-01-02"
	MySQLTime      = "15:04:05"
	PostgreSQLTime = "2006-01-02 15:04:05.000000-07:00"
	SQLiteTime     = "2006-01-02 15:04:05"

	// Web and API Formats
	HTTPDate   = "Mon, 02 Jan 2006 15:04:05 GMT"
	CookieDate = "Mon, 02-Jan-2006 15:04:05 MST"
	RSSDate    = "Mon, 02 Jan 2006 15:04:05 -0700"
	ATOMDate   = "2006-01-02T15:04:05Z"
	JSONDate   = "2006-01-02T15:04:05.000Z"

	// Filename Safe Formats
	FilenameDate      = "2006-01-02"
	FilenameDateTime  = "2006-01-02_15-04-05"
	FilenameTimestamp = "20060102_150405"
	FilenameCompact   = "20060102150405"

	// Financial and Business Formats
	FinancialDate = "02-Jan-2006"
	BusinessDate  = "2006/01/02"
	QuarterFormat = "Q1 2006"
	FiscalQuarter = "FY2006 Q1"

	// Asian Formats
	JapaneseDate     = "2006年01月02日"
	JapaneseDateTime = "2006年01月02日 15時04分05秒"
	ChineseDate      = "2006年1月2日"
	ChineseDateTime  = "2006年1月2日 15时04分05秒"
	KoreanDate       = "2006년 1월 2일"
	KoreanDateTime   = "2006년 1월 2일 15시 04분 05초"

	// Sortable Formats
	SortableDate      = "2006-01-02"
	SortableDateTime  = "2006-01-02-15-04-05"
	SortableTimestamp = "2006.01.02.15.04.05"

	// Military and Aviation
	MilitaryTime     = "1504"
	MilitaryDateTime = "021504Z JAN 06"
	ZuluTime         = "150405Z"

	// Custom Application Formats
	CompactDate     = "060102"
	CompactTime     = "1504"
	CompactDateTime = "0601021504"

	// Email Formats
	EmailDate = "Mon, 2 Jan 2006 15:04:05 -0700"
	IMAPDate  = "2-Jan-2006 15:04:05 -0700"

	// Backup and Archive Formats
	BackupFormat  = "2006-01-02_15-04-05"
	ArchiveFormat = "2006_01_02_150405"

	// Version Control Formats
	GitDate = "Mon Jan 2 15:04:05 2006 -0700"
	SVNDate = "2006-01-02 15:04:05 -0700"

	// Mobile and Short Display
	MobileDate     = "1/2/06"
	MobileTime     = "3:04"
	MobileDateTime = "1/2/06 3:04"

	// Podcast and Media
	PodcastDate    = "Mon, 02 Jan 2006 15:04:05 GMT"
	MediaTimestamp = "15:04:05.000"

	// Scientific and Precise
	ScientificTime     = "2006-01-02T15:04:05.000000000Z"
	UnixTimestamp      = "1136239445"
	UnixMilliTimestamp = "1136239445000"
)

// Duration constants for convenience
const (
	Nanosecond  = time.Nanosecond
	Microsecond = time.Microsecond
	Millisecond = time.Millisecond
	Second      = time.Second
	Minute      = time.Minute
	Hour        = time.Hour
	Day         = 24 * time.Hour
	Week        = 7 * Day
)

// TimeRange represents a time period
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Duration returns the duration of the time range
func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// Contains checks if a time falls within the range
func (tr TimeRange) Contains(t time.Time) bool {
	return !t.Before(tr.Start) && !t.After(tr.End)
}

// Overlaps checks if this range overlaps with another
func (tr TimeRange) Overlaps(other TimeRange) bool {
	return tr.Start.Before(other.End) && other.Start.Before(tr.End)
}

// FormatPresets provides common formatting presets for quick access
var FormatPresets = map[string]string{
	// Standard presets
	"default":  ISO8601,
	"iso":      ISO8601,
	"rfc3339":  time.RFC3339,
	"datetime": DateTime,
	"date":     ISO8601Date,
	"time":     TimeOnly,

	// Human readable presets
	"human":       HumanDate,
	"human-full":  HumanDateTime,
	"human-long":  HumanDateTimeLong,
	"human-short": HumanDateShort,

	// Regional presets
	"us":          USDate,
	"us-full":     USDateTime,
	"eu":          EUDate,
	"eu-full":     EUDateTime,
	"uk":          UKDate,
	"uk-full":     UKDateTime,
	"german":      GermanDate,
	"german-full": GermanDateTime,

	// System presets
	"log":       LogFormat,
	"log-short": LogFormatShort,
	"mysql":     MySQLDateTime,
	"postgres":  PostgreSQLTime,
	"sqlite":    SQLiteTime,

	// Web presets
	"http":   HTTPDate,
	"json":   JSONDate,
	"rss":    RSSDate,
	"cookie": CookieDate,

	// File presets
	"filename": FilenameDateTime,
	"backup":   BackupFormat,
	"compact":  DateTimeCompact,

	// Business presets
	"business":  BusinessDate,
	"financial": FinancialDate,
	"sortable":  SortableDateTime,

	// Mobile presets
	"mobile":      MobileDateTime,
	"mobile-date": MobileDate,
	"mobile-time": MobileTime,

	// Asian presets
	"japanese": JapaneseDate,
	"chinese":  ChineseDate,
	"korean":   KoreanDate,

	// Precise presets
	"precise":    ISO8601Nano,
	"scientific": ScientificTime,
	"unix":       UnixTimestamp,
}

// GetFormatPreset returns a format layout by preset name
func GetFormatPreset(preset string) (string, bool) {
	format, exists := FormatPresets[preset]
	return format, exists
}

// ListFormatPresets returns all available format presets
func ListFormatPresets() []string {
	presets := make([]string, 0, len(FormatPresets))
	for preset := range FormatPresets {
		presets = append(presets, preset)
	}
	return presets
}

// FormatTimestamp is a Parsing & Formatting with  better error handling and more flexible parsing.
func FormatTimestamp(ts, inputLayout, outputLayout, location string) (string, error) {
	if ts == "" {
		return "", errors.New("empty timestamp")
	}

	// Load timezone
	loc, err := parseLocation(location)
	if err != nil {
		return "", err
	}

	// Parse timestamp
	t, err := parseTimestamp(ts, inputLayout, loc)
	if err != nil {
		return "", err
	}

	// Adjust timezone
	t = t.In(loc)

	// Default output format
	if outputLayout == "" {
		outputLayout = time.RFC3339
	}

	return t.Format(outputLayout), nil
}

// parseLocation helper function to handle timezone parsing
func parseLocation(location string) (*time.Location, error) {
	switch location {
	case "", "Local":
		return time.Local, nil
	case "UTC":
		return time.UTC, nil
	default:
		loc, err := time.LoadLocation(location)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone '%s': %w", location, err)
		}
		return loc, nil
	}
}

// parseTimestamp helper function with multiple parsing strategies
func parseTimestamp(ts, inputLayout string, loc *time.Location) (time.Time, error) {
	// Strategy 1: Try Unix timestamp (seconds or milliseconds)
	if unix, err := strconv.ParseInt(ts, 10, 64); err == nil {
		if len(ts) >= 13 { // Milliseconds
			return time.UnixMilli(unix), nil
		} else if len(ts) >= 10 { // Seconds
			return time.Unix(unix, 0), nil
		}
	}

	// Strategy 2: Try Unix timestamp as float (with fractional seconds)
	if unixFloat, err := strconv.ParseFloat(ts, 64); err == nil {
		sec := int64(unixFloat)
		nsec := int64((unixFloat - float64(sec)) * 1e9)
		return time.Unix(sec, nsec), nil
	}

	// Strategy 3: Use provided layout
	if inputLayout != "" {
		return time.ParseInLocation(inputLayout, ts, loc)
	}

	// Strategy 4: Try common formats in order of likelihood
	layouts := []string{
		// Standard formats first
		time.RFC3339,
		time.RFC3339Nano,
		ISO8601,
		ISO8601DateTime,
		ISO8601Nano,
		ISO8601Milli,
		ISO8601Micro,

		// Common database and system formats
		DateTime,
		DateTimeLocal,
		DateTimeMilli,
		DateTimeMicro,
		DateTimeNano,
		MySQLDateTime,
		PostgreSQLTime,

		// Web and API formats
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		HTTPDate,
		JSONDate,
		LogFormat,

		// Date only formats
		ISO8601Date,
		MySQLDate,
		USDate,
		EUDate,
		GermanDate,
		BusinessDate,

		// Date with time variations
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		USDateTime,
		USDateTime12,
		EUDateTime,
		EUDateTime12,
		GermanDateTime,

		// Human readable
		HumanDate,
		HumanDateTime,
		HumanDateShort,
		HumanDateTimeShort,

		// Compact formats
		DateTimeCompact,
		DateCompact,
		FilenameDateTime,
		FilenameTimestamp,

		// Log formats
		ApacheLog,
		NginxLog,
		SyslogFormat,

		// Email formats
		EmailDate,
		IMAPDate,

		// Regional variations
		"01/02/2006 15:04:05",
		"02/01/2006 15:04:05",
		"2006/01/02 15:04:05",
		"02.01.2006 15:04:05",
		"2.1.2006 15:04:05",
		"1/2/2006 15:04:05",

		// Time only formats (for completion)
		TimeOnly,
		TimeShort,
		Time12Hour,
		Time12HourSec,
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, ts, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp '%s'", ts)
}

// Now returns current time in specified timezone
func Now(location string) (time.Time, error) {
	loc, err := parseLocation(location)
	if err != nil {
		return time.Time{}, err
	}
	return time.Now().In(loc), nil
}

// Today returns start of today in specified timezone
func Today(location string) (time.Time, error) {
	now, err := Now(location)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
}

// Tomorrow returns start of tomorrow in specified timezone
func Tomorrow(location string) (time.Time, error) {
	today, err := Today(location)
	if err != nil {
		return time.Time{}, err
	}
	return today.AddDate(0, 0, 1), nil
}

// Yesterday returns start of yesterday in specified timezone
func Yesterday(location string) (time.Time, error) {
	today, err := Today(location)
	if err != nil {
		return time.Time{}, err
	}
	return today.AddDate(0, 0, -1), nil
}

// StartOfWeek returns start of week (Monday) in specified timezone
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	return t.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
}

// EndOfWeek returns end of week (Sunday) in specified timezone
func EndOfWeek(t time.Time) time.Time {
	return StartOfWeek(t).AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond)
}

// StartOfMonth returns start of month in specified timezone
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns end of month in specified timezone
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond)
}

// StartOfYear returns start of year in specified timezone
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns end of year in specified timezone
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// ToUnix converts time to Unix timestamp (seconds)
func ToUnix(t time.Time) int64 {
	return t.Unix()
}

// ToUnixMilli converts time to Unix timestamp (milliseconds)
func ToUnixMilli(t time.Time) int64 {
	return t.UnixMilli()
}

// ToUnixNano converts time to Unix timestamp (nanoseconds)
func ToUnixNano(t time.Time) int64 {
	return t.UnixNano()
}

// FromUnix creates time from Unix timestamp (seconds)
func FromUnix(sec int64) time.Time {
	return time.Unix(sec, 0)
}

// FromUnixMilli creates time from Unix timestamp (milliseconds)
func FromUnixMilli(ms int64) time.Time {
	return time.UnixMilli(ms)
}

// FromUnixNano creates time from Unix timestamp (nanoseconds)
func FromUnixNano(ns int64) time.Time {
	return time.Unix(0, ns)
}

// Age calculates age based on birthdate
func Age(birthdate time.Time) int {
	now := time.Now()
	age := now.Year() - birthdate.Year()
	if now.YearDay() < birthdate.YearDay() {
		age--
	}
	return age
}

// DaysBetween calculates days between two dates
func DaysBetween(start, end time.Time) int {
	if end.Before(start) {
		start, end = end, start
	}
	return int(end.Sub(start).Hours() / 24)
}

// BusinessDaysBetween calculates business days (Monday-Friday) between two dates
func BusinessDaysBetween(start, end time.Time) int {
	if end.Before(start) {
		start, end = end, start
	}

	days := 0
	current := start
	for current.Before(end) || current.Equal(end) {
		weekday := current.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			days++
		}
		current = current.AddDate(0, 0, 1)
	}
	return days
}

// IsWeekend checks if the given time is on weekend
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsBusinessDay checks if the given time is on a business day
func IsBusinessDay(t time.Time) bool {
	return !IsWeekend(t)
}

// NextBusinessDay returns the next business day
func NextBusinessDay(t time.Time) time.Time {
	next := t.AddDate(0, 0, 1)
	for IsWeekend(next) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// PreviousBusinessDay returns the previous business day
func PreviousBusinessDay(t time.Time) time.Time {
	prev := t.AddDate(0, 0, -1)
	for IsWeekend(prev) {
		prev = prev.AddDate(0, 0, -1)
	}
	return prev
}

// FormatDuration formats duration in human readable format
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	parts := []string{}

	days := int(d.Hours()) / 24
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
		d -= time.Duration(days) * 24 * time.Hour
	}

	hours := int(d.Hours())
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
		d -= time.Duration(hours) * time.Hour
	}

	minutes := int(d.Minutes())
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
		d -= time.Duration(minutes) * time.Minute
	}

	seconds := int(d.Seconds())
	if seconds > 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	if len(parts) == 0 {
		return "0s"
	}

	return strings.Join(parts, "")
}

// ParseDuration parses duration string with support for days and weeks
func ParseDuration(s string) (time.Duration, error) {
	// Handle days and weeks
	if strings.Contains(s, "d") || strings.Contains(s, "w") {
		var total time.Duration
		remaining := s

		// Parse weeks
		if strings.Contains(remaining, "w") {
			parts := strings.Split(remaining, "w")
			if len(parts) == 2 {
				weeks, err := strconv.Atoi(parts[0])
				if err != nil {
					return 0, fmt.Errorf("invalid weeks: %w", err)
				}
				total += time.Duration(weeks) * Week
				remaining = parts[1]
			}
		}

		// Parse days
		if strings.Contains(remaining, "d") {
			parts := strings.Split(remaining, "d")
			if len(parts) == 2 {
				days, err := strconv.Atoi(parts[0])
				if err != nil {
					return 0, fmt.Errorf("invalid days: %w", err)
				}
				total += time.Duration(days) * Day
				remaining = parts[1]
			}
		}

		// Parse remaining standard duration
		if remaining != "" {
			rd, err := time.ParseDuration(remaining)
			if err != nil {
				return 0, fmt.Errorf("invalid duration '%s': %w", remaining, err)
			}
			total += rd
		}

		return total, nil
	}

	return time.ParseDuration(s)
}

// TimeUntil returns duration until target time
func TimeUntil(target time.Time) time.Duration {
	return time.Until(target)
}

// TimeSince returns duration since given time
func TimeSince(t time.Time) time.Duration {
	return time.Since(t)
}

// ConvertTimezone converts time from one timezone to another
func ConvertTimezone(t time.Time, fromTz, toTz string) (time.Time, error) {
	// If fromTz is provided, first interpret the time in that timezone
	if fromTz != "" {
		fromLoc, err := parseLocation(fromTz)
		if err != nil {
			return time.Time{}, err
		}
		// If the time doesn't have location info, set it
		if t.Location() == time.UTC && fromTz != "UTC" {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), fromLoc)
		}
	}

	// Convert to target timezone
	toLoc, err := parseLocation(toTz)
	if err != nil {
		return time.Time{}, err
	}

	return t.In(toLoc), nil
}

// IsLeapYear checks if the given year is a leap year
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth returns number of days in the given month/year
func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// QuarterStart returns the start of quarter for given time
func QuarterStart(t time.Time) time.Time {
	month := t.Month()
	var quarterMonth time.Month
	switch {
	case month >= 1 && month <= 3:
		quarterMonth = 1
	case month >= 4 && month <= 6:
		quarterMonth = 4
	case month >= 7 && month <= 9:
		quarterMonth = 7
	default:
		quarterMonth = 10
	}
	return time.Date(t.Year(), quarterMonth, 1, 0, 0, 0, 0, t.Location())
}

// QuarterEnd returns the end of quarter for given time
func QuarterEnd(t time.Time) time.Time {
	start := QuarterStart(t)
	return start.AddDate(0, 3, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond)
}

// GetQuarter returns the quarter (1-4) for given time
func GetQuarter(t time.Time) int {
	month := t.Month()
	return int((month-1)/3) + 1
}

// TimeAgo returns human-readable time difference (e.g., "2 hours ago")
func TimeAgo(t time.Time) string {
	diff := time.Since(t)

	if diff < time.Minute {
		return "just now"
	}

	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	if diff < 7*24*time.Hour {
		days := int(diff.Hours()) / 24
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	if diff < 30*24*time.Hour {
		weeks := int(diff.Hours()) / (24 * 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	if diff < 365*24*time.Hour {
		months := int(diff.Hours()) / (24 * 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	years := int(diff.Hours()) / (24 * 365)
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// Sleep pauses execution for the given duration
func Sleep(d time.Duration) {
	time.Sleep(d)
}

// CreateTimeRange creates a new TimeRange
func CreateTimeRange(start, end time.Time) TimeRange {
	return TimeRange{Start: start, End: end}
}

// Truncate truncates time to specified precision
func Truncate(t time.Time, precision time.Duration) time.Time {
	return t.Truncate(precision)
}

// Round rounds time to specified precision
func Round(t time.Time, precision time.Duration) time.Time {
	return t.Round(precision)
}

// AddBusinessDays adds business days to a date (skipping weekends)
func AddBusinessDays(t time.Time, days int) time.Time {
	current := t
	remaining := days

	for remaining > 0 {
		current = current.AddDate(0, 0, 1)
		if IsBusinessDay(current) {
			remaining--
		}
	}

	return current
}

// SubBusinessDays subtracts business days from a date (skipping weekends)
func SubBusinessDays(t time.Time, days int) time.Time {
	current := t
	remaining := days

	for remaining > 0 {
		current = current.AddDate(0, 0, -1)
		if IsBusinessDay(current) {
			remaining--
		}
	}

	return current
}
