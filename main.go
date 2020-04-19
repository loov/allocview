package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/unit"
	"golang.org/x/sync/errgroup"

	"loov.dev/allocview/internal/prof"
)

func init() {
	runtime.LockOSThread()
	gofont.Register()
}

func main() {
	ctx := context.Background()

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, `Usage: %s [flags] subcommand...

This tool visualizes allocations of a Go program.

Only programs that have imported "loov.dev/allocview/attach" are supported at the moment.

When given a subcommand, it executes that subcommand and starts live-visualization
of the program. As an example:

    allocview go run ./testdata

Flags:
`, os.Args[0])
		flag.PrintDefaults()
	}

	var profcfg prof.Config
	flag.StringVar(&profcfg.Cpu, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&profcfg.Mem, "memprofile", "", "write memory profile to `file`")

	var config Config

	flag.DurationVar(&config.SampleDuration, "sample-duration", time.Second, "sample duration")
	flag.IntVar(&config.SampleCount, "sample-count", 1024, "sample count")

	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	defer profcfg.Run()()

	// Setup command that we want to monitor.
	args := flag.Args()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	var group errgroup.Group

	server := NewServer()
	err := server.Exec(ctx, &group, cmd)
	if err != nil {
		log.Fatal(err)
	}

	group.Go(func() error {
		window := app.NewWindow(app.Size(unit.Dp(800), unit.Dp(650)))
		view := NewView(config, server)
		return view.Run(window)
	})

	app.Main()

	err = group.Wait()
	if err != nil {
		log.Println(err)
	}
}
