package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"

	"github.com/loov/allocview/internal/ui/draw"
	"github.com/loov/allocview/trace"
)

var DefaultFont *draw.Font

func init() { runtime.LockOSThread() }

func main() {
	var interval time.Duration
	flag.DurationVar(&interval, "interval", time.Second, "sampling interval")

	var simulate bool
	flag.BoolVar(&simulate, "simulate", false, "simulate memory usage")

	flag.Parse()

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
		go Parse(metrics, os.Stdin)
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
}

func Parse(metrics *Metrics, in io.Reader) {
	scanner := bufio.NewScanner(in)
	scanner.Split(SplitStack)

	for scanner.Scan() {
		blocktext := scanner.Text()
		event, ok := trace.ParseEvent(blocktext)
		if !ok {
			continue
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
