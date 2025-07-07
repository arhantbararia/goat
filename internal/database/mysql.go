package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

type mysqlService struct {
	db *sql.DB
}

var (
	mysqlDatabase = os.Getenv("BLUEPRINT_MYSQL_DATABASE")
	mysqlPassword = os.Getenv("BLUEPRINT_MYSQL_PASSWORD")
	mysqlUsername = os.Getenv("BLUEPRINT_MYSQL_USERNAME")
	mysqlPort     = os.Getenv("BLUEPRINT_MYSQL_PORT")
	mysqlHost     = os.Getenv("BLUEPRINT_MYSQL_HOST")
	mysqlInstance *mysqlService
)

func NewMySQL() Service {
	if mysqlInstance != nil {
		return mysqlInstance
	}
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", mysqlUsername, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}
	mysqlInstance = &mysqlService{
		db: db,
	}
	return mysqlInstance
}

func (s *mysqlService) Health() map[string]string {
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

	if dbStats.OpenConnections > 40 {
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

func (s *mysqlService) Close() error {
	log.Printf("Disconnected from MySQL database: %s", mysqlDatabase)
	return s.db.Close()
}
func (s *mysqlService) Create(ctx context.Context, table string, data map[string]interface{}) (interface{}, error) {
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
		return res.RowsAffected()
	}
	return id, nil
}

func (s *mysqlService) Read(ctx context.Context, table string, filter map[string]interface{}) ([]map[string]interface{}, error) {
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
	results := []map[string]interface{}{}
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
	return results, nil
}

func (s *mysqlService) Update(ctx context.Context, table string, filter map[string]interface{}, update map[string]interface{}) (int64, error) {
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
	return res.RowsAffected()
}

func (s *mysqlService) Delete(ctx context.Context, table string, filter map[string]interface{}) (int64, error) {
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
	return res.RowsAffected()
}

func (s *mysqlService) Exec(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *mysqlService) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	results := []map[string]interface{}{}
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
	return results, nil
}