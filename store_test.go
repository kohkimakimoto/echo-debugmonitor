package debugmonitor

import (
	"sync"
	"testing"
)

func TestStore_Add(t *testing.T) {
	store := NewStore(5)

	// Add some records
	for i := 1; i <= 3; i++ {
		err := store.Add(map[string]any{"message": "test", "index": i})
		if err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
	}

	if store.Len() != 3 {
		t.Errorf("Expected 3 records, got %d", store.Len())
	}

	// Get all records and verify IDs
	allData := store.GetSince("")
	if len(allData) != 3 {
		t.Errorf("Expected 3 records, got %d", len(allData))
	}

	// Verify all IDs are unique and valid UUIDs
	seen := make(map[string]bool)
	for _, entry := range allData {
		// UUIDv7 should be a valid UUID string (36 characters)
		if len(entry.Id) != 36 {
			t.Errorf("Expected UUID length 36, got %d", len(entry.Id))
		}
		if seen[entry.Id] {
			t.Errorf("Duplicate ID found: %s", entry.Id)
		}
		seen[entry.Id] = true
	}
}

// TestStore_Get is removed because Get method is no longer needed.
// Use GetSince to retrieve records by ID range.

func TestStore_MaxRecords(t *testing.T) {
	store := NewStore(3)

	// Add 5 records (exceeds limit of 3)
	for i := 1; i <= 5; i++ {
		err := store.Add(map[string]any{"index": i})
		if err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
	}

	// Should only have 3 records
	if store.Len() != 3 {
		t.Errorf("Expected 3 records, got %d", store.Len())
	}

	// Get all records and verify only the newest 3 remain
	allData := store.GetSince("")
	if len(allData) != 3 {
		t.Errorf("Expected 3 records, got %d", len(allData))
	}

	// Verify the last 3 records have the correct index values (3, 4, 5)
	expectedIndexes := []int{3, 4, 5}
	for i, entry := range allData {
		payloadMap := entry.Payload.(map[string]any)
		if payloadMap["index"] != expectedIndexes[i] {
			t.Errorf("Expected index %d at position %d, got %v", expectedIndexes[i], i, payloadMap["index"])
		}
	}
}

