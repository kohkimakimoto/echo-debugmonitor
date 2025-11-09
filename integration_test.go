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
		DisplayName: "Test monitor",
		MaxRecords:  10,
	}

	mgr.AddMonitor(mon)

	// Write some data
	for i := 1; i <= 5; i++ {
		mon.Write(map[string]any{
			"message": "test message",
			"index":   i,
		})
	}

	// Give the goroutine time to process
	time.Sleep(100 * time.Millisecond)

	// Verify data was stored using store.GetSince(0) which returns all records
	allData := mon.store.GetSince(0)
	if len(allData) != 5 {
		t.Errorf("Expected 5 records, got %d", len(allData))
	}

	// Store IDs for later verification
	var ids []int64
	for _, entry := range allData {
		ids = append(ids, entry.Id)
	}

	// Test store.GetLatest
	latest := mon.store.GetLatest(3)
	if len(latest) != 3 {
		t.Errorf("Expected 3 latest records, got %d", len(latest))
	}

	// Should be in reverse order (newest first)
	expectedIDs := []int64{ids[4], ids[3], ids[2]}
	for i, entry := range latest {
		if entry.Id != expectedIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %v", expectedIDs[i], i, entry.Id)
		}
	}

	// Test store.GetSince for cursor-based pagination
	since := mon.store.GetSince(ids[1])
	if len(since) != 3 {
		t.Errorf("Expected 3 records since ID %d, got %d", ids[1], len(since))
	}

	// Should be in chronological order
	expectedSinceIDs := []int64{ids[2], ids[3], ids[4]}
	for i, entry := range since {
		if entry.Id != expectedSinceIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %v", expectedSinceIDs[i], i, entry.Id)
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
		DisplayName: "Test monitor",
		MaxRecords:  3,
	}

	mgr.AddMonitor(mon)

	// Write 5 records (exceeds limit of 3)
	for i := 1; i <= 5; i++ {
		mon.Write(map[string]any{
			"message": "test message",
			"index":   i,
		})
	}

	// Give the goroutine time to process all records
	time.Sleep(100 * time.Millisecond)

	// Should only have 3 records (the most recent ones)
	allData := mon.store.GetSince(0)
	if len(allData) != 3 {
		t.Errorf("Expected 3 records, got %d", len(allData))
	}

	// Verify we have 3 records with valid int64 IDs
	for i, entry := range allData {
		if entry.Id <= 0 {
			t.Errorf("Expected positive ID at position %d, got %d", i, entry.Id)
		}
	}
}

func TestMonitor_ConcurrentWrites(t *testing.T) {
	mgr := New()
	mon := &Monitor{
		Name:        "test-monitor",
		DisplayName: "Test monitor",
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
	allData := mon.store.GetSince(0)
	expectedCount := numGoroutines * writesPerGoroutine
	if len(allData) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(allData))
	}

	// Verify all IDs are unique and positive
	seen := make(map[int64]bool)
	for _, entry := range allData {
		if seen[entry.Id] {
			t.Errorf("Duplicate ID found: %d", entry.Id)
		}
		if entry.Id <= 0 {
			t.Errorf("Expected positive ID, got %d", entry.Id)
		}
		seen[entry.Id] = true
	}
}
