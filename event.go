package main

import "strings"

type Event struct {
	Type    string
	Address Address
	Allocation
}

type Address uintptr

type Allocation struct {
	Size  uint64
	Stack string
}

func ParseEvent(block string) (Event, bool) {
	// ignore goroutine state changes
	if strings.HasPrefix(block, "goroutine ") {
		return Event{}, false
	}

	return Event{Allocation: Allocation{Stack: block}}, true
}
