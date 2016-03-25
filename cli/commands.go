package cli

const LogLevelArgName = "loglevel"

type Command interface{
    Name() string
    ShortDescription() string
    Execute() int
}


type ListCommand struct {
}

func (lc *ListCommand) Execute() int{
    return 0
}

func (lc *ListCommand) Name() string {
    return "LIST"
}

func (lc *ListCommand) ShortDescription() string {
    return "Shows all the available commands you can run from the command line"
}