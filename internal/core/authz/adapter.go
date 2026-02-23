package authz

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgxAdapter implements persist.BatchAdapter backed by PostgreSQL via pgx.
type pgxAdapter struct {
	pool *pgxpool.Pool
}

func newPgxAdapter(pool *pgxpool.Pool) *pgxAdapter {
	return &pgxAdapter{pool: pool}
}

// LoadPolicy loads all rules from casbin_rule into the model.
func (a *pgxAdapter) LoadPolicy(m model.Model) error {
	ctx := context.Background()
	rows, err := a.pool.Query(ctx,
		`SELECT ptype, v0, v1, v2, v3, v4, v5 FROM casbin_rule ORDER BY ptype`)
	if err != nil {
		return fmt.Errorf("authz adapter LoadPolicy: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ptype, v0, v1, v2, v3, v4, v5 string
		if err := rows.Scan(&ptype, &v0, &v1, &v2, &v3, &v4, &v5); err != nil {
			return fmt.Errorf("authz adapter LoadPolicy scan: %w", err)
		}
		rule := filterEmpty([]string{v0, v1, v2, v3, v4, v5})
		persist.LoadPolicyLine(fmt.Sprintf("%s, %s", ptype, joinRule(rule)), m)
	}
	return rows.Err()
}

// SavePolicy replaces all rows with the current model state (full rewrite).
func (a *pgxAdapter) SavePolicy(m model.Model) error {
	ctx := context.Background()
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("authz adapter SavePolicy begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `TRUNCATE casbin_rule`); err != nil {
		return fmt.Errorf("authz adapter SavePolicy truncate: %w", err)
	}

	for ptype, assertions := range m["p"] {
		for _, rule := range assertions.Policy {
			v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
			if _, err := tx.Exec(ctx,
				`INSERT INTO casbin_rule(ptype,v0,v1,v2,v3,v4,v5) VALUES($1,$2,$3,$4,$5,$6,$7)
				 ON CONFLICT DO NOTHING`,
				ptype, v0, v1, v2, v3, v4, v5,
			); err != nil {
				return fmt.Errorf("authz adapter SavePolicy insert p: %w", err)
			}
		}
	}
	for ptype, assertions := range m["g"] {
		for _, rule := range assertions.Policy {
			v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
			if _, err := tx.Exec(ctx,
				`INSERT INTO casbin_rule(ptype,v0,v1,v2,v3,v4,v5) VALUES($1,$2,$3,$4,$5,$6,$7)
				 ON CONFLICT DO NOTHING`,
				ptype, v0, v1, v2, v3, v4, v5,
			); err != nil {
				return fmt.Errorf("authz adapter SavePolicy insert g: %w", err)
			}
		}
	}
	return tx.Commit(ctx)
}

// AddPolicy inserts a single rule (ON CONFLICT DO NOTHING — idempotent).
func (a *pgxAdapter) AddPolicy(sec, ptype string, rule []string) error {
	v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
	_, err := a.pool.Exec(context.Background(),
		`INSERT INTO casbin_rule(ptype,v0,v1,v2,v3,v4,v5) VALUES($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT DO NOTHING`,
		ptype, v0, v1, v2, v3, v4, v5,
	)
	return err
}

// AddPolicies inserts multiple rules in a single transaction.
func (a *pgxAdapter) AddPolicies(sec, ptype string, rules [][]string) error {
	ctx := context.Background()
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, rule := range rules {
		v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
		if _, err := tx.Exec(ctx,
			`INSERT INTO casbin_rule(ptype,v0,v1,v2,v3,v4,v5) VALUES($1,$2,$3,$4,$5,$6,$7)
			 ON CONFLICT DO NOTHING`,
			ptype, v0, v1, v2, v3, v4, v5,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// RemovePolicy deletes a single rule by exact match.
func (a *pgxAdapter) RemovePolicy(sec, ptype string, rule []string) error {
	v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
	_, err := a.pool.Exec(context.Background(),
		`DELETE FROM casbin_rule WHERE ptype=$1 AND v0=$2 AND v1=$3 AND v2=$4 AND v3=$5 AND v4=$6 AND v5=$7`,
		ptype, v0, v1, v2, v3, v4, v5,
	)
	return err
}

// RemovePolicies deletes multiple rules in a single transaction.
func (a *pgxAdapter) RemovePolicies(sec, ptype string, rules [][]string) error {
	ctx := context.Background()
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, rule := range rules {
		v0, v1, v2, v3, v4, v5 := ruleToValues(rule)
		if _, err := tx.Exec(ctx,
			`DELETE FROM casbin_rule WHERE ptype=$1 AND v0=$2 AND v1=$3 AND v2=$4 AND v3=$5 AND v4=$6 AND v5=$7`,
			ptype, v0, v1, v2, v3, v4, v5,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// RemoveFilteredPolicy deletes rules matching a field-index prefix filter.
func (a *pgxAdapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	query := `DELETE FROM casbin_rule WHERE ptype=$1`
	args := []interface{}{ptype}
	cols := []string{"v0", "v1", "v2", "v3", "v4", "v5"}

	for i, val := range fieldValues {
		if val == "" {
			continue
		}
		col := cols[fieldIndex+i]
		args = append(args, val)
		query += fmt.Sprintf(" AND %s=$%d", col, len(args))
	}

	_, err := a.pool.Exec(context.Background(), query, args...)
	return err
}

// ruleToValues pads a rule slice to exactly 6 string slots.
func ruleToValues(rule []string) (v0, v1, v2, v3, v4, v5 string) {
	padded := make([]string, 6)
	copy(padded, rule)
	return padded[0], padded[1], padded[2], padded[3], padded[4], padded[5]
}

// filterEmpty removes trailing empty strings from a slice.
func filterEmpty(ss []string) []string {
	last := len(ss) - 1
	for last >= 0 && ss[last] == "" {
		last--
	}
	return ss[:last+1]
}

// joinRule joins rule values with ", " for persist.LoadPolicyLine.
func joinRule(rule []string) string {
	out := ""
	for i, v := range rule {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}
