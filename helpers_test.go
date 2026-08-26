package debugmonitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestHandleSSEStreamSendsInitialDataAndStopsOnDisconnect(t *testing.T) {
	requestContext, cancel := context.WithCancel(context.Background())
	request := httptest.NewRequest(http.MethodGet, "/stream", nil).WithContext(requestContext)
	response := newFlushRecorder()
	ctx := echo.New().NewContext(request, response)
	store := NewStore(10)
	store.Add(map[string]string{"message": "ready"})

	done := make(chan error, 1)
	go func() {
		done <- HandleSSEStream(ctx, store)
	}()

	select {
	case <-response.flushed:
		cancel()
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("timed out waiting for the initial SSE flush")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("SSE handler returned an error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("SSE handler did not stop after the request was canceled")
	}

	if contentType := response.Header().Get("Content-Type"); contentType != "text/event-stream" {
		t.Fatalf("expected SSE content type, got %q", contentType)
	}
	body := response.Body.String()
	if !strings.Contains(body, `"message":"ready"`) {
		t.Fatalf("expected initial SSE data, got %q", body)
	}
	// Ensure Snowflake IDs are string-encoded so browsers keep full precision.
	if !strings.Contains(body, `"id":"`) {
		t.Fatalf("expected SSE id to be a JSON string, got %q", body)
	}
}

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed chan struct{}
	once    sync.Once
}

func newFlushRecorder() *flushRecorder {
	return &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		flushed:          make(chan struct{}),
	}
}

func (r *flushRecorder) Flush() {
	r.ResponseRecorder.Flush()
	r.once.Do(func() {
		close(r.flushed)
	})
}
