package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

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
		if err, ok := err.(*exec.ExitError); ok && err != nil {
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
