# pounce

simple find and replace in files.

# Install

```
$ go install github.com/itayd/pounce@latest
```

# Use
```
$ tree
.
├── README.md
└── some
    └── dir
        └── dogs.txt

# First collect all lines to be modified.
$ pounce -crs 'dog' > /tmp/lines
$ cat /tmp/lines
README.md:1:I like dogs
some/dir/dogs.txt:3:dogs are the best!

# The idea is that this can be done interactively using vim.
# This allows trial and error and faster feedback, but for
# the sake of demo, we'll use sed.
$ sed -i '' -e 's/dog/cat/g' /tmp/lines

# Apply modifications.
$ pounce -a < /tmp/lines
$ cat README.md
I like cats
```

Interactive one liner version (will launch $EDITOR to modify lines):
```
$ pounce -s 'dog'
```

One liner version, using sed:
```
$ pounce -crs 'dog' | sed -e 's/dog/cat/g' | pounce -a
```

Edit all lines in all files:
```
$ pounce -s ''
```

# TODO

- [ ] deal with colons in file names.
- [ ] deal with no EOL on last line of files.
- [ ] `\n` vs `\r\n`?
- [ ] apply: only generate backup files if content actually changed?
- [ ] apply: read/write piece wise.
- [ ] cli flags docs can be better phrased.

# Disclaimer

Both cats and dogs are awesome.
