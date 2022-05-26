package main

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

const wdprefix = "# wd="

var trimPathPrefix = os.Getenv("TRIM_CWD_PREFIX")

func trimPath(p string) string {
	return strings.TrimPrefix(p, trimPathPrefix)
}

// ---

var commonConfig struct {
	print, nowd, collect, apply bool
}

var commonFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:        "print",
		Aliases:     []string{"p"},
		Destination: &commonConfig.print,
		Usage:       "print each processed line or file path",
	},
	&cli.BoolFlag{
		Name:        "nowd",
		Destination: &commonConfig.nowd,
		Usage:       "no wd comment",
	},
	&cli.BoolFlag{
		Name:        "collect",
		Aliases:     []string{"c"},
		Destination: &commonConfig.collect,
		Usage:       "collect data as specified by collect flags",
	},
	&cli.BoolFlag{
		Name:        "apply",
		Aliases:     []string{"a"},
		Destination: &commonConfig.apply,
		Usage:       "apply changes as specified by apply flags",
	},
}
