package debugmonitor

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// DataEntry represents a single data record with its ID.
type DataEntry struct {
	Id      string
	Payload any
}

// Store is an in-memory data store that provides O(1) access by ID
// while maintaining insertion order like a linked hash map.
// It automatically removes old records when the maximum capacity is reached.
// It generates UUIDv7 IDs internally to guarantee uniqueness and time-based ordering.
type Store struct {
	mu         sync.RWMutex
	maxRecords int
	records    map[string]*list.Element // map for O(1) access by ID
	order      *list.List               // doubly linked list to maintain insertion order
}

// NewStore creates a new Store with the specified maximum number of records.
// When the limit is reached, the oldest records are automatically removed.
func NewStore(maxRecords int) *Store {
	if maxRecords <= 0 {
		maxRecords = 1000 // Default maximum
	}
	return &Store{
		maxRecords: maxRecords,
		records:    make(map[string]*list.Element),
		order:      list.New(),
	}
}

// Add adds a new record to the store with an auto-generated UUIDv7 ID.
// UUIDv7 provides both uniqueness and time-based ordering.
// If the store is at capacity, the oldest record is removed.
// Returns any error that occurred during ID generation.
func (s *Store) Add(payload any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate a new UUIDv7 ID
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}

	entry := &DataEntry{
		Id:      id.String(),
		Payload: payload,
	}

	// Since UUIDv7 IDs are time-ordered, new IDs maintain chronological order.
	// Simply add to the end of the list for O(1) insertion.
	element := s.order.PushBack(entry)

	s.records[entry.Id] = element

	// Remove the oldest record if we exceed maxRecords
	if s.order.Len() > s.maxRecords {
		oldest := s.order.Front()
		if oldest != nil {
			oldEntry := oldest.Value.(*DataEntry)
			delete(s.records, oldEntry.Id)
			s.order.Remove(oldest)
		}
	}

	return nil
}

// GetLatest returns the N most recent data entries in reverse chronological order (newest first).
// If n is greater than the number of records, all records are returned.
func (s *Store) GetLatest(n int) []*DataEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n <= 0 {
		return []*DataEntry{}
	}

	count := n
	if count > s.order.Len() {
		count = s.order.Len()
	}

	result := make([]*DataEntry, 0, count)
	element := s.order.Back()
	for i := 0; i < count && element != nil; i++ {
		entry := element.Value.(*DataEntry)
		result = append(result, entry)
		element = element.Prev()
	}

	return result
}

// GetSince returns all data entries with ID greater than the specified ID,
// in chronological order (oldest first).
// This is optimized for cursor-based pagination in log streaming.
// Time complexity: O(m) where m is the number of results.
func (s *Store) GetSince(sinceID string) []*DataEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*DataEntry, 0)

	var startElement *list.Element
	if sinceID == "" {
		// Start from the beginning if sinceID is empty
		startElement = s.order.Front()
	} else {
		// Find the element with sinceID and start from the next one
		if element, exists := s.records[sinceID]; exists {
			startElement = element.Next()
		} else {
			// If sinceID doesn't exist, find the first element with ID > sinceID
			// This handles the case where sinceID was already removed from the store
			// Since UUIDv7 is time-ordered, we can use string comparison
			for element := s.order.Front(); element != nil; element = element.Next() {
				entry := element.Value.(*DataEntry)
				if entry.Id > sinceID {
					startElement = element
					break
				}
			}
		}
	}

	// Collect all records from startElement to the end
	for element := startElement; element != nil; element = element.Next() {
		entry := element.Value.(*DataEntry)
		result = append(result, entry)
	}

	return result
}

// GetById returns a single data entry by its ID.
// Returns nil if the entry is not found.
// Time complexity: O(1).
func (s *Store) GetById(id string) *DataEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if element, exists := s.records[id]; exists {
		return element.Value.(*DataEntry)
	}
	return nil
}

// Len returns the current number of records in the store.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.order.Len()
}

// Clear removes all records from the store.
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = make(map[string]*list.Element)
	s.order.Init()
}
