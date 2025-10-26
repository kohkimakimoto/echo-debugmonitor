package debugmonitor

type Monitor struct {
	// Name is the name of this monitor.
	Name string
	// RowLimit is the maximum number of rows to store for this monitor.
	RowLimit int
}
