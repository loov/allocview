package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

func init() { runtime.LockOSThread() }

func main() {
	flag.Parse()

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	// glfw.WindowHint(glfw.Visible, glfw.False) // do not steal focus
	glfw.WindowHint(glfw.Samples, 2)

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

	metrics := NewMetrics(time.Now(), 300*time.Millisecond, 2<<10)

	go func() {
		for {
			for i := 0; i < 10; i++ {
				traceName := fmt.Sprintf("Trace %d", rand.Intn(10))

				metrics.Update(traceName, time.Now(), Sample{
					Allocs: rand.Int63n(1000),
					Frees:  rand.Int63n(1000),
				})
			}

			millis := 100 + rand.Intn(100)
			time.Sleep(time.Duration(millis) * time.Millisecond)
		}
	}()

	view := NewMetricsView(metrics)
	app := NewApp(window, view)
	app.Run()
}
