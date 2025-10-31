package debugmonitor

type Data map[string]any

type Monitor struct {
	// Name is the name of this monitor.
	// It must be unique among all monitors.
	Name string
	// DisplayName is the display name of this monitor.
	DisplayName string
	// MaxRecords is the maximum number of records to keep in the data storage.
	MaxRecords int
	// ChannelBufferSize is the size of the buffered channel for data communication.
	ChannelBufferSize int
	// Icon
	Icon string
	// dataChan is the channel for sending data to the Manager.
	dataChan chan Data
}

func (m *Monitor) Write(data Data) error {
	// Send data to the channel if it's initialized
	// Use a goroutine to prevent blocking the caller
	if m.dataChan != nil {
		go func() {
			m.dataChan <- data
		}()
	}
	return nil
}
