package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

type sqliteService struct {
	db *sql.DB
}

var (
	sqlitePath     = os.Getenv("BLUEPRINT_SQLITE_PATH")
	sqliteInstance *sqliteService
)

func NewSQLite() Service {
	if sqliteInstance != nil {
		return sqliteInstance
	}
	if sqlitePath == "" {
		log.Fatal("BLUEPRINT_SQLITE_PATH environment variable not set")
	}
	db, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		log.Fatal(err)
	}
	sqliteInstance = &sqliteService{
		db: db,
	}
	return sqliteInstance
}

func (s *sqliteService) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(fmt.Sprintf("db down: %v", err))
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	return stats
}

func (s *sqliteService) Close() error {
	log.Printf("Disconnected from SQLite database: %s", sqlitePath)
	return s.db.Close()
}


func (s *sqliteService) Create(ctx context.Context, table string, data map[string]interface{}) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to insert")
	}
	columns := ""
	values := ""
	args := []interface{}{}
	i := 0
	for k, v := range data {
		if i > 0 {
			columns += ", "
			values += ", "
		}
		columns += k
		values += "?"
		args = append(args, v)
		i++
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, columns, values)
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return id, nil
}

func (s *sqliteService) Read(ctx context.Context, table string, filter map[string]interface{}) ([]map[string]interface{}, error) {
	where := ""
	args := []interface{}{}
	i := 0
	for k, v := range filter {
		if i == 0 {
			where += "WHERE "
		} else {
			where += " AND "
		}
		where += fmt.Sprintf("%s = ?", k)
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
		row := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			row[colName] = *val
		}
		results = append(results, row)
	}
	return results, nil
}

func (s *sqliteService) Update(ctx context.Context, table string, filter map[string]interface{}, update map[string]interface{}) (int64, error) {
	if len(update) == 0 {
		return 0, fmt.Errorf("no update data provided")
	}
	set := ""
	args := []interface{}{}
	i := 0
	for k, v := range update {
		if i > 0 {
			set += ", "
		}
		set += fmt.Sprintf("%s = ?", k)
		args = append(args, v)
		i++
	}
	where := ""
	j := 0
	for k, v := range filter {
		if j == 0 {
			where += "WHERE "
		} else {
			where += " AND "
		}
		where += fmt.Sprintf("%s = ?", k)
		args = append(args, v)
		j++
	}
	query := fmt.Sprintf("UPDATE %s SET %s %s", table, set, where)
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (s *sqliteService) Delete(ctx context.Context, table string, filter map[string]interface{}) (int64, error) {
	where := ""
	args := []interface{}{}
	i := 0
	for k, v := range filter {
		if i == 0 {
			where += "WHERE "
		} else {
			where += " AND "
		}
		where += fmt.Sprintf("%s = ?", k)
		args = append(args, v)
		i++
	}
	query := fmt.Sprintf("DELETE FROM %s %s", table, where)
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (s *sqliteService) Exec(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *sqliteService) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
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
		row := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			row[colName] = *val
		}
		results = append(results, row)
	}
	return results, nil
}