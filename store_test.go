package debugmonitor

import (
	"sync"
	"testing"
)

func TestStore_Add(t *testing.T) {
	store := NewStore(5)

	// Add some records
	for i := int64(1); i <= 3; i++ {
		store.Add(i, Data{"message": "test", "index": i})
	}

	if store.Len() != 3 {
		t.Errorf("Expected 3 records, got %d", store.Len())
	}
}

func TestStore_Get(t *testing.T) {
	store := NewStore(10)

	// Add a record
	store.Add(1, Data{"message": "hello"})

	// Get the record
	record, exists := store.Get(1)
	if !exists {
		t.Fatal("Record should exist")
	}

	if record.ID != 1 {
		t.Errorf("Expected ID 1, got %d", record.ID)
	}

	if record.Data["message"] != "hello" {
		t.Errorf("Expected message 'hello', got %v", record.Data["message"])
	}

	// Try to get a non-existent record
	_, exists = store.Get(999)
	if exists {
		t.Error("Record 999 should not exist")
	}
}

func TestStore_MaxRecords(t *testing.T) {
	store := NewStore(3)

	// Add 5 records (exceeds limit of 3)
	for i := int64(1); i <= 5; i++ {
		store.Add(i, Data{"index": i})
	}

	// Should only have 3 records
	if store.Len() != 3 {
		t.Errorf("Expected 3 records, got %d", store.Len())
	}

	// The oldest records (1, 2) should be removed
	_, exists := store.Get(1)
	if exists {
		t.Error("Record 1 should have been removed")
	}

	_, exists = store.Get(2)
	if exists {
		t.Error("Record 2 should have been removed")
	}

	// The newest records (3, 4, 5) should remain
	for i := int64(3); i <= 5; i++ {
		_, exists := store.Get(i)
		if !exists {
			t.Errorf("Record %d should exist", i)
		}
	}
}

func TestStore_GetLatest(t *testing.T) {
	store := NewStore(10)

	// Add records
	for i := int64(1); i <= 5; i++ {
		store.Add(i, Data{"index": i})
	}

	// Get latest 3 records
	latest := store.GetLatest(3)
	if len(latest) != 3 {
		t.Errorf("Expected 3 records, got %d", len(latest))
	}

	// Should be in reverse chronological order: 5, 4, 3
	expectedIDs := []int64{5, 4, 3}
	for i, record := range latest {
		if record.ID != expectedIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %d", expectedIDs[i], i, record.ID)
		}
	}

	// Request more than available
	all := store.GetLatest(100)
	if len(all) != 5 {
		t.Errorf("Expected 5 records, got %d", len(all))
	}
}

func TestStore_GetSince(t *testing.T) {
	store := NewStore(10)

	// Add records
	for i := int64(1); i <= 5; i++ {
		store.Add(i, Data{"index": i})
	}

	// Get records since ID 2
	since := store.GetSince(2)
	if len(since) != 3 {
		t.Errorf("Expected 3 records, got %d", len(since))
	}

	// Should be in chronological order: 3, 4, 5
	expectedIDs := []int64{3, 4, 5}
	for i, record := range since {
		if record.ID != expectedIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %d", expectedIDs[i], i, record.ID)
		}
	}

	// Get since non-existent ID
	since = store.GetSince(0)
	if len(since) != 5 {
		t.Errorf("Expected 5 records, got %d", len(since))
	}

	// Get since last ID
	since = store.GetSince(5)
	if len(since) != 0 {
		t.Errorf("Expected 0 records, got %d", len(since))
	}
}

func TestStore_GetAll(t *testing.T) {
	store := NewStore(10)

	// Add records
	for i := int64(1); i <= 5; i++ {
		store.Add(i, Data{"index": i})
	}

	all := store.GetAll()
	if len(all) != 5 {
		t.Errorf("Expected 5 records, got %d", len(all))
	}

	// Should be in chronological order
	for i, record := range all {
		expectedID := int64(i + 1)
		if record.ID != expectedID {
			t.Errorf("Expected ID %d at position %d, got %d", expectedID, i, record.ID)
		}
	}
}

func TestStore_Clear(t *testing.T) {
	store := NewStore(10)

	// Add records
	for i := int64(1); i <= 5; i++ {
		store.Add(i, Data{"index": i})
	}

	if store.Len() != 5 {
		t.Errorf("Expected 5 records before clear, got %d", store.Len())
	}

	// Clear the store
	store.Clear()

	if store.Len() != 0 {
		t.Errorf("Expected 0 records after clear, got %d", store.Len())
	}

	// Verify records are actually gone
	_, exists := store.Get(1)
	if exists {
		t.Error("Records should not exist after clear")
	}
}

func TestStore_Concurrency(t *testing.T) {
	store := NewStore(1000)
	const numGoroutines = 50
	const recordsPerGoroutine = 20

	var wg sync.WaitGroup

	// Multiple goroutines adding records concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < recordsPerGoroutine; j++ {
				id := int64(offset*recordsPerGoroutine + j + 1)
				store.Add(id, Data{"goroutine": offset, "index": j})
			}
		}(i)
	}

	wg.Wait()

	// Verify all records were added
	expectedCount := numGoroutines * recordsPerGoroutine
	if store.Len() != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, store.Len())
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.GetLatest(10)
			_ = store.GetSince(100)
			_ = store.GetAll()
		}()
	}

	wg.Wait()
}

func TestStore_DefaultMaxRecords(t *testing.T) {
	// Test with invalid maxRecords
	store := NewStore(0)

	// Should use default value (1000)
	for i := int64(1); i <= 1001; i++ {
		store.Add(i, Data{"index": i})
	}

	// Should have default max (1000) records
	if store.Len() != 1000 {
		t.Errorf("Expected default 1000 records, got %d", store.Len())
	}
}
