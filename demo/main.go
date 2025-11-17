package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/kohkimakimoto/echo-debugmonitor/monitors"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

func main() {
	e := echo.New()

	m := debugmonitor.New()

	// ----------------------------------------------
	// requests monitor
	// ----------------------------------------------
	requestsMonitor, requestsMonitorMiddleware := monitors.NewRequestsMonitor(&monitors.RequestsMonitorConfig{
		Skipper: func(c echo.Context) bool {
			// Skip monitoring requests to the /monitor endpoint
			return c.Path() == "/monitor"
		},
	})
	// Apply the middleware to monitor all incoming requests
	e.Use(requestsMonitorMiddleware)
	m.AddMonitor(requestsMonitor)

	// ----------------------------------------------
	// logs monitor
	// ----------------------------------------------
	logsMonitor, wrappedLogger := monitors.NewLogsMonitor(e.Logger)
	e.Logger = wrappedLogger
	m.AddMonitor(logsMonitor)

	// ----------------------------------------------
	// writer monitor
	// ----------------------------------------------
	m.AddMonitor(monitors.NewLoggerWriterMonitor(e.Logger))

	// ----------------------------------------------
	// queries monitor
	// ----------------------------------------------
	// Setup SQLite in-memory database with monitoring
	dsn := ":memory:"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

	// Wrap the database driver with query monitoring
	var queriesMonitor *debugmonitor.Monitor
	queriesMonitor, db = monitors.NewQueriesMonitor(dsn, db.Driver())
	m.AddMonitor(queriesMonitor)

	// Initialize database schema
	initDB(db, e)

	// ----------------------------------------------
	// errors monitor
	// ----------------------------------------------
	errorsMonitor, errorRecorder := monitors.NewErrorsMonitor()
	m.AddMonitor(errorsMonitor)

	// Wrap the default error handler to record errors
	e.HTTPErrorHandler = monitors.HTTPErrorHandlerWrapper(errorRecorder, e.HTTPErrorHandler)

	// Register the monitor handler
	e.GET("/monitor", m.Handler())

	// Test endpoints to demonstrate various request types
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "Test endpoint - check the monitor!")
	})

	// Test endpoint for different log levels
	e.GET("/test/logs", func(c echo.Context) error {
		e.Logger.Debug("This is a debug message")
		e.Logger.Info("This is an info message")
		e.Logger.Warn("This is a warning message")
		e.Logger.Error("This is an error message")
		e.Logger.Print("This is a print message")
		return c.String(http.StatusOK, "Check the logger monitor for different log levels!")
	})

	// Endpoint that simulates slow response
	e.GET("/slow", func(c echo.Context) error {
		e.Logger.Info("Slow endpoint called - processing...")
		// Simulate slow processing
		time.Sleep(2 * time.Second)
		e.Logger.Info("Slow endpoint completed")
		return c.String(http.StatusOK, "Slow endpoint - took 2 seconds")
	})

	// Test endpoint for database queries
	e.GET("/test/db/select", func(c echo.Context) error {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error: "+err.Error())
		}
		return c.String(http.StatusOK, fmt.Sprintf("Total users: %d", count))
	})

	e.GET("/test/db/insert", func(c echo.Context) error {
		name := c.QueryParam("name")
		if name == "" {
			name = "Test User"
		}
		result, err := db.Exec("INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)",
			name, fmt.Sprintf("%s@example.com", name), time.Now())
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error: "+err.Error())
		}
		id, _ := result.LastInsertId()
		return c.String(http.StatusOK, fmt.Sprintf("Inserted user with ID: %d", id))
	})

	e.GET("/test/db/query", func(c echo.Context) error {
		rows, err := db.Query("SELECT id, name, email FROM users ORDER BY id DESC LIMIT 10")
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error: "+err.Error())
		}
		defer rows.Close()

		type User struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		var users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
				return c.String(http.StatusInternalServerError, "Error: "+err.Error())
			}
			users = append(users, user)
		}

		return c.JSON(http.StatusOK, users)
	})

	e.GET("/test/db/transaction", func(c echo.Context) error {
		tx, err := db.Begin()
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error: "+err.Error())
		}

		_, err = tx.Exec("INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)",
			"Transaction User 1", "tx1@example.com", time.Now())
		if err != nil {
			tx.Rollback()
			return c.String(http.StatusInternalServerError, "Error: "+err.Error())
		}

		_, err = tx.Exec("INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)",
			"Transaction User 2", "tx2@example.com", time.Now())
		if err != nil {
			tx.Rollback()
			return c.String(http.StatusInternalServerError, "Error: "+err.Error())
		}

		if err := tx.Commit(); err != nil {
			return c.String(http.StatusInternalServerError, "Error: "+err.Error())
		}

		return c.String(http.StatusOK, "Transaction completed successfully")
	})

	// Test endpoints for error monitoring
	e.GET("/test/error/400", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "This is a bad request error")
	})

	e.GET("/test/error/404", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Resource not found")
	})

	e.GET("/test/error/500", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error occurred")
	})

	e.GET("/test/error/custom", func(c echo.Context) error {
		// Record a custom error manually
		customErr := fmt.Errorf("custom application error: something went wrong in business logic")
		errorRecorder(customErr)
		return c.String(http.StatusOK, "Custom error recorded - check the errors monitor!")
	})

	e.GET("/test/error/wrapped", func(c echo.Context) error {
		// Simulate a nested error scenario
		err := simulateNestedError()
		errorRecorder(err)
		return c.String(http.StatusOK, "Wrapped error with context recorded - check the errors monitor!")
	})

	e.GET("/test/error/stacktrace", func(c echo.Context) error {
		// Demonstrate error with stack trace using pkg/errors
		err := simulateErrorWithStackTrace()
		errorRecorder(err)
		return c.String(http.StatusOK, "Error with stack trace recorded - check the errors monitor!")
	})

	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err)
	}
}

func initDB(db *sql.DB, e *echo.Echo) {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		e.Logger.Fatal("Failed to create users table: ", err)
	}

	// Insert some sample data
	for i := 1; i <= 5; i++ {
		_, err := db.Exec("INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)",
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@example.com", i),
			time.Now())
		if err != nil {
			e.Logger.Fatal("Failed to insert sample data: ", err)
		}
	}

	e.Logger.Info("Database initialized with sample data")
}

// simulateNestedError demonstrates error wrapping with context
func simulateNestedError() error {
	// Simulate a low-level error
	err := processData()
	if err != nil {
		// Wrap the error with additional context
		return fmt.Errorf("failed to process user data: %w", err)
	}
	return nil
}

func processData() error {
	// Simulate a database or validation error
	return fmt.Errorf("invalid data format: expected JSON but got XML")
}

// simulateErrorWithStackTrace demonstrates error with stack trace using pkg/errors
func simulateErrorWithStackTrace() error {
	// Call a nested function to create a more interesting stack trace
	return performDatabaseOperation()
}

func performDatabaseOperation() error {
	// Simulate calling another function
	err := validateUserInput()
	if err != nil {
		// Wrap the error with stack trace
		return errors.Wrap(err, "database operation failed")
	}
	return nil
}

func validateUserInput() error {
	// Create a new error with stack trace
	return errors.New("validation failed: username cannot be empty")
}
