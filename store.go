package debugmonitor

type Store interface {
	Write() error
}
