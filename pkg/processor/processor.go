package processor

const Subsystem = "processors"

type Processor interface {
	Init() error
	Start()
	Stop()
}
