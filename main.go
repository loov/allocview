package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
)

func main() {
	fmt.Printf("pid=%v\n", os.Getpid())

	live := NewLive()

	dump := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(dump, syscall.SIGUSR1, os.Kill)

	counter := 0
	go func() {
		defer close(done)
		for sig := range dump {
			if sig == os.Kill {
				break
			}

			w, _ := os.Create(fmt.Sprintf("snapshot-%03d.log", counter))
			buf := bufio.NewWriter(w)
			live.DeltaSnapshot(buf)
			buf.Flush()
			w.Close()

			counter++
		}
	}()

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

	dump <- os.Kill
	<-done

	for typ, alloc := range live.TotalAllocs {
		fmt.Println(live.TypeToName[typ], alloc)
	}
}

type Live struct {
	sync.Mutex

	NameToType map[string]Type
	TypeToName []string

	Heap  map[Address]Allocation
	Delta map[Address]Allocation

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

		Heap:  make(map[Address]Allocation, 1<<20),
		Delta: make(map[Address]Allocation, 1<<20),

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
	live.Lock()
	defer live.Unlock()

	typ := live.findType(event.Type)
	switch event.Kind {
	case Alloc:
		live.Heap[event.Address] = Allocation{
			Type:  typ,
			Size:  event.Size,
			Stack: event.Stack,
		}
		live.Delta[event.Address] = Allocation{
			Type:  typ,
			Size:  event.Size,
			Stack: event.Stack,
		}
		live.Allocated[typ] += event.Size
		live.TotalAllocs[typ] += event.Size

	case Free:
		delete(live.Heap, event.Address)
		delete(live.Delta, event.Address)
		live.Allocated[typ] -= event.Size
	}
}

func (live *Live) DeltaSnapshot(w io.Writer) {
	live.Lock()
	size := len(live.Delta)
	live.Unlock()

	delta := make(map[Address]Allocation, size)

	live.Lock()
	live.Delta, delta = delta, live.Delta
	typeName := live.TypeToName
	live.Unlock()

	type TypeAllocation struct {
		Type Type
		Size int64
	}

	allocationsByType := make([]TypeAllocation, len(typeName))
	for typ := range allocationsByType {
		allocationsByType[typ].Type = Type(typ)
	}
	for _, alloc := range delta {
		allocationsByType[alloc.Type].Size += alloc.Size
	}
	sort.Slice(allocationsByType, func(i, k int) bool {
		return allocationsByType[i].Size > allocationsByType[k].Size
	})

	for _, alloc := range allocationsByType {
		if alloc.Size == 0 {
			continue
		}
		fmt.Fprintf(w, "%s\t%v\n", typeName[alloc.Type], alloc.Size)
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
