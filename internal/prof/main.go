package prof

import (
	"log"
	"os"
	"runtime/pprof"
)

type Config struct {
	Cpu string
	Mem string

	cpufile *os.File
}

func (conf *Config) Run() func() {
	conf.Start()
	return conf.Stop
}

func (conf *Config) Start() {
	if conf.Cpu != "" {
		f, err := os.Create(conf.Cpu)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		conf.cpufile = f
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}
}

func (conf *Config) Stop() {
	if conf.Mem != "" {
		f, err := os.Create(conf.Mem)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}

	if conf.Cpu != "" {
		pprof.StopCPUProfile()
		conf.cpufile.Close()
	}
}
