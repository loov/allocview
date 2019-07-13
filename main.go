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

	for typ, alloc := range live.TotalAllocs {
		fmt.Println(live.TypeToName[typ], alloc)
	}
}

type Live struct {
	NameToType map[string]Type
	TypeToName []string

	Heap map[Address]Allocation

	// indexed by type
	Allocated   []int64
	TotalAllocs []int64
}

type Type int

type Allocation struct {
	Type  Type
	Size  int64
	Stack string
}

func NewLive() *Live {
	return &Live{
		NameToType: make(map[string]Type, 1<<20),
		TypeToName: make([]string, 0, 1<<20),

		Heap: make(map[Address]Allocation, 1<<20),

		Allocated:   make([]int64, 0, 1<<20),
		TotalAllocs: make([]int64, 0, 1<<20),
	}
}

func (live *Live) findType(name string) Type {
	typ, ok := live.NameToType[name]
	if ok {
		return typ
	}

	typ = Type(len(live.TypeToName))
	live.TypeToName = append(live.TypeToName, name)
	live.NameToType[name] = typ

	live.Allocated = append(live.Allocated, 0)
	live.TotalAllocs = append(live.TotalAllocs, 0)

	return typ
}

func (live *Live) Include(event Event) {
	typ := live.findType(event.Type)
	switch event.Kind {
	case Alloc:
		live.Heap[event.Address] = Allocation{
			Type:  typ,
			Size:  event.Size,
			Stack: event.Stack,
		}
		live.Allocated[typ] += event.Size
		live.TotalAllocs[typ] += event.Size
	case Free:
		delete(live.Heap, event.Address)
		live.Allocated[typ] -= event.Size
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
