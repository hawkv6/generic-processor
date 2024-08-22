package processor

const Subsystem = "processors"

type Processor interface {
	Start()
	Stop()
}
