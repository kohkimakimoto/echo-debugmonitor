package debugmonitor

import "sync"

// IDGenerator generates sequential IDs similar to MySQL's AUTO_INCREMENT.
// It's designed for simple use cases and doesn't focus on extreme optimization.
type IDGenerator struct {
	mu      sync.Mutex
	current int64
}

// NewIDGenerator creates a new ID generator starting from 1.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		current: 0,
	}
}

// Next returns the next sequential ID.
// It's thread-safe and guarantees unique IDs.
func (g *IDGenerator) Next() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.current++
	return g.current
}
