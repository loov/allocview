package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Kind byte

const (
	Invalid Kind = iota
	Alloc
	Free
	GC
)

func (kind Kind) String() string {
	switch kind {
	case Alloc:
		return "alloc"
	case Free:
		return "free"
	case GC:
		return "gc"
	default:
		return "invalid"
	}
}

type Event struct {
	Kind    Kind
	Address Address
	Type    string
	Size    int64
	Stack   string
}

type Address uintptr

func ParseEvent(block string) (Event, bool) {
	header, stack := splitBlock(block)
	kind, address, typ, size := parseHeader(header)
	if kind == Invalid || kind == GC {
		return Event{}, false
	}

	return Event{
		Kind:    kind,
		Address: address,
		Type:    typ,
		Size:    size,
		Stack:   stack,
	}, true
}

var (
	rxAlloc = regexp.MustCompile(`^\(0x([0-9a-f]+), 0x([0-9a-f]+)(?:, (.*))?\)$`)
	rxFree  = regexp.MustCompile(`^\(0x([0-9a-f]+), 0x([0-9a-f]+)\)$`)
)

func parseHeader(header string) (Kind, Address, string, int64) {
	p := strings.IndexAny(header, "( ")
	if p < 0 {
		return Invalid, 0, "", 0
	}

	switch header[:p] {
	case "tracealloc":
		// tracealloc(0xc00005ea80, 0x180, runtime.g)
		// tracealloc(0xc00005ea80, 0x180)
		tokens := rxAlloc.FindStringSubmatch(header[p:])
		if len(tokens) != 4 {
			fmt.Printf("%q %v\n", header[p:], tokens)
			panic(header)
		}
		address, err := strconv.ParseUint(tokens[1], 16, 64)
		if err != nil {
			panic(err)
		}
		size, err := strconv.ParseInt(tokens[2], 16, 64)
		if err != nil {
			panic(err)
		}
		return Alloc, Address(address), tokens[3], size
	case "tracefree":
		// tracefree(0xc0006a2090, 0x30)
		tokens := rxFree.FindStringSubmatch(header[p:])
		if len(tokens) != 3 {
			fmt.Printf("%q\n", header[p:])
			panic(header)
		}
		address, err := strconv.ParseUint(tokens[1], 16, 64)
		if err != nil {
			panic(err)
		}
		size, err := strconv.ParseInt(tokens[2], 16, 64)
		if err != nil {
			panic(err)
		}
		return Free, Address(address), "", size
	case "tracegc":
		return GC, 0, "", 0
	case "goroutine":
		return Invalid, 0, "", 0
	default:
		return Invalid, 0, "", 0
	}
}

func splitBlock(block string) (header, stack string) {
	tokens := strings.SplitN(block, "\n", 2)
	if len(tokens) != 2 {
		return block, ""
	}
	return tokens[0], tokens[1]
}
