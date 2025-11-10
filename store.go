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

// AddEvent represents a subscription to Add events.
// Use the C channel to receive notifications when new data is added.
// Call Close() when done to clean up resources.
type AddEvent struct {
	C      <-chan *DataEntry // Channel to receive Add events
	store  *Store
	ch     chan *DataEntry
	closed bool
	mu     sync.Mutex
}

// Close unsubscribes from the Store and closes the event channel.
// After calling Close, the C channel will be closed and no more events will be received.
func (e *AddEvent) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return
	}
	e.closed = true

	e.store.unsubscribeAdd(e)
	close(e.ch)
}

// ClearEvent represents a subscription to Clear events.
// Use the C channel to receive notifications when the store is cleared.
// Call Close() when done to clean up resources.
type ClearEvent struct {
	C      <-chan struct{} // Channel to receive Clear events
	store  *Store
	ch     chan struct{}
	closed bool
	mu     sync.Mutex
}

// Close unsubscribes from the Store and closes the event channel.
// After calling Close, the C channel will be closed and no more events will be received.
func (e *ClearEvent) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return
	}
	e.closed = true

	e.store.unsubscribeClear(e)
	close(e.ch)
}

// Store is an in-memory data store that provides O(1) access by ID
// while maintaining insertion order like a linked hash map.
// It automatically removes old records when the maximum capacity is reached.
// It uses Snowflake-style int64 IDs to guarantee uniqueness and ordering.
// Store supports channel-based event subscriptions for Add and Clear events.
type Store struct {
	mu               sync.RWMutex
	maxRecords       int
	idGen            *IDGenerator            // Snowflake-style ID generator
	entries          map[int64]*list.Element // map for O(1) access by ID
	order            *list.List              // doubly linked list to maintain insertion order
	addEventsMu      sync.RWMutex            // protects addEvents slice
	addEvents        []*AddEvent             // active Add event subscriptions
	clearEventsMu    sync.RWMutex            // protects clearEvents slice
	clearEvents      []*ClearEvent           // active Clear event subscriptions
}

// NewStore creates a new Store with the specified maximum number of records.
// When the limit is reached, the oldest records are automatically removed.
func NewStore(maxRecords int) *Store {
	if maxRecords <= 0 {
		maxRecords = 1000 // Default maximum
	}
	return &Store{
		maxRecords:  maxRecords,
		idGen:       NewIDGenerator(),
		entries:     make(map[int64]*list.Element),
		order:       list.New(),
		addEvents:   make([]*AddEvent, 0),
		clearEvents: make([]*ClearEvent, 0),
	}
}

// Add adds a new record to the store with a Snowflake-style int64 ID.
// The ID is generated using a time-based algorithm for uniqueness and ordering.
// If the store is at capacity, the oldest record is removed.
// After adding, all registered listeners are notified with the new entry.
func (s *Store) Add(payload any) {
	s.mu.Lock()

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

	s.mu.Unlock()

	// Notify add event subscribers outside the lock to prevent deadlocks
	s.notifyAddEvents(entry)
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
// After clearing, all registered clear listeners are notified.
func (s *Store) Clear() {
	s.mu.Lock()

	s.idGen = NewIDGenerator()
	s.entries = make(map[int64]*list.Element)
	s.order.Init()

	s.mu.Unlock()

	// Notify clear event subscribers outside the lock to prevent deadlocks
	s.notifyClearEvents()
}

// NewAddEvent creates a new subscription to Add events.
// The returned AddEvent provides a channel that will receive notifications
// when new data is added to the Store.
// Call Close() on the returned AddEvent when done to clean up resources.
func (s *Store) NewAddEvent() *AddEvent {
	ch := make(chan *DataEntry, 10) // Buffered to prevent blocking
	event := &AddEvent{
		C:     ch,
		store: s,
		ch:    ch,
	}

	s.addEventsMu.Lock()
	s.addEvents = append(s.addEvents, event)
	s.addEventsMu.Unlock()

	return event
}

// NewClearEvent creates a new subscription to Clear events.
// The returned ClearEvent provides a channel that will receive notifications
// when the Store is cleared.
// Call Close() on the returned ClearEvent when done to clean up resources.
func (s *Store) NewClearEvent() *ClearEvent {
	ch := make(chan struct{}, 1) // Buffered to prevent blocking
	event := &ClearEvent{
		C:     ch,
		store: s,
		ch:    ch,
	}

	s.clearEventsMu.Lock()
	s.clearEvents = append(s.clearEvents, event)
	s.clearEventsMu.Unlock()

	return event
}

// unsubscribeAdd removes an AddEvent from the active subscriptions.
func (s *Store) unsubscribeAdd(event *AddEvent) {
	s.addEventsMu.Lock()
	defer s.addEventsMu.Unlock()

	for i, e := range s.addEvents {
		if e == event {
			s.addEvents = append(s.addEvents[:i], s.addEvents[i+1:]...)
			break
		}
	}
}

// unsubscribeClear removes a ClearEvent from the active subscriptions.
func (s *Store) unsubscribeClear(event *ClearEvent) {
	s.clearEventsMu.Lock()
	defer s.clearEventsMu.Unlock()

	for i, e := range s.clearEvents {
		if e == event {
			s.clearEvents = append(s.clearEvents[:i], s.clearEvents[i+1:]...)
			break
		}
	}
}

// notifyAddEvents sends notifications to all active Add event subscribers.
// Non-blocking sends are used to prevent slow consumers from blocking the Store.
func (s *Store) notifyAddEvents(entry *DataEntry) {
	s.addEventsMu.RLock()
	defer s.addEventsMu.RUnlock()

	for _, event := range s.addEvents {
		select {
		case event.ch <- entry:
		default:
			// Channel is full, skip this subscriber to avoid blocking
		}
	}
}

// notifyClearEvents sends notifications to all active Clear event subscribers.
// Non-blocking sends are used to prevent slow consumers from blocking the Store.
func (s *Store) notifyClearEvents() {
	s.clearEventsMu.RLock()
	defer s.clearEventsMu.RUnlock()

	for _, event := range s.clearEvents {
		select {
		case event.ch <- struct{}{}:
		default:
			// Channel is full, skip this subscriber to avoid blocking
		}
	}
}
