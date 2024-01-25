package output

const Subsystem = "output"

type Output interface {
	Init() error
	Start()
	Stop()
}
