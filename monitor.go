package debugmonitor

import "github.com/kohkimakimoto/echo-viewkit/pongo2"

type Monitor struct {
	// Name is the name of this monitor.
	// It must be unique among all monitors.
	Name string
	// DisplayName is the display name of this monitor.
	DisplayName string
	// MaxRecords is the maximum number of records to keep in the data storage.
	MaxRecords int
	// Icon
	Icon string

	// store is the in-memory data store for records.
	store *Store
}

func (m *Monitor) Write(payload any) error {
	if m.store == nil {
		// noop if the store is not initialized
		// It means the monitor is not connected to a Manager
		return nil
	}

	// Store the payload directly
	return m.store.Add(payload)
}

// GetLatestData returns the N most recent data entries.
// Each entry includes the ID as key "id".
// This is typically used for the initial display of logs.
func (m *Monitor) GetLatestData(n int) []map[string]any {
	if m.store == nil {
		return []map[string]any{}
	}

	entries := m.store.GetLatest(n)
	return convertEntriesToMaps(entries)
}

// GetDataSince returns all data entries with ID greater than the specified ID.
// Each entry includes the ID as key "id".
// This is optimized for cursor-based pagination in log streaming.
// Pass sinceID="" to get all records from the beginning.
func (m *Monitor) GetDataSince(sinceID string) []map[string]any {
	if m.store == nil {
		return []map[string]any{}
	}

	entries := m.store.GetSince(sinceID)
	return convertEntriesToMaps(entries)
}

// convertEntriesToMaps converts []*DataEntry to []map[string]any for backward compatibility.
func convertEntriesToMaps(entries []*DataEntry) []map[string]any {
	result := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		data := make(map[string]any)

		// If Payload is a map, copy its fields
		if payloadMap, ok := entry.Payload.(map[string]any); ok {
			for k, v := range payloadMap {
				data[k] = v
			}
		} else {
			// If Payload is not a map, store it under "data" key
			data["data"] = entry.Payload
		}

		data["id"] = entry.Id
		result = append(result, data)
	}
	return result
}

const (
	IconExclamationCircle = `<svg style="width: 16px; height: 16px;" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>`
	IconCircleStack       = `<svg style="width: 16px; height: 16px;" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6"><path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" /></svg>`
)

// viewMonitor is a wrapper around Monitor for view rendering.
type viewMonitor struct {
	Monitor *Monitor
}

func newViewMonitor(mo *Monitor) *viewMonitor {
	return &viewMonitor{
		Monitor: mo,
	}
}

func newViewMonitorSlice(mos []*Monitor) []*viewMonitor {
	vms := make([]*viewMonitor, 0, len(mos))
	for _, mo := range mos {
		vms = append(vms, newViewMonitor(mo))
	}
	return vms
}

func (m *viewMonitor) Name() string {
	return m.Monitor.Name
}

func (m *viewMonitor) DisplayName() string {
	return m.Monitor.DisplayName
}

func (m *viewMonitor) Icon() *pongo2.Value {
	return pongo2.AsSafeValue(m.Monitor.Icon)
}
