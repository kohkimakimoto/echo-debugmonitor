package debugmonitor

import (
	"sync"
	"testing"
	"time"
)

func TestStore_Add(t *testing.T) {
	store := NewStore(5)

	// Add some records
	for i := 1; i <= 3; i++ {
		store.Add(map[string]any{"message": "test", "index": i})
	}

	if store.Len() != 3 {
		t.Errorf("Expected 3 records, got %d", store.Len())
	}

	// Get all records and verify IDs
	allData := store.GetSince(0)
	if len(allData) != 3 {
		t.Errorf("Expected 3 records, got %d", len(allData))
	}

	// Verify all IDs are unique and in ascending order (Snowflake IDs)
	seen := make(map[int64]bool)
	var prevID int64 = 0
	for _, entry := range allData {
		// IDs should be positive
		if entry.Id <= 0 {
			t.Errorf("Expected positive ID, got %d", entry.Id)
		}
		// IDs should be unique
		if seen[entry.Id] {
			t.Errorf("Duplicate ID found: %d", entry.Id)
		}
		// IDs should be in ascending order
		if entry.Id <= prevID {
			t.Errorf("IDs not in ascending order: prev=%d, current=%d", prevID, entry.Id)
		}
		seen[entry.Id] = true
		prevID = entry.Id
	}
}

// TestStore_Get is removed because Get method is no longer needed.
// Use GetSince to retrieve records by ID range.

func TestStore_MaxRecords(t *testing.T) {
	store := NewStore(3)

	// Add 5 records (exceeds limit of 3)
	for i := 1; i <= 5; i++ {
		store.Add(map[string]any{"index": i})
	}

	// Should only have 3 records
	if store.Len() != 3 {
		t.Errorf("Expected 3 records, got %d", store.Len())
	}

	// Get all records and verify only the newest 3 remain
	allData := store.GetSince(0)
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
		store.Add(map[string]any{"index": i})
	}

	// Get all records
	all := store.GetLatest()
	if len(all) != 5 {
		t.Errorf("Expected 5 records, got %d", len(all))
	}

	// Get all records first to extract IDs for verification
	allData := store.GetSince(0)
	var ids []int64
	for _, entry := range allData {
		ids = append(ids, entry.Id)
	}

	// Should be in reverse chronological order (newest first)
	expectedIDs := []int64{ids[4], ids[3], ids[2], ids[1], ids[0]}
	for i, entry := range all {
		if entry.Id != expectedIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %v", expectedIDs[i], i, entry.Id)
		}
	}

	// Test GetLatestWithLimit
	latest := store.GetLatestWithLimit(3)
	if len(latest) != 3 {
		t.Errorf("Expected 3 records, got %d", len(latest))
	}

	// Should be in reverse chronological order (newest first)
	expectedIDsLimited := []int64{ids[4], ids[3], ids[2]}
	for i, entry := range latest {
		if entry.Id != expectedIDsLimited[i] {
			t.Errorf("Expected ID %d at position %d, got %v", expectedIDsLimited[i], i, entry.Id)
		}
	}

	// Request more than available
	allWithLimit := store.GetLatestWithLimit(100)
	if len(allWithLimit) != 5 {
		t.Errorf("Expected 5 records, got %d", len(allWithLimit))
	}
}

