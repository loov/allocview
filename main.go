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

	Parse(live, os.Stdin)

	dump <- os.Kill
	<-done

	for typ, alloc := range live.TotalAllocs {
		fmt.Println(live.TypeToName[typ], alloc)
	}
}

func Parse(live *Live, in io.Reader) {
	scanner := bufio.NewScanner(in)
	scanner.Split(SplitStack)

	for scanner.Scan() {
		blocktext := scanner.Text()
		event, ok := ParseEvent(blocktext)
		if !ok {
			continue
		}

		live.Include(event)
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
