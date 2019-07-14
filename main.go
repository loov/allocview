package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/loov/allocview/trace"
)

func init() { runtime.LockOSThread() }

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	var group Group

	live := NewLive()

	group.Go(func() {
		MonitorSignals(ctx, live)
	})

	Parse(live, os.Stdin)

	cancel()
	group.Wait()

	live.WriteSummary(os.Stdout, live.TypeToName, live.Heap)
}

func MonitorSignals(ctx context.Context, live *Live) {
	dump := make(chan os.Signal, 1)

	go func() {
		select {
		case <-ctx.Done():
			dump <- os.Kill
		}
	}()

	signal.Notify(dump,
		syscall.SIGUSR1, syscall.SIGUSR2,
		os.Kill,
	)

	snapshot := 0
	delta := 0

	for sig := range dump {
		if sig == os.Kill {
			break
		}

		switch sig {
		case os.Kill:
			return
		case syscall.SIGUSR1:
			w, _ := os.Create(fmt.Sprintf("snapshot-%03d.log", snapshot))
			buf := bufio.NewWriter(w)
			live.DeltaSnapshot(buf)
			buf.Flush()
			w.Close()
			snapshot++

		case syscall.SIGUSR2:
			w, _ := os.Create(fmt.Sprintf("delta-%03d.log", delta))
			buf := bufio.NewWriter(w)
			live.Snapshot(buf)
			buf.Flush()
			w.Close()
			delta++
		}
	}
}

func Parse(live *Live, in io.Reader) {
	scanner := bufio.NewScanner(in)
	scanner.Split(SplitStack)

	for scanner.Scan() {
		blocktext := scanner.Text()
		event, ok := trace.ParseEvent(blocktext)
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
