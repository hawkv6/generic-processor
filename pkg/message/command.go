package message

type Command interface{ isCommand() }
type BaseCommand struct{}

func (BaseCommand) isCommand() {}

type StartListeningCommand struct{ BaseCommand }
