package message

type Command interface{ isCommand() }
type BaseCommand struct{}

func (BaseCommand) isCommand() {}

type StartListeningCommand struct{ BaseCommand }

type InfluxCommand struct {
	BaseCommand
	Measurement string
	Field       string
	Method      string
	GroupBy     []string
	Interval    uint
}
