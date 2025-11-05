package debugmonitor

import (
	"testing"
	"time"
)

func TestMonitor_WriteWithStoreIntegration(t *testing.T) {
	// Create a manager and monitor
	mgr := New()
	mon := &Monitor{
		Name:        "test-monitor",
		DisplayName: "Test Monitor",
		MaxRecords:  10,
	}

	mgr.AddMonitor(mon)

	// Write some data
	for i := 1; i <= 5; i++ {
		err := mon.Write(map[string]any{
			"message": "test message",
			"index":   i,
		})
		if err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}
	}

	// Give the goroutine time to process
	time.Sleep(100 * time.Millisecond)

	// Verify data was stored using store.GetSince("") which returns all records
	allData := mon.store.GetSince("")
	if len(allData) != 5 {
		t.Errorf("Expected 5 records, got %d", len(allData))
	}

	// Store IDs for later verification
	var ids []string
	for _, entry := range allData {
		ids = append(ids, entry.Id)
	}

	// Test store.GetLatest
	latest := mon.store.GetLatest(3)
	if len(latest) != 3 {
		t.Errorf("Expected 3 latest records, got %d", len(latest))
	}

	// Should be in reverse order (newest first)
	expectedIDs := []string{ids[4], ids[3], ids[2]}
	for i, entry := range latest {
		if entry.Id != expectedIDs[i] {
			t.Errorf("Expected ID %s at position %d, got %v", expectedIDs[i], i, entry.Id)
		}
	}

	// Test store.GetSince for cursor-based pagination
	since := mon.store.GetSince(ids[1])
	if len(since) != 3 {
		t.Errorf("Expected 3 records since ID %s, got %d", ids[1], len(since))
	}

	// Should be in chronological order
	expectedSinceIDs := []string{ids[2], ids[3], ids[4]}
	for i, entry := range since {
		if entry.Id != expectedSinceIDs[i] {
			t.Errorf("Expected ID %s at position %d, got %v", expectedSinceIDs[i], i, entry.Id)
		}
		// Verify data structure
		payload := entry.Payload.(map[string]any)
		if payload["message"] != "test message" {
			t.Errorf("Expected message 'test message', got %v", payload["message"])
		}
	}
}

func TestMonitor_MaxRecordsLimit(t *testing.T) {
	// Create a manager and monitor with small MaxRecords
	mgr := New()
	mon := &Monitor{
		Name:        "test-monitor",
		DisplayName: "Test Monitor",
		MaxRecords:  3,
	}

	mgr.AddMonitor(mon)

	// Write 5 records (exceeds limit of 3)
	for i := 1; i <= 5; i++ {
		err := mon.Write(map[string]any{
			"message": "test message",
			"index":   i,
		})
		if err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}
	}

	// Give the goroutine time to process all records
	time.Sleep(100 * time.Millisecond)

	// Should only have 3 records (the most recent ones)
	allData := mon.store.GetSince("")
	if len(allData) != 3 {
		t.Errorf("Expected 3 records, got %d", len(allData))
	}

	// Verify we have 3 records with valid UUIDs
	for i, entry := range allData {
		if len(entry.Id) != 36 {
			t.Errorf("Expected UUID length 36 at position %d, got %d", i, len(entry.Id))
		}
	}
}

func TestMonitor_ConcurrentWrites(t *testing.T) {
	mgr := New()
	mon := &Monitor{
		Name:        "test-monitor",
		DisplayName: "Test Monitor",
		MaxRecords:  1000,
	}

	mgr.AddMonitor(mon)

	// Write records concurrently from multiple goroutines
	const numGoroutines = 10
	const writesPerGoroutine = 10

	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < writesPerGoroutine; j++ {
				mon.Write(map[string]any{
					"goroutine": goroutineID,
					"index":     j,
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Give time for all data to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify all records were stored
	allData := mon.store.GetSince("")
	expectedCount := numGoroutines * writesPerGoroutine
	if len(allData) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(allData))
	}

	// Verify all IDs are unique UUIDs
	seen := make(map[string]bool)
	for _, entry := range allData {
		if seen[entry.Id] {
			t.Errorf("Duplicate ID found: %s", entry.Id)
		}
		if len(entry.Id) != 36 {
			t.Errorf("Expected UUID length 36, got %d", len(entry.Id))
		}
		seen[entry.Id] = true
	}
}
