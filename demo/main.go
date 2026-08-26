package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor/v5"
	"github.com/kohkimakimoto/echo-debugmonitor/v5/monitors"
	"github.com/labstack/echo/v5"
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
		Skipper: func(c *echo.Context) bool {
			return c.Path() == "/monitor"
		},
	})
	e.Use(requestsMonitorMiddleware)
	m.AddMonitor(requestsMonitor)

	// ----------------------------------------------
	// logs monitor
	// ----------------------------------------------
	logsMonitor, wrappedLogger := monitors.NewLogsMonitor(monitors.LogsMonitorConfig{
		Logger: e.Logger,
	})
	e.Logger = wrappedLogger
	m.AddMonitor(logsMonitor)

	// ----------------------------------------------
	// queries monitor
	// ----------------------------------------------
	dsn := ":memory:"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		e.Logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	var queriesMonitor *debugmonitor.Monitor
	queriesMonitor, db = monitors.NewQueriesMonitor(monitors.QueriesMonitorConfig{
		DSN:    dsn,
		Driver: db.Driver(),
	})
	m.AddMonitor(queriesMonitor)

	initDB(db, e)

	// ----------------------------------------------
	// errors monitor
	// ----------------------------------------------
	errorsMonitor, errorRecorder := monitors.NewErrorsMonitor(monitors.ErrorsMonitorConfig{})
	m.AddMonitor(errorsMonitor)

	e.HTTPErrorHandler = monitors.HTTPErrorHandlerWrapper(errorRecorder, e.HTTPErrorHandler)

	e.GET("/monitor", m.Handler())

	e.GET("/test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Test endpoint - check the monitor!")
	})

	e.GET("/test/logs", func(c *echo.Context) error {
		e.Logger.Debug("This is a debug message")
		e.Logger.Info("This is an info message")
		e.Logger.Warn("This is a warning message")
		e.Logger.Error("This is an error message")
		return c.String(http.StatusOK, "Check the logger monitor for different log levels!")
	})

	e.GET("/slow", func(c *echo.Context) error {
		e.Logger.Info("Slow endpoint called - processing...")
		time.Sleep(2 * time.Second)
		e.Logger.Info("Slow endpoint completed")
		return c.String(http.StatusOK, "Slow endpoint - took 2 seconds")
	})

	e.GET("/test/db/select", func(c *echo.Context) error {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error: "+err.Error())
		}
		return c.String(http.StatusOK, fmt.Sprintf("Total users: %d", count))
	})

	e.GET("/test/db/insert", func(c *echo.Context) error {
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

	e.GET("/test/db/query", func(c *echo.Context) error {
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

	e.GET("/test/db/transaction", func(c *echo.Context) error {
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

	e.GET("/test/error/400", func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "This is a bad request error")
	})

	e.GET("/test/error/404", func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Resource not found")
	})

	e.GET("/test/error/500", func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error occurred")
	})

	e.GET("/test/error/custom", func(c *echo.Context) error {
		customErr := fmt.Errorf("custom application error: something went wrong in business logic")
		errorRecorder(customErr)
		return c.String(http.StatusOK, "Custom error recorded - check the errors monitor!")
	})

	e.GET("/test/error/wrapped", func(c *echo.Context) error {
		err := simulateNestedError()
		errorRecorder(err)
		return c.String(http.StatusOK, "Wrapped error with context recorded - check the errors monitor!")
	})

	e.GET("/test/error/stacktrace", func(c *echo.Context) error {
		err := simulateErrorWithStackTrace()
		errorRecorder(err)
		return c.String(http.StatusOK, "Error with stack trace recorded - check the errors monitor!")
	})

	if err := e.Start(":8080"); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}

func initDB(db *sql.DB, e *echo.Echo) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		e.Logger.Error("Failed to create users table", "error", err)
		os.Exit(1)
	}

	for i := 1; i <= 5; i++ {
		_, err := db.Exec("INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)",
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@example.com", i),
			time.Now())
		if err != nil {
			e.Logger.Error("Failed to insert sample data", "error", err)
			os.Exit(1)
		}
	}

	e.Logger.Info("Database initialized with sample data")
}

func simulateNestedError() error {
	err := processData()
	if err != nil {
		return fmt.Errorf("failed to process user data: %w", err)
	}
	return nil
}

func processData() error {
	return fmt.Errorf("invalid data format: expected JSON but got XML")
}

func simulateErrorWithStackTrace() error {
	return performDatabaseOperation()
}

func performDatabaseOperation() error {
	err := validateUserInput()
	if err != nil {
		return errors.Wrap(err, "database operation failed")
	}
	return nil
}

func validateUserInput() error {
	return errors.New("validation failed: username cannot be empty")
}
