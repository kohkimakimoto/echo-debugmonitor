package debugmonitor

import (
	"container/list"
	"sync"
)

// record represents a single data record with its ID.
// This is an internal type and should not be exposed outside the package.
type record struct {
	id   int64
	data Data
}

// Store is an in-memory data store that provides O(1) access by ID
// while maintaining insertion order like a linked hash map.
// It automatically removes old records when the maximum capacity is reached.
// It generates sequential IDs internally to guarantee sorted order.
type Store struct {
	mu         sync.RWMutex
	maxRecords int
	records    map[int64]*list.Element // map for O(1) access by ID
	order      *list.List              // doubly linked list to maintain insertion order
	idGen      *IDGenerator            // ID generator for sequential IDs
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
		idGen:      NewIDGenerator(),
	}
}

// Add adds a new record to the store with an auto-generated sequential ID.
// The ID is guaranteed to be larger than all existing IDs.
// If the store is at capacity, the oldest record (the lowest ID) is removed.
// Returns the generated ID.
func (s *Store) Add(data Data) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate a new sequential ID
	id := s.idGen.Next()

	rec := &record{
		id:   id,
		data: data,
	}

	// Since IDs are generated sequentially, new IDs are always larger than existing ones.
	// Simply add to the end of the list for O(1) insertion.
	element := s.order.PushBack(rec)

	s.records[id] = element

	// Remove the oldest record if we exceed maxRecords
	if s.order.Len() > s.maxRecords {
		oldest := s.order.Front()
		if oldest != nil {
			oldRecord := oldest.Value.(*record)
			delete(s.records, oldRecord.id)
			s.order.Remove(oldest)
		}
	}

	return id
}

// GetLatest returns the N most recent data entries in reverse chronological order (newest first).
// Each data entry includes the ID (key "id").
// If n is greater than the number of records, all records are returned.
func (s *Store) GetLatest(n int) []Data {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n <= 0 {
		return []Data{}
	}

	count := n
	if count > s.order.Len() {
		count = s.order.Len()
	}

	result := make([]Data, 0, count)
	element := s.order.Back()
	for i := 0; i < count && element != nil; i++ {
		rec := element.Value.(*record)
		data := make(Data, len(rec.data)+1)
		for k, v := range rec.data {
			data[k] = v
		}
		data["id"] = rec.id
		result = append(result, data)
		element = element.Prev()
	}

	return result
}

// GetSince returns all data entries with ID greater than the specified ID,
// in chronological order (oldest first).
// Each data entry includes the ID (key "id").
// This is optimized for cursor-based pagination in log streaming.
// Time complexity: O(m) where m is the number of results.
func (s *Store) GetSince(sinceID int64) []Data {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Data, 0)

	var startElement *list.Element
	if sinceID == 0 {
		// Start from the beginning if sinceID is 0
		startElement = s.order.Front()
	} else {
		// Find the element with sinceID and start from the next one
		if element, exists := s.records[sinceID]; exists {
			startElement = element.Next()
		} else {
			// If sinceID doesn't exist, find the first element with ID > sinceID
			// This handles the case where sinceID was already removed from the store
			for element := s.order.Front(); element != nil; element = element.Next() {
				rec := element.Value.(*record)
				if rec.id > sinceID {
					startElement = element
					break
				}
			}
		}
	}

	// Collect all records from startElement to the end
	for element := startElement; element != nil; element = element.Next() {
		rec := element.Value.(*record)
		data := make(Data, len(rec.data)+1)
		for k, v := range rec.data {
			data[k] = v
		}
		data["id"] = rec.id
		result = append(result, data)
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
