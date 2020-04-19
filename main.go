package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/unit"
	"golang.org/x/sync/errgroup"

	"loov.dev/allocview/internal/allocfreetrace"
)

func init() {
	runtime.LockOSThread()
	gofont.Register()
}

func main() {
	var profile Profile
	flag.StringVar(&profile.Cpu, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&profile.Mem, "memprofile", "", "write memory profile to `file`")

	var interval time.Duration
	flag.DurationVar(&interval, "interval", time.Second, "sampling interval")

	var simulate bool
	flag.BoolVar(&simulate, "simulate", false, "simulate memory usage")

	flag.Parse()

	defer profile.Run()()

	metrics := NewMetrics(time.Now(), interval, 2<<10)

	var group errgroup.Group
	if simulate {
		group.Go(func() error { return ReadSim(metrics) })
	} else {
		group.Go(func() error { return ReadInput(metrics, os.Stdin) })
		group.Go(func() error { return RegularUpdates(metrics, interval) })
	}

	group.Go(func() error {
		window := app.NewWindow(app.Size(unit.Dp(800), unit.Dp(650)))
		view := NewMetricsView(metrics)
		return view.Run(window)
	})

	app.Main()

	err := group.Wait()
	if err != nil {
		log.Println(err)
	}
}

func ReadSim(metrics *Metrics) error {
	for {
		for i := 0; i < 10; i++ {
			span := fmt.Sprintf("allocfreetrace %d", i)
			metrics.Update(span, time.Now(), Sample{
				Allocs: 100 + rand.Int63n(10000),
				Frees:  100 + rand.Int63n(10000),
			})
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func ReadInput(metrics *Metrics, r io.Reader) error {
	reader := allocfreetrace.NewReader(os.Stdin)
	for {
		event, err := reader.Read()
		if err != nil {
			return err
		}

		now := time.Now()

		switch event.Kind {
		case allocfreetrace.Alloc:
			metrics.Update(event.Type, now, Sample{
				Allocs: event.Size,
			})
		case allocfreetrace.Free:
			metrics.Update(event.Type, now, Sample{
				Frees: event.Size,
			})
		}
	}
}

func RegularUpdates(metrics *Metrics, interval time.Duration) error {
	tick := time.NewTicker(interval)
	for range tick.C {
		metrics.Update("time", time.Now(), Sample{})
	}
	return nil
}
