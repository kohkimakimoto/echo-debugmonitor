package monitors

import (
	"context"
	"database/sql"
	"database/sql/driver"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

// QueryPayload represents the data structure for database query monitoring
type QueryPayload struct {
	Query     string        `json:"query"`
	Args      []interface{} `json:"args,omitempty"`
	Duration  int64         `json:"duration"` // in milliseconds
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Operation string        `json:"operation"` // Query, Exec, Prepare, Begin, Commit, Rollback
}

//go:embed queries.html
var queriesView string

// queriesViewTemplate is the parsed template for the queries view
var queriesViewTemplate = template.Must(template.New("queriesView").Parse(queriesView))

// QueriesMonitorConfig defines the config for Queries monitor.
type QueriesMonitorConfig struct {
	// DSN is the data source name for the database connection.
	DSN string
	// Driver is the database driver to wrap with monitoring.
	Driver driver.Driver
	// UsePolling enables polling mode instead of SSE for real-time updates.
	UsePolling bool
}

// NewQueriesMonitor creates a new monitor for database queries and returns a wrapped *sql.DB.
// This function wraps an existing database driver with monitoring capabilities without requiring
// changes to existing *sql.DB usage code.
func NewQueriesMonitor(config QueriesMonitorConfig) (*debugmonitor.Monitor, *sql.DB) {
	m := &debugmonitor.Monitor{
		Name:        "queries",
		DisplayName: "Queries",
		MaxRecords:  1000,
		Icon:        debugmonitor.IconCircleStack,
		ActionHandler: func(c echo.Context, store *debugmonitor.Store, action string) error {
			switch action {
			case "render":
				return debugmonitor.RenderTemplate(c, queriesViewTemplate, map[string]any{
					"UsePolling": config.UsePolling,
				})
			case "stream":
				// SSE endpoint for real-time updates
				return debugmonitor.HandleSSEStream(c, store)
			case "data":
				// JSON endpoint for polling mode
				return debugmonitor.HandleDataJSON(c, store)
			default:
				return echo.NewHTTPError(http.StatusBadRequest)
			}
		},
	}

	// Create a monitored connector
	connector := &monitoredConnector{
		driver:  config.Driver,
		dsn:     config.DSN,
		monitor: m,
	}

	// Open database with the monitored connector
	db := sql.OpenDB(connector)

	return m, db
}

// monitoredConnector implements driver.Connector
type monitoredConnector struct {
	driver  driver.Driver
	dsn     string
	monitor *debugmonitor.Monitor
}

func (c *monitoredConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.driver.Open(c.dsn)
	if err != nil {
		return nil, err
	}
	return &monitoredConn{conn: conn, monitor: c.monitor}, nil
}

func (c *monitoredConnector) Driver() driver.Driver {
	return c.driver
}

// monitoredConn wraps a sql connection
type monitoredConn struct {
	conn    driver.Conn
	monitor *debugmonitor.Monitor
}

func (c *monitoredConn) Prepare(query string) (driver.Stmt, error) {
	start := time.Now()
	stmt, err := c.conn.Prepare(query)
	duration := time.Since(start)

	payload := &QueryPayload{
		Query:     query,
		Duration:  duration.Milliseconds(),
		Timestamp: start,
		Operation: "Prepare",
	}
	if err != nil {
		payload.Error = err.Error()
	}
	c.monitor.Add(payload)

	if err != nil {
		return nil, err
	}
	return &monitoredStmt{stmt: stmt, query: query, monitor: c.monitor}, nil
}

func (c *monitoredConn) Close() error {
	return c.conn.Close()
}

func (c *monitoredConn) Begin() (driver.Tx, error) {
	start := time.Now()
	tx, err := c.conn.Begin()
	duration := time.Since(start)

	payload := &QueryPayload{
		Query:     "BEGIN",
		Duration:  duration.Milliseconds(),
		Timestamp: start,
		Operation: "Begin",
	}
	if err != nil {
		payload.Error = err.Error()
	}
	c.monitor.Add(payload)

	if err != nil {
		return nil, err
	}
	return &monitoredTx{tx: tx, monitor: c.monitor}, nil
}

// Implement ExecerContext interface
func (c *monitoredConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if execer, ok := c.conn.(driver.ExecerContext); ok {
		start := time.Now()
		result, err := execer.ExecContext(ctx, query, args)
		duration := time.Since(start)

		payload := &QueryPayload{
			Query:     query,
			Args:      namedValuesToInterface(args),
			Duration:  duration.Milliseconds(),
			Timestamp: start,
			Operation: "Exec",
		}
		if err != nil {
			payload.Error = err.Error()
		}
		c.monitor.Add(payload)

		return result, err
	}
	return nil, driver.ErrSkip
}

// Implement QueryerContext interface
func (c *monitoredConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if queryer, ok := c.conn.(driver.QueryerContext); ok {
		start := time.Now()
		rows, err := queryer.QueryContext(ctx, query, args)
		duration := time.Since(start)

		payload := &QueryPayload{
			Query:     query,
			Args:      namedValuesToInterface(args),
			Duration:  duration.Milliseconds(),
			Timestamp: start,
			Operation: "Query",
		}
		if err != nil {
			payload.Error = err.Error()
		}
		c.monitor.Add(payload)

		return rows, err
	}
	return nil, driver.ErrSkip
}

// monitoredStmt wraps a sql statement
type monitoredStmt struct {
	stmt    driver.Stmt
	query   string
	monitor *debugmonitor.Monitor
}

func (s *monitoredStmt) Close() error {
	return s.stmt.Close()
}

func (s *monitoredStmt) NumInput() int {
	return s.stmt.NumInput()
}

func (s *monitoredStmt) Exec(args []driver.Value) (driver.Result, error) {
	start := time.Now()
	result, err := s.stmt.Exec(args)
	duration := time.Since(start)

	payload := &QueryPayload{
		Query:     s.query,
		Args:      valuesToInterface(args),
		Duration:  duration.Milliseconds(),
		Timestamp: start,
		Operation: "Exec",
	}
	if err != nil {
		payload.Error = err.Error()
	}
	s.monitor.Add(payload)

	return result, err
}

func (s *monitoredStmt) Query(args []driver.Value) (driver.Rows, error) {
	start := time.Now()
	rows, err := s.stmt.Query(args)
	duration := time.Since(start)

	payload := &QueryPayload{
		Query:     s.query,
		Args:      valuesToInterface(args),
		Duration:  duration.Milliseconds(),
		Timestamp: start,
		Operation: "Query",
	}
	if err != nil {
		payload.Error = err.Error()
	}
	s.monitor.Add(payload)

	return rows, err
}

// monitoredTx wraps a sql transaction
type monitoredTx struct {
	tx      driver.Tx
	monitor *debugmonitor.Monitor
}

func (t *monitoredTx) Commit() error {
	start := time.Now()
	err := t.tx.Commit()
	duration := time.Since(start)

	payload := &QueryPayload{
		Query:     "COMMIT",
		Duration:  duration.Milliseconds(),
		Timestamp: start,
		Operation: "Commit",
	}
	if err != nil {
		payload.Error = err.Error()
	}
	t.monitor.Add(payload)

	return err
}

func (t *monitoredTx) Rollback() error {
	start := time.Now()
	err := t.tx.Rollback()
	duration := time.Since(start)

	payload := &QueryPayload{
		Query:     "ROLLBACK",
		Duration:  duration.Milliseconds(),
		Timestamp: start,
		Operation: "Rollback",
	}
	if err != nil {
		payload.Error = err.Error()
	}
	t.monitor.Add(payload)

	return err
}

// Helper functions to convert driver values to interface{}
func valuesToInterface(values []driver.Value) []interface{} {
	result := make([]interface{}, len(values))
	for i, v := range values {
		result[i] = v
	}
	return result
}

func namedValuesToInterface(values []driver.NamedValue) []interface{} {
	result := make([]interface{}, len(values))
	for i, v := range values {
		if v.Name != "" {
			result[i] = fmt.Sprintf("%s=%v", v.Name, v.Value)
		} else {
			result[i] = v.Value
		}
	}
	return result
}
