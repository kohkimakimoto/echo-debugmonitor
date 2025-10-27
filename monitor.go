package debugmonitor

type Monitor struct {
	// Name is the name of this monitor.
	Name string
	// DisplayName is the display name of this monitor.
	DisplayName string
	// MaxRecords is the maximum number of records to keep.
	MaxRecords int
}
