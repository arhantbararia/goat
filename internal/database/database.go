package database

import (
	"context"
)

// Database is a generic interface for database operations.
type Database interface {
	// Health returns a map of health status information.
	Health() map[string]string

	// Close terminates the database connection.
	Close() error

	// Create inserts a new record into the database.
	Create(ctx context.Context, table string, data map[string]interface{}) (interface{}, error)

	// Read retrieves records from the database based on a query/filter.
	Read(ctx context.Context, table string, filter map[string]interface{}) ([]map[string]interface{}, error)

	// Update modifies existing records in the database.
	Update(ctx context.Context, table string, filter map[string]interface{}, update map[string]interface{}) (int64, error)

	// Delete removes records from the database.
	Delete(ctx context.Context, table string, filter map[string]interface{}) (int64, error)

	// Exec executes a raw query or command (for SQL databases).
	Exec(ctx context.Context, query string, args ...interface{}) (interface{}, error)

	// Query executes a raw query and returns results (for SQL databases).
	Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
}
