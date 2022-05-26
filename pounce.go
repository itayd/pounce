package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

var version string // set by linker

var app = cli.App{
	UseShortOptionHandling: true,
	EnableBashCompletion:   true,
	Version:                version,
	Flags:                  append(append(commonFlags, collectFlags...), applyFlags...),
	Action: func(c *cli.Context) error {
		if dir := commonConfig.dir; dir != "" {
			if err := os.Chdir(dir); err != nil {
				return fmt.Errorf("chdir: %w", err)
			}
		}

		args := c.Args().Slice()

		if commonConfig.collect && commonConfig.apply ||
			(!commonConfig.collect && !commonConfig.apply) {
			// both
			return both(os.Stderr, args)
		}

		if commonConfig.collect {
			// collect only
			return collect(os.Stdout, os.Stderr, args)
		}

		if commonConfig.apply {
			// apply only
			return apply(os.Stdin)
		}

		return nil
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