func TestStore_GetLatest(t *testing.T) {
	store := NewStore(10)

	// Add records
	for i := 1; i <= 5; i++ {
		err := store.Add(map[string]any{"index": i})
		if err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
	}

	// Get all records first to extract IDs
	allData := store.GetSince("")
	var ids []string
	for _, entry := range allData {
		ids = append(ids, entry.Id)
	}

	// Get latest 3 records
	latest := store.GetLatest(3)
	if len(latest) != 3 {
		t.Errorf("Expected 3 records, got %d", len(latest))
	}

	// Should be in reverse chronological order (newest first)
	expectedIDs := []string{ids[4], ids[3], ids[2]}
	for i, entry := range latest {
		if entry.Id != expectedIDs[i] {
			t.Errorf("Expected ID %s at position %d, got %v", expectedIDs[i], i, entry.Id)
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
	for i := 1; i <= 5; i++ {
		err := store.Add(map[string]any{"index": i})
		if err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
	}

	// Get all records first to extract IDs
	allData := store.GetSince("")
	var ids []string
	for _, entry := range allData {
		ids = append(ids, entry.Id)
	}

	// Get records since ID 2 (third, fourth, and fifth records)
	since := store.GetSince(ids[1])
	if len(since) != 3 {
		t.Errorf("Expected 3 records, got %d", len(since))
	}

	// Should be in chronological order
	expectedIDs := []string{ids[2], ids[3], ids[4]}
	for i, entry := range since {
		if entry.Id != expectedIDs[i] {
			t.Errorf("Expected ID %s at position %d, got %v", expectedIDs[i], i, entry.Id)
		}
	}

	// Get since empty string (all records)
	since = store.GetSince("")
	if len(since) != 5 {
		t.Errorf("Expected 5 records, got %d", len(since))
	}

	// Get since last ID (no records)
	since = store.GetSince(ids[4])
	if len(since) != 0 {
		t.Errorf("Expected 0 records, got %d", len(since))
	}
}

func TestStore_GetSince_WithRemovedID(t *testing.T) {
	store := NewStore(3)

	// Add 5 records and capture IDs after each addition
	var ids []string
	for i := 1; i <= 5; i++ {
		err := store.Add(map[string]any{"index": i})
		if err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
		// Get all records to capture the latest ID
		allData := store.GetSince("")
		if len(allData) > 0 {
			// Get the last added ID
			lastID := allData[len(allData)-1].Id
			ids = append(ids, lastID)
		}
	}

	// GetSince with an ID that was removed (first ID)
	// Should return all records that exist with ID > first ID (which are the last 3)
	since := store.GetSince(ids[0])
	if len(since) != 3 {
		t.Errorf("Expected 3 records, got %d", len(since))
	}

	expectedIDs := []string{ids[2], ids[3], ids[4]}
	for i, entry := range since {
		if entry.Id != expectedIDs[i] {
			t.Errorf("Expected ID %s at position %d, got %v", expectedIDs[i], i, entry.Id)
		}
	}

	// GetSince with an ID that exists (third ID)
	// Should return the last 2 records
	since = store.GetSince(ids[2])
	if len(since) != 2 {
		t.Errorf("Expected 2 records, got %d", len(since))
	}

	expectedIDs = []string{ids[3], ids[4]}
	for i, entry := range since {
		if entry.Id != expectedIDs[i] {
			t.Errorf("Expected ID %s at position %d, got %v", expectedIDs[i], i, entry.Id)
		}
	}
}

// TestStore_GetAll is removed because GetAll method is no longer needed.
// Use GetSince("") to get all records instead.
func TestStore_GetAll_ViaGetSince(t *testing.T) {
	store := NewStore(10)

	// Add records
	for i := 1; i <= 5; i++ {
		err := store.Add(map[string]any{"index": i})
		if err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
	}

	// GetSince("") should return all records
	all := store.GetSince("")
	if len(all) != 5 {
		t.Errorf("Expected 5 records, got %d", len(all))
	}

	// Verify all records have valid IDs and are in chronological order
	var prevID string
	for i, entry := range all {
		if len(entry.Id) != 36 {
			t.Errorf("Expected UUID length 36 at position %d, got %d", i, len(entry.Id))
		}
		// UUIDv7 is time-ordered, so IDs should be in ascending order
		if i > 0 && entry.Id <= prevID {
			t.Errorf("IDs not in chronological order at position %d", i)
		}
		prevID = entry.Id
	}
}

func TestStore_Clear(t *testing.T) {
	store := NewStore(10)

	// Add records
	for i := 1; i <= 5; i++ {
		if err := store.Add(map[string]any{"index": i}); err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
	}

	if store.Len() != 5 {
		t.Errorf("Expected 5 records before clear, got %d", store.Len())
	}

	// Clear the store
	store.Clear()

	if store.Len() != 0 {
		t.Errorf("Expected 0 records after clear, got %d", store.Len())
	}

	// Verify records are actually gone by checking GetSince("")
	allData := store.GetSince("")
	if len(allData) != 0 {
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
				if err := store.Add(map[string]any{"goroutine": offset, "index": j}); err != nil {
					t.Errorf("Failed to add data: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all records were added
	expectedCount := numGoroutines * recordsPerGoroutine
	if store.Len() != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, store.Len())
	}

	// Verify all IDs are unique
	allData := store.GetSince("")
	seen := make(map[string]bool)
	for _, entry := range allData {
		if seen[entry.Id] {
			t.Errorf("Duplicate ID found: %s", entry.Id)
		}
		seen[entry.Id] = true
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.GetLatest(10)
			latest := store.GetLatest(1)
			if len(latest) > 0 {
				_ = store.GetSince(latest[0].Id)
			}
			_ = store.GetSince("")
		}()
	}

	wg.Wait()
}

func TestStore_DefaultMaxRecords(t *testing.T) {
	// Test with invalid maxRecords
	store := NewStore(0)

	// Should use default value (1000)
	for i := 1; i <= 1001; i++ {
		if err := store.Add(map[string]any{"index": i}); err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
	}

	// Should have default max (1000) records
	if store.Len() != 1000 {
		t.Errorf("Expected default 1000 records, got %d", store.Len())
	}
}
