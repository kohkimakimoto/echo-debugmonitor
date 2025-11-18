package debugmonitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// RenderTemplate executes a template with the given data and returns the result as HTML response.
func RenderTemplate(c echo.Context, tmpl *template.Template, data any) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}
	return c.HTML(http.StatusOK, buf.String())
}

func HandleSSEStream(c echo.Context, store *Store) error {
	// Parse the sinceID parameter
	sinceID := int64(0)
	if sinceIDStr := c.QueryParam("since"); sinceIDStr != "" {
		if id, err := strconv.ParseInt(sinceIDStr, 10, 64); err == nil {
			sinceID = id
		}
	}

	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	// Subscribe to add events
	addEvent := store.NewAddEvent()
	defer addEvent.Close()

	// Send initial data since the provided ID
	entries := store.GetSince(sinceID)
	for _, entry := range entries {
		if err := sendSSEEvent(c, entry); err != nil {
			return err
		}
		sinceID = entry.Id
	}

	// Flush to send initial data
	if f, ok := c.Response().Writer.(http.Flusher); ok {
		f.Flush()
	}

	// Listen for new add events
	ctx := c.Request().Context()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return nil
		case entry, ok := <-addEvent.C:
			if !ok {
				// Channel closed
				return nil
			}
			if err := sendSSEEvent(c, entry); err != nil {
				return err
			}
			if f, ok := c.Response().Writer.(http.Flusher); ok {
				f.Flush()
			}
		case <-ticker.C:
			// Send a comment as keepalive
			fmt.Fprintf(c.Response().Writer, ": keepalive\n\n")
			if f, ok := c.Response().Writer.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

func sendSSEEvent(c echo.Context, entry *DataEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.Response().Writer, "data: %s\n\n", data)
	return err
}

// HandleDataJSON returns store entries as JSON for polling mode.
// It accepts a "since" query parameter to return only entries with ID greater than the specified value.
func HandleDataJSON(c echo.Context, store *Store) error {
	// Parse the sinceID parameter
	sinceID := int64(0)
	if sinceIDStr := c.QueryParam("since"); sinceIDStr != "" {
		if id, err := strconv.ParseInt(sinceIDStr, 10, 64); err == nil {
			sinceID = id
		}
	}

	entries := store.GetSince(sinceID)
	return c.JSON(http.StatusOK, entries)
}
