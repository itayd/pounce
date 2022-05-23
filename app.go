package main

import (
	"github.com/urfave/cli/v2"
)

var app = cli.App{
	Commands: []*cli.Command{
		&collectCmd,
		&applyCmd,
	},
}
