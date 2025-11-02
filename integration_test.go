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

	// Verify data was stored using GetDataSince(0) which returns all records
	allData := mon.GetDataSince(0)
	if len(allData) != 5 {
		t.Errorf("Expected 5 records, got %d", len(allData))
	}

	// Verify IDs are sequential
	for i, data := range allData {
		expectedID := int64(i + 1)
		if data["id"] != expectedID {
			t.Errorf("Expected ID %d, got %v", expectedID, data["id"])
		}
	}

	// Test GetLatestData
	latest := mon.GetLatestData(3)
	if len(latest) != 3 {
		t.Errorf("Expected 3 latest records, got %d", len(latest))
	}

	// Should be in reverse order: 5, 4, 3
	expectedIDs := []int64{5, 4, 3}
	for i, data := range latest {
		if data["id"] != expectedIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %v", expectedIDs[i], i, data["id"])
		}
	}

	// Test GetDataSince for cursor-based pagination
	since := mon.GetDataSince(2)
	if len(since) != 3 {
		t.Errorf("Expected 3 records since ID 2, got %d", len(since))
	}

	// Should be in chronological order: 3, 4, 5
	expectedSinceIDs := []int64{3, 4, 5}
	for i, data := range since {
		if data["id"] != expectedSinceIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %v", expectedSinceIDs[i], i, data["id"])
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
	allData := mon.GetDataSince(0)
	if len(allData) != 3 {
		t.Errorf("Expected 3 records, got %d", len(allData))
	}

	// Should have records 3, 4, 5 (oldest 1, 2 should be removed)
	expectedIDs := []int64{3, 4, 5}
	for i, data := range allData {
		if data["id"] != expectedIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %v", expectedIDs[i], i, data["id"])
		}
	}

	// Verify old records are gone by checking GetDataSince returns no records before ID 3
	beforeThree := mon.GetDataSince(0)
	if len(beforeThree) > 0 && beforeThree[0]["id"].(int64) < 3 {
		t.Error("Records 1 and 2 should have been removed")
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
	allData := mon.GetDataSince(0)
	expectedCount := numGoroutines * writesPerGoroutine
	if len(allData) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(allData))
	}

	// Verify all IDs are unique and sequential
	seen := make(map[int64]bool)
	for _, data := range allData {
		id := data["id"].(int64)
		if seen[id] {
			t.Errorf("Duplicate ID found: %d", id)
		}
		seen[id] = true
	}
}
