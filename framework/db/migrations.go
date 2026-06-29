// Package db exposes the embedded migration files for use by the migrate runner.
package db

import "embed"

// Migrations is the embedded filesystem containing all framework SQL migrations.
//
//go:embed migrations
var Migrations embed.FS
