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
	"unicode/utf8"

	"golang.org/x/exp/slices"

	"github.com/urfave/cli/v2"
)

const wdprefix = "# wd="

type collectFlags struct {
	exts, regexs, strs               cli.StringSlice
	print, bin, recursive, abs, nowd bool
}

var cliCollectFlags collectFlags

var collectCmd = cli.Command{
	Name:        "collect",
	UsageText:   "pounce collect [-opts] files",
	Description: "print lines that match given matchers in files",
	Aliases:     []string{"c"},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "print",
			Aliases:     []string{"p"},
			Destination: &cliCollectFlags.print,
			Usage:       "print matched files to stderr",
		},
		&cli.BoolFlag{
			Name:        "recursive",
			Aliases:     []string{"r"},
			Destination: &cliCollectFlags.recursive,
			Usage:       "process subdirectories recursively",
		},
		&cli.StringSliceFlag{
			Name:        "ext",
			Aliases:     []string{"x"},
			Destination: &cliCollectFlags.exts,
			Usage:       "only process files with extention",
		},
		&cli.StringSliceFlag{
			Name:        "regex",
			Aliases:     []string{"e"},
			Destination: &cliCollectFlags.regexs,
			Usage:       "filter content for regexp",
		},
		&cli.StringSliceFlag{
			Name:        "str",
			Aliases:     []string{"s"},
			Destination: &cliCollectFlags.strs,
			Usage:       "filter content for strings",
		},
		&cli.BoolFlag{
			Name:        "bin",
			Destination: &cliCollectFlags.bin,
			Usage:       "also search in binary files",
		},
		&cli.BoolFlag{
			Name:        "abs",
			Aliases:     []string{"a"},
			Destination: &cliCollectFlags.abs,
			Usage:       "write all file's absolute paths",
		},
		&cli.BoolFlag{
			Name:        "nowd",
			Destination: &cliCollectFlags.nowd,
			Usage:       "supress working directory comment",
		},
	},
	Action: func(c *cli.Context) error {
		return collect(os.Stdout, os.Stderr, cliCollectFlags, c.Args().Slice())
	},
}

func collect(wout, werr io.Writer, flags collectFlags, args []string) error {
	m, err := matcher(flags)
	if err != nil {
		return err
	}

	if !flags.abs && !flags.nowd {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd: %w", err)
		}

		fmt.Fprintf(wout, "%s%s\n", wdprefix, cwd)
	}

	if len(args) == 0 {
		return gather(wout, werr, flags, "", ".", m)
	}

	for _, arg := range args {
		if err := gather(wout, werr, flags, "", arg, m); err != nil {
			return err
		}
	}

	return nil
}

func matcher(flags collectFlags) (func(string) bool, error) {
	var ms []func(string) bool

	for _, s := range flags.strs.Value() {
		ms = append(ms, func(x string) bool { return strings.Contains(x, s) })
	}

	for _, s := range flags.regexs.Value() {
		re, err := regexp.Compile(s)
		if err != nil {
			return nil, fmt.Errorf("matcher %q: %w", s, err)
		}

		ms = append(ms, func(x string) bool { return re.MatchString(x) })
	}

	if len(ms) == 0 {
		return nil, fmt.Errorf("no content filters set")
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

func gather(wout, werr io.Writer, flags collectFlags, prev, path string, matcher func(string) bool) error {
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
		if prev != "" && !flags.recursive {
			return nil
		}

		names, err := f.Readdirnames(0)
		if err != nil {
			return fmt.Errorf("readdir %s: %w", path, err)
		}

		for _, name := range names {
			if err := gather(wout, werr, flags, path, filepath.Join(path, name), matcher); err != nil {
				return err
			}
		}

		return nil
	}

	if !fi.Mode().IsRegular() {
		return nil
	}

	if exts := flags.exts.Value(); len(exts) > 0 {
		ext := filepath.Ext(path)
		if len(ext) > 0 && ext[0] == '.' {
			ext = ext[1:]
		}

		if !slices.Contains(exts, ext) {
			return nil
		}
	}

	var r io.Reader = f

	if !flags.bin {
		pre := make([]byte, 32)
		n, err := f.Read(pre)
		if err != nil && err != io.EOF {
			return fmt.Errorf("read %s: %w", path, err)
		}

		pre = pre[:n]

		if !isText(pre) {
			return nil
		}

		r = io.MultiReader(bytes.NewReader(pre), f)
	}

	if flags.abs {
		apath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("abs %s: %w", path, err)
		}

		path = apath
	}

	if flags.print {
		fmt.Fprintln(werr, path)
	}

	s := bufio.NewScanner(r)
	for lineNum := 1; s.Scan(); lineNum++ {

		if text := s.Text(); matcher(text) {
			fmt.Fprintf(wout, "%s:%d:%s\n", path, lineNum, text)
		}
	}

	return nil
}

// from golang.org/x/tools/godoc/util.
func isText(s []byte) bool {
	const max = 1024 // at least utf8.UTFMax
	if len(s) > max {
		s = s[0:max]
	}
	for i, c := range string(s) {
		if i+utf8.UTFMax > len(s) {
			// last char may be incomplete - ignore
			break
		}
		if c == 0xFFFD || c < ' ' && c != '\n' && c != '\t' && c != '\f' {
			// decoding error or control character - not a text file
			return false
		}
	}
	return true
}
