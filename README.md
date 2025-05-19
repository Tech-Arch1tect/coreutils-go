This is an attempt at a basic re-implementation of GNU coreutils in Go.

It is just a project used to learn more Go, learn more about coreutils, etc. These tools are not designed to be exactly 1:1 with GNU coreutils. Many things like error messages will probably be different. But the functionality (e.g. usage, flags etc) _should_ be the same.

## Status

7 passing test(s)

```
Total tests:    658
Passed:         7
Skipped:        596
Expected fail:  0
Unexpected pass:0
Failures:       55
Errors:         0
```

## Goals

- ~~Create initial (empty) `cp` program~~
- ~~Create a way to re-use GNU coreutils tests with our example binary~~
- Start actualy implementing coreutils
- Create CI for automated testing / tracking # of tests

## Dev Summary

- Build all binaries: `just build`
- Run tests `just test`
- Build and run all tests: `just bt`
- Build and run only `tests/cksum/md5sum.pl` test: `TESTS="tests/cksum/md5sum.pl" just bt`
