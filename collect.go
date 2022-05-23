package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"
	"golang.org/x/tools/godoc/util"

	"github.com/urfave/cli/v2"
)

var collectFlags = struct {
	exts      cli.StringSlice
	regexs    cli.StringSlice
	strs      cli.StringSlice
	print     bool
	bin       bool
	recursive bool
}{}

var collectCmd = cli.Command{
	Name:        "collect",
	UsageText:   "pounce collect [-opts] files",
	Description: "print lines that match given matchers in files",
	Aliases:     []string{"c"},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "print",
			Aliases:     []string{"p"},
			Destination: &collectFlags.print,
			Usage:       "print matched files to stderr",
		},
		&cli.BoolFlag{
			Name:        "recursive",
			Aliases:     []string{"r"},
			Destination: &collectFlags.recursive,
			Usage:       "process subdirectories recursively",
		},
		&cli.StringSliceFlag{
			Name:        "ext",
			Aliases:     []string{"x"},
			Destination: &collectFlags.exts,
			Usage:       "only process files with extention",
		},
		&cli.StringSliceFlag{
			Name:        "regex",
			Aliases:     []string{"e"},
			Destination: &collectFlags.regexs,
			Usage:       "filter content for regexp",
		},
		&cli.StringSliceFlag{
			Name:        "str",
			Aliases:     []string{"s"},
			Destination: &collectFlags.strs,
			Usage:       "filter content for strings",
		},
		&cli.BoolFlag{
			Name:        "bin",
			Destination: &collectFlags.bin,
			Usage:       "also search in binary files",
		},
	},
	Action: func(c *cli.Context) error {
		m, err := matcher()
		if err != nil {
			return err
		}

		args := c.Args().Slice()
		if len(args) == 0 {
			return gather("", ".", m)
		}

		for _, arg := range args {
			if err := gather("", arg, m); err != nil {
				return err
			}
		}

		return nil
	},
}

func matcher() (func(string) bool, error) {
	var ms []func(string) bool

	for _, s := range collectFlags.strs.Value() {
		ms = append(ms, func(x string) bool { return strings.Contains(x, s) })
	}

	for _, s := range collectFlags.regexs.Value() {
		re, err := regexp.Compile(s)
		if err != nil {
			return nil, fmt.Errorf("matcher %q: %w", s, err)
		}

		ms = append(ms, func(x string) bool { return re.MatchString(x) })
	}

	return func(s string) bool {
		for _, m := range ms {
			if m(s) {
				return true
			}
		}

		return false
	}, nil
}

func gather(prev, path string, matcher func(string) bool) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}

	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	if fi.IsDir() {
		if prev != "" && !collectFlags.recursive {
			return nil
		}

		names, err := f.Readdirnames(0)
		if err != nil {
			return fmt.Errorf("readdir %s: %w", path, err)
		}

		for _, name := range names {
			if err := gather(path, filepath.Join(path, name), matcher); err != nil {
				return err
			}
		}

		return nil
	}

	if exts := collectFlags.exts.Value(); len(exts) > 0 {
		ext := filepath.Ext(path)
		if len(ext) > 0 && ext[0] == '.' {
			ext = ext[1:]
		}

		if !slices.Contains(exts, ext) {
			return nil
		}
	}

	var r io.Reader = f

	if !collectFlags.bin {
		pre := make([]byte, 32)
		n, err := f.Read(pre)
		if err != nil && err != io.EOF {
			return fmt.Errorf("read %s: %w", path, err)
		}

		pre = pre[:n]

		if !util.IsText(pre) {
			return nil
		}

		r = io.MultiReader(bytes.NewReader(pre), f)
	}

	if collectFlags.print {
		fmt.Fprintln(os.Stderr, path)
	}

	s := bufio.NewScanner(r)
	for lineNum := 1; s.Scan(); lineNum++ {
		text := s.Text()

		if matcher(text) {
			fmt.Printf("%s:%d:%s\n", path, lineNum, text)
		}
	}

	return nil
}
