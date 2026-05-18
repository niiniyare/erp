package blocks

import (
	"awo.so/internal/web/ast"
	"awo.so/internal/web/ui"
)

// TrendDirection controls how the trend arrow is coloured.
type TrendDirection string

const (
	TrendUp      TrendDirection = "up_is_good"
	TrendDown    TrendDirection = "down_is_good"
	TrendNeutral TrendDirection = ""
)

// StatCardConfig configures a single KPI stat card.
type StatCardConfig struct {
	Label     string
	ValueKey  string
	Format    string         // "number"|"currency"|"percent"
	Currency  string
	TrendKey  string
	Trend     TrendDirection
	IconClass string
}

// StatCardBlock renders one KPI metric with label, value, and optional trend.
func StatCardBlock(sess ui.UISessionContext, cfg StatCardConfig) ast.Node {
	currency := cfg.Currency
	if currency == "" && cfg.Format == "currency" {
		currency = sess.Currency
	}
	return ast.StatNode{
		Label:     cfg.Label,
		ValueKey:  cfg.ValueKey,
		Format:    cfg.Format,
		Currency:  currency,
		TrendKey:  cfg.TrendKey,
		TrendMode: string(cfg.Trend),
		IconClass: cfg.IconClass,
	}
}
