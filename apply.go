package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

var applyFlags = struct {
	print bool
	bak   string
}{}

var applyCmd = cli.Command{
	Name:        "apply",
	UsageText:   "pounce apply [-opts]",
	Description: "apply modified modified lines",
	Aliases:     []string{"a"},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "print",
			Aliases:     []string{"p"},
			Destination: &applyFlags.print,
			Usage:       "print each incoming line",
		},
		&cli.StringFlag{
			Name:        "bak",
			Aliases:     []string{"i"},
			Destination: &applyFlags.bak,
			Usage:       "if not empty, backup originals with given suffix",
		},
	},
	Action: func(c *cli.Context) error {
		return processReplaceInput(os.Stdin)
	},
}

func processReplaceInput(r io.Reader) error {
	s := bufio.NewScanner(r)

	acc := struct {
		path string
		data map[int]string
	}{
		data: make(map[int]string),
	}

	flush := func() error {
		if len(acc.data) == 0 {
			return nil
		}

		if err := apply(acc.path, acc.data); err != nil {
			return fmt.Errorf("%s: %w", acc.path, err)
		}

		acc.path = ""
		acc.data = make(map[int]string)

		return nil
	}

	for i := 0; s.Scan(); i++ {
		text := strings.TrimLeft(s.Text(), " \t")

		if len(text) == 0 || text[0] == '#' {
			continue
		}

		if applyFlags.print {
			fmt.Fprintln(os.Stderr, text)
		}

		invErr := func() error { return fmt.Errorf("invalid input line %d", i) }

		parts := strings.SplitN(text, ":", 3)
		if len(parts) != 3 {
			return invErr()
		}

		lineNum, err := strconv.ParseInt(parts[1], 10, 32)
		if err != nil || lineNum <= 0 {
			return invErr()
		}

		if path := parts[0]; acc.path != path {
			if err := flush(); err != nil {
				return err
			}

			acc.path = path
		}

		acc.data[int(lineNum)] = parts[2]
	}

	if err := flush(); err != nil {
		return err
	}

	return nil
}

// TODO: all contents is read into memory. Need to do it piecewise.
func apply(path string, data map[int]string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	if bak := applyFlags.bak; len(bak) > 0 {
		if err := os.WriteFile(path+bak, content, 0755); err != nil {
			return fmt.Errorf("backup %s%s: %w", path, bak, err)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	defer f.Close()

	s := bufio.NewScanner(strings.NewReader(string(content)))

	for lineNum := 1; s.Scan(); lineNum++ {
		out, ok := data[lineNum]
		if !ok {
			out = s.Text()
		}

		if _, err := f.WriteString(out + "\n"); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	return nil
}