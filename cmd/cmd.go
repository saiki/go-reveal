package cmd

import (
	"github.com/urfave/cli"
)

var commands []cli.Command
var Flags []cli.Flag

func init() {
	commands = make([]cli.Command, 0)
	Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "root, r",
			Value: ".",
		},
	}
}

func GetCommands() []cli.Command {
	return commands
}

func addCommand(cmd cli.Command) {
	commands = append(commands, cmd)
}
