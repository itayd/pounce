package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var version string // set by linker

var app = cli.App{
	UseShortOptionHandling: true,
	EnableBashCompletion:   true,
	Version:                version,
	Flags:                  append(append(commonFlags, collectFlags...), applyFlags...),
	Action: func(c *cli.Context) error {
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

func both(werr io.Writer, args []string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return fmt.Errorf("$EDITOR must be set")
	}

	f, err := os.CreateTemp("", "pounce-")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	fname := f.Name()

	defer os.Remove(fname)

	if err := collect(f, werr, args); err != nil {
		return fmt.Errorf("collect: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	cmd := exec.Command(editor, fname)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if err := err.(*exec.ExitError); err != nil {
			return fmt.Errorf("editor exited with code %d", err.ExitCode())
		}

		return fmt.Errorf("run editor %q: %w", editor, err)
	}

	if f, err = os.Open(fname); err != nil {
		return fmt.Errorf("open temp file: %w", err)
	}

	if err := apply(f); err != nil {
		return fmt.Errorf("apply: %w", err)
	}

	return nil
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
