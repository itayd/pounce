# pounce

simple find and replace in files.

```
$ pounce collect -r -s 'dog' > /tmp/dogs.txt
$ cat /tmp/dogs.txt
README.md:1:I like dogs
some/dir/dogs.txt:3:dogs are the best!
$ # The idea is that this can be done interactively using vim.
$ # This allows trial and error and faster feedback, but for
$ # the sake of demo, we'll use sed.
$ sed -i '' -e 's/dog/cat/g' /tmp/dogs.txt
$ pounce apply < /tmp/dogs.txt
$ cat README.md
I like cats
```

One liner version:
```
$ pounce collect -r -s 'dog' | sed -e 's/dog/cat/g' | pounce apply
```

# TODO

- [ ] tests
- [ ] deal with colons in file names.
- [ ] deal with no EOL last line.
- [ ] apply: only generate backup files if content actually changed.
- [ ] apply: read/write piece wise.
- [ ] \n vs \r\n?

# Disclaimer

Both cats and dogs are awesome.
