package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// // Service represents a service that interacts with a database.
// type Service interface {
// 	// Health returns a map of health status information.
// 	// The keys and values in the map are service-specific.
// 	Health() map[string]string

// 	// Close terminates the database connection.
// 	// It returns an error if the connection cannot be closed.
// 	Close() error
// }

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	schema     = os.Getenv("BLUEPRINT_DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(fmt.Sprintf("db down: %v", err)) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}
func (s *service) Create(ctx context.Context, table string, data map[string]interface{}) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to insert")
	}
	columns := ""
	values := ""
	args := []interface{}{}
	i := 1
	for k, v := range data {
		if columns != "" {
			columns += ", "
			values += ", "
		}
		columns += k
		values += fmt.Sprintf("$%d", i)
		args = append(args, v)
		i++
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", table, columns, values)
	var id interface{}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func (s *service) Read(ctx context.Context, table string, filter map[string]interface{}) ([]map[string]interface{}, error) {
	where := ""
	args := []interface{}{}
	i := 1
	for k, v := range filter {
		if where == "" {
			where = "WHERE "
		} else {
			where += " AND "
		}
		where += fmt.Sprintf("%s = $%d", k, i)
		args = append(args, v)
		i++
	}
	query := fmt.Sprintf("SELECT * FROM %s %s", table, where)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}
		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			rowMap[colName] = *val
		}
		results = append(results, rowMap)
	}
	return results, rows.Err()
}

func (s *service) Update(ctx context.Context, table string, filter map[string]interface{}, update map[string]interface{}) (int64, error) {
	if len(update) == 0 {
		return 0, fmt.Errorf("no update data provided")
	}
	set := ""
	args := []interface{}{}
	i := 1
	for k, v := range update {
		if set != "" {
			set += ", "
		}
		set += fmt.Sprintf("%s = $%d", k, i)
		args = append(args, v)
		i++
	}
	where := ""
	for k, v := range filter {
		if where == "" {
			where = "WHERE "
		} else {
			where += " AND "
		}
		where += fmt.Sprintf("%s = $%d", k, i)
		args = append(args, v)
		i++
	}
	query := fmt.Sprintf("UPDATE %s SET %s %s", table, set, where)
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *service) Delete(ctx context.Context, table string, filter map[string]interface{}) (int64, error) {
	where := ""
	args := []interface{}{}
	i := 1
	for k, v := range filter {
		if where == "" {
			where = "WHERE "
		} else {
			where += " AND "
		}
		where += fmt.Sprintf("%s = $%d", k, i)
		args = append(args, v)
		i++
	}
	query := fmt.Sprintf("DELETE FROM %s %s", table, where)
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *service) Exec(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return res.RowsAffected()
}

func (s *service) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}
		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			rowMap[colName] = *val
		}
		results = append(results, rowMap)
	}
	return results, rows.Err()
}