package debugmonitor

import (
	"sync"
	"time"
)

const (
	// Bit allocation for 64-bit ID:
	// | 1 bit (sign) | 45 bits (timestamp) | 18 bits (sequence) |
	sequenceBits = 18

	// Maximum values
	maxSequence = (1 << sequenceBits) - 1 // 262,143 (2^18 - 1)

	// Bit shifts
	timestampShift = sequenceBits // 18

	// Custom epoch: 2025-01-01 00:00:00 UTC
	// Using a recent epoch maximizes the usable time range (~1,115 years from this date)
	customEpoch = 1735657200000 // milliseconds
)

// IDGenerator generates unique int64 IDs using a Snowflake-like algorithm.
// The ID structure:
// - 1 bit: sign (always 0 for positive values)
// - 45 bits: timestamp in milliseconds since custom epoch (provides ~1,115 years range)
// - 18 bits: sequence number (allows 262,144 IDs per millisecond)
//
// This provides roughly time-ordered IDs with high throughput capacity.
type IDGenerator struct {
	mu            sync.Mutex
	lastTimestamp int64
	sequence      int64
}

// NewIDGenerator creates a new ID generator.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		lastTimestamp: 0,
		sequence:      0,
	}
}

// Generate generates a new unique int64 ID.
// This method is thread-safe and blocks if called more than maxSequence times
// within the same millisecond, or if the clock moves backwards, waiting for
// the appropriate time to generate a valid ID.
func (g *IDGenerator) Generate() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := g.currentTimestamp()

	// Handle clock moving backwards by waiting until it catches up
	if timestamp < g.lastTimestamp {
		timestamp = g.waitNextMillis(g.lastTimestamp - 1)
	}

	if timestamp == g.lastTimestamp {
		// Same millisecond: increment sequence
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			// Sequence overflow: wait for next millisecond
			timestamp = g.waitNextMillis(timestamp)
		}
	} else {
		// New millisecond: reset sequence to 0
		g.sequence = 0
	}

	g.lastTimestamp = timestamp

	// Construct the ID:
	// | 1 bit (sign=0) | 45 bits (timestamp) | 18 bits (sequence) |
	id := (timestamp << timestampShift) | g.sequence

	return id
}

// currentTimestamp returns the current timestamp in milliseconds since the custom epoch.
func (g *IDGenerator) currentTimestamp() int64 {
	return time.Now().UnixMilli() - customEpoch
}

// waitNextMillis waits until the next millisecond and returns the new timestamp.
func (g *IDGenerator) waitNextMillis(lastTimestamp int64) int64 {
	timestamp := g.currentTimestamp()
	for timestamp <= lastTimestamp {
		time.Sleep(time.Millisecond)
		timestamp = g.currentTimestamp()
	}
	return timestamp
}

// ExtractTimestamp extracts the timestamp component from an ID and returns
// the actual time.Time value.
func ExtractTimestamp(id int64) time.Time {
	timestamp := id >> timestampShift
	unixMillis := timestamp + customEpoch
	return time.UnixMilli(unixMillis)
}

// ExtractSequence extracts the sequence number component from an ID.
func ExtractSequence(id int64) int64 {
	return id & maxSequence
}
