package debugmonitor

import (
	"sync"
	"testing"
	"time"
)

func TestIDGenerator_Generate(t *testing.T) {
	gen := NewIDGenerator()

	// Generate a single ID
	id := gen.Generate()

	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}
}

func TestIDGenerator_UniqueIDs(t *testing.T) {
	gen := NewIDGenerator()
	count := 10000
	ids := make(map[int64]bool, count)

	// Generate multiple IDs and check uniqueness
	for i := 0; i < count; i++ {
		id := gen.Generate()

		if ids[id] {
			t.Errorf("Duplicate ID generated: %d", id)
		}
		ids[id] = true
	}

	if len(ids) != count {
		t.Errorf("Expected %d unique IDs, got %d", count, len(ids))
	}
}

func TestIDGenerator_Ordering(t *testing.T) {
	gen := NewIDGenerator()
	prevID := int64(0)

	// Generate IDs and verify they are increasing
	for i := 0; i < 1000; i++ {
		id := gen.Generate()

		if id <= prevID {
			t.Errorf("IDs not in ascending order: prev=%d, current=%d", prevID, id)
		}
		prevID = id
	}
}

func TestIDGenerator_Concurrent(t *testing.T) {
	gen := NewIDGenerator()
	count := 10000
	goroutines := 10
	idsPerGoroutine := count / goroutines

	mu := sync.Mutex{}
	ids := make(map[int64]bool, count)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Generate IDs concurrently
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < idsPerGoroutine; i++ {
				id := gen.Generate()

				mu.Lock()
				if ids[id] {
					t.Errorf("Duplicate ID generated in concurrent test: %d", id)
				}
				ids[id] = true
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(ids) != count {
		t.Errorf("Expected %d unique IDs, got %d", count, len(ids))
	}
}

func TestExtractTimestamp(t *testing.T) {
	gen := NewIDGenerator()
	beforeGen := time.Now()

	id := gen.Generate()

	afterGen := time.Now()
	extractedTime := ExtractTimestamp(id)

	// The extracted time should be close to the generation time (within 2ms margin)
	// We use a margin because the timestamp is stored at millisecond precision
	margin := 2 * time.Millisecond
	if extractedTime.Before(beforeGen.Add(-margin)) || extractedTime.After(afterGen.Add(margin)) {
		t.Errorf("Extracted timestamp %v is not within expected range [%v, %v]",
			extractedTime, beforeGen.Add(-margin), afterGen.Add(margin))
	}
}

func TestExtractSequence(t *testing.T) {
	gen := NewIDGenerator()

	// Generate multiple IDs in quick succession (likely same millisecond)
	ids := make([]int64, 100)
	for i := 0; i < 100; i++ {
		id := gen.Generate()
		ids[i] = id
	}

	// Check that sequence numbers are valid
	for i, id := range ids {
		seq := ExtractSequence(id)
		if seq < 0 || seq > maxSequence {
			t.Errorf("ID %d has invalid sequence number: %d (should be 0-%d)",
				id, seq, maxSequence)
		}

		// If we have at least 2 IDs with the same timestamp,
		// their sequences should be different
		if i > 0 {
			prevTimestamp := ExtractTimestamp(ids[i-1])
			currTimestamp := ExtractTimestamp(id)

			// If timestamps are the same (within 1ms), sequences should differ
			if currTimestamp.Sub(prevTimestamp).Abs() < time.Millisecond {
				prevSeq := ExtractSequence(ids[i-1])
				currSeq := ExtractSequence(id)
				if prevSeq >= currSeq {
					t.Errorf("Sequences not incrementing properly within same ms: prev=%d, curr=%d",
						prevSeq, currSeq)
				}
			}
		}
	}
}

func TestIDGenerator_BitStructure(t *testing.T) {
	gen := NewIDGenerator()

	id := gen.Generate()

	// Verify that the ID is positive (sign bit is 0)
	if id < 0 {
		t.Errorf("Generated ID is negative: %d", id)
	}

	// Extract and verify components
	timestamp := id >> timestampShift
	sequence := id & maxSequence

	// Timestamp should be positive and reasonable
	if timestamp < 0 {
		t.Errorf("Extracted timestamp is negative: %d", timestamp)
	}

	// Sequence should be within valid range
	if sequence < 0 || sequence > maxSequence {
		t.Errorf("Sequence out of range: %d (max: %d)", sequence, maxSequence)
	}

	// Reconstruct the ID from components
	reconstructed := (timestamp << timestampShift) | sequence
	if reconstructed != id {
		t.Errorf("Failed to reconstruct ID: original=%d, reconstructed=%d",
			id, reconstructed)
	}
}

func BenchmarkIDGenerator_Generate(b *testing.B) {
	gen := NewIDGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Generate()
	}
}

func BenchmarkIDGenerator_GenerateParallel(b *testing.B) {
	gen := NewIDGenerator()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = gen.Generate()
		}
	})
}