func TestStore_GetSince(t *testing.T) {
	store := NewStore(10)

	// Add records
	for i := 1; i <= 5; i++ {
		store.Add(map[string]any{"index": i})
	}

	// Get all records first to extract IDs
	allData := store.GetSince(0)
	var ids []int64
	for _, entry := range allData {
		ids = append(ids, entry.Id)
	}

	// Get records since ID 2 (third, fourth, and fifth records)
	since := store.GetSince(ids[1])
	if len(since) != 3 {
		t.Errorf("Expected 3 records, got %d", len(since))
	}

	// Should be in chronological order
	expectedIDs := []int64{ids[2], ids[3], ids[4]}
	for i, entry := range since {
		if entry.Id != expectedIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %v", expectedIDs[i], i, entry.Id)
		}
	}

	// Get since 0 (all records)
	since = store.GetSince(0)
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
	var ids []int64
	for i := 1; i <= 5; i++ {
		store.Add(map[string]any{"index": i})
		// Get all records to capture the latest ID
		allData := store.GetSince(0)
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

	expectedIDs := []int64{ids[2], ids[3], ids[4]}
	for i, entry := range since {
		if entry.Id != expectedIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %v", expectedIDs[i], i, entry.Id)
		}
	}

	// GetSince with an ID that exists (third ID)
	// Should return the last 2 records
	since = store.GetSince(ids[2])
	if len(since) != 2 {
		t.Errorf("Expected 2 records, got %d", len(since))
	}

	expectedIDs = []int64{ids[3], ids[4]}
	for i, entry := range since {
		if entry.Id != expectedIDs[i] {
			t.Errorf("Expected ID %d at position %d, got %v", expectedIDs[i], i, entry.Id)
		}
	}
}

// TestStore_GetAll is removed because GetAll method is no longer needed.
// Use GetSince(0) to get all records instead.
func TestStore_GetAll_ViaGetSince(t *testing.T) {
	store := NewStore(10)

	// Add records
	for i := 1; i <= 5; i++ {
		store.Add(map[string]any{"index": i})
	}

	// GetSince(0) should return all records
	all := store.GetSince(0)
	if len(all) != 5 {
		t.Errorf("Expected 5 records, got %d", len(all))
	}

	// Verify all records have valid Snowflake IDs and are in chronological order
	var prevID int64 = 0
	for i, entry := range all {
		// IDs should be positive
		if entry.Id <= 0 {
			t.Errorf("Expected positive ID at position %d, got %d", i, entry.Id)
		}
		// IDs should be in ascending order
		if entry.Id <= prevID {
			t.Errorf("IDs not in chronological order at position %d: prev=%d, current=%d", i, prevID, entry.Id)
		}
		prevID = entry.Id
	}
}

func TestStore_Clear(t *testing.T) {
	store := NewStore(10)

	// Add records
	for i := 1; i <= 5; i++ {
		store.Add(map[string]any{"index": i})
	}

	if store.Len() != 5 {
		t.Errorf("Expected 5 records before clear, got %d", store.Len())
	}

	// Clear the store
	store.Clear()

	if store.Len() != 0 {
		t.Errorf("Expected 0 records after clear, got %d", store.Len())
	}

	// Verify records are actually gone by checking GetSince(0)
	allData := store.GetSince(0)
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
				store.Add(map[string]any{"goroutine": offset, "index": j})
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
	allData := store.GetSince(0)
	seen := make(map[int64]bool)
	for _, entry := range allData {
		if seen[entry.Id] {
			t.Errorf("Duplicate ID found: %d", entry.Id)
		}
		seen[entry.Id] = true
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.GetLatestWithLimit(10)
			latest := store.GetLatestWithLimit(1)
			if len(latest) > 0 {
				_ = store.GetSince(latest[0].Id)
			}
			_ = store.GetSince(0)
		}()
	}

	wg.Wait()
}

func TestStore_DefaultMaxRecords(t *testing.T) {
	// Test with invalid maxRecords
	store := NewStore(0)

	// Should use default value (1000)
	for i := 1; i <= 1001; i++ {
		store.Add(map[string]any{"index": i})
	}

	// Should have default max (1000) records
	if store.Len() != 1000 {
		t.Errorf("Expected default 1000 records, got %d", store.Len())
	}
}

