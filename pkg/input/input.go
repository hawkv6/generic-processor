package input

const Subsystem = "inputs"

type Input interface {
	Init() error
	Start()
	Stop()
}
