package debugmonitor

type MonitorType string

const (
	MonitorTypeLog MonitorType = "log"
)

type Monitor struct {
	Name string
	Type MonitorType
}
