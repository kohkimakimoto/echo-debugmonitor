package debugmonitor

import (
	"sync"
	"testing"
)

func TestIDGenerator_Next(t *testing.T) {
	gen := NewIDGenerator()

	// Test sequential ID generation
	for i := int64(1); i <= 10; i++ {
		id := gen.Next()
		if id != i {
			t.Errorf("Expected ID %d, got %d", i, id)
		}
	}
}

func TestIDGenerator_Concurrency(t *testing.T) {
	gen := NewIDGenerator()
	const numGoroutines = 100
	const idsPerGoroutine = 100

	ids := make(chan int64, numGoroutines*idsPerGoroutine)
	var wg sync.WaitGroup

	// Start multiple goroutines generating IDs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				ids <- gen.Next()
			}
		}()
	}

	wg.Wait()
	close(ids)

	// Collect all IDs and check for uniqueness
	seen := make(map[int64]bool)
	count := 0
	for id := range ids {
		if seen[id] {
			t.Errorf("Duplicate ID found: %d", id)
		}
		seen[id] = true
		count++
	}

	expectedCount := numGoroutines * idsPerGoroutine
	if count != expectedCount {
		t.Errorf("Expected %d IDs, got %d", expectedCount, count)
	}
}
