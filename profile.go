package main

import (
	"log"
	"os"
	"runtime/pprof"
)

type Profile struct {
	Cpu string
	Mem string

	cpufile *os.File
}

func (profile *Profile) Run() func() {
	profile.Start()
	return profile.Stop
}

func (profile *Profile) Start() {
	if profile.Cpu != "" {
		f, err := os.Create(profile.Cpu)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		profile.cpufile = f
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}
}

func (profile *Profile) Stop() {
	if profile.Mem != "" {
		f, err := os.Create(profile.Mem)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}

	if profile.Cpu != "" {
		pprof.StopCPUProfile()
		profile.cpufile.Close()
	}
}
