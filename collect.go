package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/urfave/cli/v2"
)

var (
	collectConfig struct {
		exts, regexs, strs  cli.StringSlice
		bin, recursive, abs bool
	}

	collectFlags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "recursive",
			Aliases:     []string{"r"},
			Destination: &collectConfig.recursive,
			Usage:       "process subdirectories recursively",
			Category:    "collect",
		},
		&cli.StringSliceFlag{
			Name:        "ext",
			Aliases:     []string{"x"},
			Destination: &collectConfig.exts,
			Usage:       "only process files with extention",
			Category:    "collect",
		},
		&cli.StringSliceFlag{
			Name:        "regex",
			Aliases:     []string{"e"},
			Destination: &collectConfig.regexs,
			Usage:       "filter content for given regexp",
			Category:    "collect",
		},
		&cli.StringSliceFlag{
			Name:        "str",
			Aliases:     []string{"s"},
			Destination: &collectConfig.strs,
			Usage:       "filter content for given string",
			Category:    "collect",
		},
		&cli.BoolFlag{
			Name:        "bin",
			Destination: &collectConfig.bin,
			Usage:       "also search in binary files",
			Category:    "collect",
		},
		&cli.BoolFlag{
			Name:        "abs",
			Destination: &collectConfig.abs,
			Usage:       "write all file's absolute paths",
			Category:    "collect",
		},
	}
)

func collect(wout, werr io.Writer, args []string) error {
	m, err := matcher()
	if err != nil {
		return err
	}

	if !collectConfig.abs && !commonConfig.nowd {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd: %w", err)
		}

		fmt.Fprintf(wout, "%s%s\n", wdprefix, trimPath(cwd))
	}

	if len(args) == 0 {
		return gather(wout, werr, "", ".", m)
	}

	for _, arg := range args {
		if err := gather(wout, werr, "", arg, m); err != nil {
			return err
		}
	}

	return nil
}

func matcher() (func(string) bool, error) {
	var ms []func(string) bool

	for _, s := range collectConfig.strs.Value() {
		s1 := s
		ms = append(ms, func(x string) bool { return strings.Contains(x, s1) })
	}

	for _, s := range collectConfig.regexs.Value() {
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

func gather(wout, werr io.Writer, prev, path string, matcher func(string) bool) error {
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
		if prev != "" && !collectConfig.recursive {
			return nil
		}

		names, err := f.Readdirnames(0)
		if err != nil {
			return fmt.Errorf("readdir %s: %w", path, err)
		}

		// Make output predictable.
		sort.Strings(names)

		for _, name := range names {
			if err := gather(wout, werr, path, filepath.Join(path, name), matcher); err != nil {
				return err
			}
		}

		return nil
	}

	if !fi.Mode().IsRegular() {
		return nil
	}

	if exts := collectConfig.exts.Value(); len(exts) > 0 {
		ext := filepath.Ext(path)
		if len(ext) > 0 && ext[0] == '.' {
			ext = ext[1:]
		}

		if !slices.Contains(exts, ext) {
			return nil
		}
	}

	var r io.Reader = f

	if !collectConfig.bin {
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

	if collectConfig.abs {
		apath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("abs %s: %w", path, err)
		}

		if trimPathPrefix != "" {
			apath = strings.TrimPrefix(apath, trimPathPrefix)
		}

		path = apath
	}

	if commonConfig.print {
		fmt.Fprintln(werr, trimPath(path))
	}

	s := bufio.NewScanner(r)
	for lineNum := 1; s.Scan(); lineNum++ {

		if text := s.Text(); matcher(text) {
			fmt.Fprintf(wout, "%s:%d:%s\n", trimPath(path), lineNum, text)
		}
	}

	return nil
}
