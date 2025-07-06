package database

import "context"

// Health checks the database connection
func (d *database) HealthCheck(ctx context.Context) error {
	return d.db.PingContext(ctx)
}
