package debugmonitor

import "github.com/labstack/echo/v4"

const (
	IconExclamationCircle = `<svg style="width: 16px; height: 16px;" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>`
	IconCircleStack       = `<svg style="width: 16px; height: 16px;" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6"><path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" /></svg>`
)

type TableRowRendererFunc func(c echo.Context, dataEntry *DataEntry) string

type Monitor struct {
	// Name is the name of this monitor.
	// It must be unique among all monitors.
	Name string
	// DisplayName is the display name of this monitor.
	DisplayName string
	// MaxRecords is the maximum number of records to keep in the data storage.
	MaxRecords int
	// Icon is an HTML element string representing the icon for this monitor.
	// Typically, it is an SVG string.
	Icon string
	// TableHeader is an HTML element string representing the table header row for this monitor.
	TableHeader string
	// TableRowRenderer is a function that renders a table row for a given data entry.
	TableRowRenderer TableRowRendererFunc
	// store is the in-memory data store for records.
	store *Store
}

func (m *Monitor) Write(payload any) error {
	if m.store == nil {
		// noop if the store is not initialized
		// It means the monitor is not connected to a Manager
		return nil
	}

	return m.store.Add(payload)
}
