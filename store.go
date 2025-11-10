package debugmonitor

import (
	"container/list"
	"sync"
)

// DataEntry represents a single data record with its ID.
type DataEntry struct {
	Id      int64
	Payload any
}

// Store is an in-memory data store that provides O(1) access by ID
// while maintaining insertion order like a linked hash map.
// It automatically removes old records when the maximum capacity is reached.
// It uses Snowflake-style int64 IDs to guarantee uniqueness and ordering.
type Store struct {
	mu         sync.RWMutex
	maxRecords int
	idGen      *IDGenerator            // Snowflake-style ID generator
	entries    map[int64]*list.Element // map for O(1) access by ID
	order      *list.List              // doubly linked list to maintain insertion order
}

// NewStore creates a new Store with the specified maximum number of records.
// When the limit is reached, the oldest records are automatically removed.
func NewStore(maxRecords int) *Store {
	if maxRecords <= 0 {
		maxRecords = 1000 // Default maximum
	}
	return &Store{
		maxRecords: maxRecords,
		idGen:      NewIDGenerator(),
		entries:    make(map[int64]*list.Element),
		order:      list.New(),
	}
}

// Add adds a new record to the store with a Snowflake-style int64 ID.
// The ID is generated using a time-based algorithm for uniqueness and ordering.
// If the store is at capacity, the oldest record is removed.
func (s *Store) Add(payload any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate Snowflake-style ID
	id := s.idGen.Generate()

	entry := &DataEntry{
		Id:      id,
		Payload: payload,
	}

	// Add to the end of the list for O(1) insertion
	element := s.order.PushBack(entry)

	s.entries[entry.Id] = element

	// Remove the oldest record if we exceed maxRecords
	if s.order.Len() > s.maxRecords {
		oldest := s.order.Front()
		if oldest != nil {
			oldEntry := oldest.Value.(*DataEntry)
			delete(s.entries, oldEntry.Id)
			s.order.Remove(oldest)
		}
	}
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
func (s *Store) GetSince(sinceID int64) []*DataEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*DataEntry, 0)

	var startElement *list.Element
	if sinceID == 0 {
		// Start from the beginning if sinceID is 0
		startElement = s.order.Front()
	} else {
		// Find the element with sinceID and start from the next one
		if element, exists := s.entries[sinceID]; exists {
			startElement = element.Next()
		} else {
			// If sinceID doesn't exist, find the first element with ID > sinceID
			// This handles the case where sinceID was already removed from the store
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
func (s *Store) GetById(id int64) *DataEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if element, exists := s.entries[id]; exists {
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

	s.idGen = NewIDGenerator()
	s.entries = make(map[int64]*list.Element)
	s.order.Init()
}
