package debugmonitor

import (
	"testing"
	"time"
)

func TestMonitor_WriteWithStoreIntegration(t *testing.T) {
	// Create a manager and monitor
	mgr := New()
	mon := &Monitor{
		Name:              "test-monitor",
		DisplayName:       "Test Monitor",
		MaxRecords:        10,
		ChannelBufferSize: 10,
	}

	mgr.AddMonitor(mon)

	// Write some data
	for i := 1; i <= 5; i++ {
		err := mon.Write(Data{
			"message": "test message",
			"index":   i,
		})
		if err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}
	}

	// Give the goroutine time to process
	time.Sleep(100 * time.Millisecond)

	// Verify data was stored using GetDataSince("") which returns all records
	allData := mon.GetDataSince("")
	if len(allData) != 5 {
		t.Errorf("Expected 5 records, got %d", len(allData))
	}

	// Store IDs for later verification
	var ids []string
	for _, data := range allData {
		id, ok := data["id"].(string)
		if !ok {
			t.Fatalf("Expected ID to be string, got %T", data["id"])
		}
		ids = append(ids, id)
	}

	// Test GetLatestData
	latest := mon.GetLatestData(3)
	if len(latest) != 3 {
		t.Errorf("Expected 3 latest records, got %d", len(latest))
	}

	// Should be in reverse order (newest first)
	expectedIDs := []string{ids[4], ids[3], ids[2]}
	for i, data := range latest {
		if data["id"] != expectedIDs[i] {
			t.Errorf("Expected ID %s at position %d, got %v", expectedIDs[i], i, data["id"])
		}
	}

	// Test GetDataSince for cursor-based pagination
	since := mon.GetDataSince(ids[1])
	if len(since) != 3 {
		t.Errorf("Expected 3 records since ID %s, got %d", ids[1], len(since))
	}

	// Should be in chronological order
	expectedSinceIDs := []string{ids[2], ids[3], ids[4]}
	for i, data := range since {
		if data["id"] != expectedSinceIDs[i] {
			t.Errorf("Expected ID %s at position %d, got %v", expectedSinceIDs[i], i, data["id"])
		}
		// Verify data structure
		if data["message"] != "test message" {
			t.Errorf("Expected message 'test message', got %v", data["message"])
		}
	}
}

func TestMonitor_MaxRecordsLimit(t *testing.T) {
	// Create a manager and monitor with small MaxRecords
	mgr := New()
	mon := &Monitor{
		Name:              "test-monitor",
		DisplayName:       "Test Monitor",
		MaxRecords:        3,
		ChannelBufferSize: 20,
	}

	mgr.AddMonitor(mon)

	// Write 5 records (exceeds limit of 3)
	for i := 1; i <= 5; i++ {
		err := mon.Write(Data{
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
	allData := mon.GetDataSince("")
	if len(allData) != 3 {
		t.Errorf("Expected 3 records, got %d", len(allData))
	}

	// Verify we have 3 records with valid UUIDs
	for i, data := range allData {
		id, ok := data["id"].(string)
		if !ok {
			t.Errorf("Expected ID to be string at position %d, got %T", i, data["id"])
		}
		if len(id) != 36 {
			t.Errorf("Expected UUID length 36 at position %d, got %d", i, len(id))
		}
	}
}

func TestMonitor_ConcurrentWrites(t *testing.T) {
	mgr := New()
	mon := &Monitor{
		Name:              "test-monitor",
		DisplayName:       "Test Monitor",
		MaxRecords:        1000,
		ChannelBufferSize: 200,
	}

	mgr.AddMonitor(mon)

	// Write records concurrently from multiple goroutines
	const numGoroutines = 10
	const writesPerGoroutine = 10

	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < writesPerGoroutine; j++ {
				mon.Write(Data{
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
	allData := mon.GetDataSince("")
	expectedCount := numGoroutines * writesPerGoroutine
	if len(allData) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(allData))
	}

	// Verify all IDs are unique UUIDs
	seen := make(map[string]bool)
	for _, data := range allData {
		id, ok := data["id"].(string)
		if !ok {
			t.Fatalf("Expected ID to be string, got %T", data["id"])
		}
		if seen[id] {
			t.Errorf("Duplicate ID found: %s", id)
		}
		if len(id) != 36 {
			t.Errorf("Expected UUID length 36, got %d", len(id))
		}
		seen[id] = true
	}
}
