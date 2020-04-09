package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/unit"
	"golang.org/x/sync/errgroup"

	"github.com/loov/allocview/trace"
)

func init() { runtime.LockOSThread() }

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile := flag.String("memprofile", "", "write memory profile to `file`")

	var interval time.Duration
	flag.DurationVar(&interval, "interval", time.Second, "sampling interval")

	var simulate bool
	flag.BoolVar(&simulate, "simulate", false, "simulate memory usage")

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	gofont.Register()

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

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func ReadSim(metrics *Metrics) error {
	for {
		for i := 0; i < 10; i++ {
			span := fmt.Sprintf("trace %d", i)
			metrics.Update(span, time.Now(), Sample{
				Allocs: 100 + rand.Int63n(10000),
				Frees:  100 + rand.Int63n(10000),
			})
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func ReadInput(metrics *Metrics, r io.Reader) error {
	reader := trace.NewReader(os.Stdin)
	for {
		event, err := reader.Read()
		if err != nil {
			return err
		}

		now := time.Now()

		switch event.Kind {
		case trace.Alloc:
			metrics.Update(event.Type, now, Sample{
				Allocs: event.Size,
			})
		case trace.Free:
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
