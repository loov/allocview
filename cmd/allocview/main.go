package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"

	"github.com/loov/allocview/internal/ui/draw"
	"github.com/loov/allocview/trace"
)

var DefaultFont *draw.Font

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

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	// glfw.WindowHint(glfw.Visible, glfw.False) // do not steal focus
	// glfw.WindowHint(glfw.Samples, 2)

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	window, err := glfw.CreateWindow(800, 600, "AllocView", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	window.Restore()
	window.SetPos(32, 64)

	if err := gl.Init(); err != nil {
		panic(err)
	}
	if err := gl.GetError(); err != 0 {
		panic(err)
	}

	DefaultFont, err = draw.LoadTTF("DefaultFont.ttf", 72, 32)
	if err != nil {
		panic(err)
	}
	DefaultFont.LoadExtendedAscii()

	metrics := NewMetrics(time.Now(), interval, 2<<10)

	if simulate {
		go func() {
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
		}()
	} else {
		go func() {
			reader := trace.NewReader(os.Stdin)
			for {
				event, err := reader.Read()
				if err != nil {
					log.Println(err)
					return
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
		}()

		go func() {
			tick := time.NewTicker(interval)
			for range tick.C {
				metrics.Update("time", time.Now(), Sample{})
			}
		}()
	}

	view := NewMetricsView(metrics)
	app := NewApp(window, view)
	app.Run()
	defer runtime.KeepAlive(app)

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
