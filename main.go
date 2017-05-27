package main

import (
	"github.com/saiki/go-reveal/cmd"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "go-reveal"
	app.Commands = cmd.GetCommands()
	app.Flags = cmd.Flags
	app.Run(os.Args)
}
