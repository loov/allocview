package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

func main() {
	live := NewLive()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(SplitStack)

	for scanner.Scan() {
		blocktext := scanner.Text()
		event, ok := ParseEvent(blocktext)
		if !ok {
			continue
		}

		live.Include(event)
	}

	for kind, alloc := range live.TotalAllocs {
		fmt.Println(kind, alloc)
	}
}

type Live struct {
	Heap        map[Address]Allocation
	Allocated   map[string]int64
	TotalAllocs map[string]int64
}

func NewLive() *Live {
	return &Live{
		Heap:        make(map[Address]Allocation),
		Allocated:   make(map[string]int64),
		TotalAllocs: make(map[string]int64),
	}
}

func (live *Live) Include(event Event) {
	switch event.Kind {
	case Alloc:
		live.Heap[event.Address] = event.Allocation
		live.Allocated[event.Type] += event.Size
		live.TotalAllocs[event.Type] += event.Size
	case Free:
		delete(live.Heap, event.Address)
		live.Allocated[event.Type] -= event.Size
	}
}

func SplitStack(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\n', '\n'}); i >= 0 {
		return i + 2, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
