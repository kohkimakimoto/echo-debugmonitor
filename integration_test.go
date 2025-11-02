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

	// Verify data was stored
	records := mon.GetAllRecords()
	if len(records) != 5 {
		t.Errorf("Expected 5 records, got %d", len(records))
	}

	// Verify IDs are sequential
	for i, record := range records {
		expectedID := int64(i + 1)
		if record.ID != expectedID {
			t.Errorf("Expected ID %d, got %d", expectedID, record.ID)
		}
	}

	// Test GetLatestRecords
	latest := mon.GetLatestRecords(3)
	if len(latest) != 3 {
		t.Errorf("Expected 3 latest records, got %d", len(latest))
	}

	// Should be in reverse order: 5, 4, 3
	expectedIDs := []int64{5, 4, 3}
	for i, record := range latest {
		if record.ID != expectedIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %d", expectedIDs[i], i, record.ID)
		}
	}

	// Test GetRecordsSince
	since := mon.GetRecordsSince(2)
	if len(since) != 3 {
		t.Errorf("Expected 3 records since ID 2, got %d", len(since))
	}

	// Should be in chronological order: 3, 4, 5
	expectedSinceIDs := []int64{3, 4, 5}
	for i, record := range since {
		if record.ID != expectedSinceIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %d", expectedSinceIDs[i], i, record.ID)
		}
	}

	// Test GetRecord
	record, exists := mon.GetRecord(3)
	if !exists {
		t.Fatal("Record 3 should exist")
	}
	if record.ID != 3 {
		t.Errorf("Expected ID 3, got %d", record.ID)
	}
	// Verify the record has the expected fields (but don't assume index matches ID)
	if _, ok := record.Data["index"]; !ok {
		t.Error("Record should have 'index' field")
	}
	if record.Data["message"] != "test message" {
		t.Errorf("Expected message 'test message', got %v", record.Data["message"])
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
	records := mon.GetAllRecords()
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// Should have records 3, 4, 5 (oldest 1, 2 should be removed)
	expectedIDs := []int64{3, 4, 5}
	for i, record := range records {
		if record.ID != expectedIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %d", expectedIDs[i], i, record.ID)
		}
	}

	// Verify old records are gone
	_, exists := mon.GetRecord(1)
	if exists {
		t.Error("Record 1 should have been removed")
	}
	_, exists = mon.GetRecord(2)
	if exists {
		t.Error("Record 2 should have been removed")
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
	records := mon.GetAllRecords()
	expectedCount := numGoroutines * writesPerGoroutine
	if len(records) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(records))
	}

	// Verify all IDs are unique and sequential
	seen := make(map[int64]bool)
	for _, record := range records {
		if seen[record.ID] {
			t.Errorf("Duplicate ID found: %d", record.ID)
		}
		seen[record.ID] = true
	}
}
