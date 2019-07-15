# allocview

allocview analyses `GODEBUG=allocfreetrace=1 ./program` output and prints summary.

TODO:

* [ ] rewrite history to remove accidetnally comitted binaries
* [ ] summary of stack traces
* [ ] type, summary of stack traces
* [ ] graph output
* [ ] multiple draw buffers
* [ ] text output

# How to use

Run the program with:

```
GODEBUG=allocfreetrace=1 ./program  3>&1 1>&2 2>&3 3>&- | allocview

GODEBUG=allocfreetrace=1 ./program  3>&1 1>&2 2>&3 3>&- | allocmonitor
```

The magic incantation `3>&1 1>&2 2>&3 3>&-` swaps stdout and stderr such that Go trace output can be parsed by allocview tools.