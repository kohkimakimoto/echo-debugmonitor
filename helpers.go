package debugmonitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
)

// RenderTemplate executes a template with the given data and returns the result as HTML response.
func RenderTemplate(c *echo.Context, tmpl *template.Template, data any) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}
	return c.HTML(http.StatusOK, buf.String())
}

func HandleSSEStream(c *echo.Context, store *Store) error {
	sinceID := int64(0)
	if sinceIDStr := c.QueryParam("since"); sinceIDStr != "" {
		if id, err := strconv.ParseInt(sinceIDStr, 10, 64); err == nil {
			sinceID = id
		}
	}

	rw := c.Response()
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.WriteHeader(http.StatusOK)

	addEvent := store.NewAddEvent()
	defer addEvent.Close()

	entries := store.GetSince(sinceID)
	for _, entry := range entries {
		if err := sendSSEEvent(rw, entry); err != nil {
			return err
		}
		sinceID = entry.Id
	}

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	ctx := c.Request().Context()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry, ok := <-addEvent.C:
			if !ok {
				return nil
			}
			if err := sendSSEEvent(rw, entry); err != nil {
				return err
			}
			if f, ok := rw.(http.Flusher); ok {
				f.Flush()
			}
		case <-ticker.C:
			fmt.Fprintf(rw, ": keepalive\n\n")
			if f, ok := rw.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

func sendSSEEvent(rw http.ResponseWriter, entry *DataEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(rw, "data: %s\n\n", data)
	return err
}

// HandleDataJSON returns store entries as JSON for polling mode.
// It accepts a "since" query parameter to return only entries with ID greater than the specified value.
func HandleDataJSON(c *echo.Context, store *Store) error {
	sinceID := int64(0)
	if sinceIDStr := c.QueryParam("since"); sinceIDStr != "" {
		if id, err := strconv.ParseInt(sinceIDStr, 10, 64); err == nil {
			sinceID = id
		}
	}

	entries := store.GetSince(sinceID)
	return c.JSON(http.StatusOK, entries)
}