func TestStore_NewAddEvent(t *testing.T) {
	store := NewStore(10)

	// Create an Add event subscription
	event := store.NewAddEvent()
	defer event.Close()

	// Add a record
	testPayload := map[string]any{"message": "test notification"}
	store.Add(testPayload)

	// Wait for notification
	select {
	case entry := <-event.C:
		if entry.Payload.(map[string]any)["message"] != "test notification" {
			t.Errorf("Expected notification with 'test notification', got %v", entry.Payload)
		}
		if entry.Id <= 0 {
			t.Errorf("Expected positive ID in notification, got %d", entry.Id)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for notification")
	}
}

func TestStore_MultipleAddSubscribers(t *testing.T) {
	store := NewStore(10)

	// Create multiple event subscriptions
	const numSubscribers = 3
	events := make([]*AddEvent, numSubscribers)

	for i := 0; i < numSubscribers; i++ {
		events[i] = store.NewAddEvent()
		defer events[i].Close()
	}

	// Add a record
	testPayload := map[string]any{"index": 1}
	store.Add(testPayload)

	// All subscribers should receive the notification
	for i := 0; i < numSubscribers; i++ {
		select {
		case entry := <-events[i].C:
			if entry.Payload.(map[string]any)["index"] != 1 {
				t.Errorf("Subscriber %d: Expected index 1, got %v", i, entry.Payload)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Subscriber %d did not receive notification", i)
		}
	}
}

func TestStore_NewClearEvent(t *testing.T) {
	store := NewStore(10)

	// Create event subscriptions
	addEvent := store.NewAddEvent()
	defer addEvent.Close()

	clearEvent := store.NewClearEvent()
	defer clearEvent.Close()

	// Add a record first
	store.Add(map[string]any{"message": "test"})

	// Wait for Add event
	select {
	case entry := <-addEvent.C:
		if entry.Payload.(map[string]any)["message"] != "test" {
			t.Errorf("Expected 'test', got %v", entry.Payload)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for Add event")
	}

	// Clear the store
	store.Clear()

	// Wait for Clear event
	select {
	case <-clearEvent.C:
		// Expected - clear event received
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for Clear event")
	}

	// Verify store is actually cleared
	if store.Len() != 0 {
		t.Errorf("Expected store length 0 after clear, got %d", store.Len())
	}
}

func TestStore_ConcurrentSubscriptions(t *testing.T) {
	store := NewStore(100)

	const numSubscribers = 10
	const numRecords = 20

	var wg sync.WaitGroup

	// Create multiple concurrent add event subscriptions
	events := make([]*AddEvent, numSubscribers)
	allReceived := make([]chan bool, numSubscribers)

	for i := 0; i < numSubscribers; i++ {
		events[i] = store.NewAddEvent()
		defer events[i].Close()

		allReceived[i] = make(chan bool, 1)
		wg.Add(1)
		idx := i
		go func(event *AddEvent, done chan bool) {
			defer wg.Done()
			count := 0
			for range event.C {
				count++
				if count == numRecords {
					done <- true
				}
			}
		}(events[idx], allReceived[idx])
	}

	// Add records
	for i := 0; i < numRecords; i++ {
		store.Add(map[string]any{"index": i})
	}

	// Wait for all subscribers to receive all notifications
	for i := 0; i < numSubscribers; i++ {
		select {
		case <-allReceived[i]:
			// Subscriber received all notifications
		case <-time.After(2 * time.Second):
			t.Errorf("Subscriber %d did not receive all notifications in time", i)
		}
	}

	// Close all events to stop goroutines
	for _, event := range events {
		event.Close()
	}

	wg.Wait()
}

func TestStore_EventClose(t *testing.T) {
	store := NewStore(10)

	// Create an event and close it
	event := store.NewAddEvent()
	event.Close()

	// Add a record - closed event should not receive it
	store.Add(map[string]any{"message": "test"})

	// Channel should be closed
	select {
	case _, ok := <-event.C:
		if ok {
			t.Error("Expected channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected channel to be closed immediately")
	}

	// Calling Close again should be safe
	event.Close()
}
