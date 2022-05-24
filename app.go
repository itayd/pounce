package main

import (
	"github.com/urfave/cli/v2"
)

var version string // set by linker

var app = cli.App{
	Commands: []*cli.Command{
		&collectCmd,
		&applyCmd,
	},
	UseShortOptionHandling: true,
	EnableBashCompletion:   true,
	Version:                version,
}
