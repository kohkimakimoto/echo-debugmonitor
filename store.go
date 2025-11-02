package debugmonitor

import (
	"container/list"
	"sync"
)

// Record represents a single data record with its ID.
type Record struct {
	ID   int64
	Data Data
}

// Store is an in-memory data store that provides O(1) access by ID
// while maintaining insertion order like a linked hash map.
// It automatically removes old records when the maximum capacity is reached.
type Store struct {
	mu         sync.RWMutex
	maxRecords int
	records    map[int64]*list.Element // map for O(1) access by ID
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
		records:    make(map[int64]*list.Element),
		order:      list.New(),
	}
}

// Add adds a new record to the store in sorted order by ID.
// If the store is at capacity, the oldest record (the lowest ID) is removed.
func (s *Store) Add(id int64, data Data) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := &Record{
		ID:   id,
		Data: data,
	}

	// Find the correct position to insert (maintaining sorted order by ID)
	var element *list.Element
	if s.order.Len() == 0 {
		// Empty list, just add at the end
		element = s.order.PushBack(record)
	} else {
		// Find insertion point - insert before the first record with the larger ID
		inserted := false
		for e := s.order.Front(); e != nil; e = e.Next() {
			existingRecord := e.Value.(*Record)
			if existingRecord.ID > id {
				element = s.order.InsertBefore(record, e)
				inserted = true
				break
			}
		}
		if !inserted {
			// ID is larger than all existing IDs, add at the end
			element = s.order.PushBack(record)
		}
	}

	s.records[id] = element

	// Remove the oldest record if we exceed maxRecords
	if s.order.Len() > s.maxRecords {
		oldest := s.order.Front()
		if oldest != nil {
			oldRecord := oldest.Value.(*Record)
			delete(s.records, oldRecord.ID)
			s.order.Remove(oldest)
		}
	}
}

// Get retrieves a record by its ID.
// Returns the record and true if found, nil and false otherwise.
func (s *Store) Get(id int64) (*Record, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	element, exists := s.records[id]
	if !exists {
		return nil, false
	}
	return element.Value.(*Record), true
}

// GetLatest returns the N most recent records in reverse chronological order (newest first).
// If n is greater than the number of records, all records are returned.
func (s *Store) GetLatest(n int) []*Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n <= 0 {
		return []*Record{}
	}

	count := n
	if count > s.order.Len() {
		count = s.order.Len()
	}

	result := make([]*Record, 0, count)
	element := s.order.Back()
	for i := 0; i < count && element != nil; i++ {
		result = append(result, element.Value.(*Record))
		element = element.Prev()
	}

	return result
}

// GetSince returns all records with ID greater than the specified ID,
// in chronological order (oldest first).
// This is useful for getting new records since a specific point.
func (s *Store) GetSince(sinceID int64) []*Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Record, 0)
	for element := s.order.Front(); element != nil; element = element.Next() {
		record := element.Value.(*Record)
		if record.ID > sinceID {
			result = append(result, record)
		}
	}

	return result
}

// GetAll returns all records in chronological order (oldest first).
func (s *Store) GetAll() []*Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Record, 0, s.order.Len())
	for element := s.order.Front(); element != nil; element = element.Next() {
		result = append(result, element.Value.(*Record))
	}

	return result
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

	s.records = make(map[int64]*list.Element)
	s.order.Init()
}
