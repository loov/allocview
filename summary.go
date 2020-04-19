package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"loov.dev/allocview/internal/series"
	"loov.dev/allocview/internal/symbols"
)

type Summary struct {
	Config Config

	Symbols    *symbols.Binary
	Collection *series.Collection3
}

func NewSummary(config Config) *Summary {
	return &Summary{
		Config:     config,
		Collection: series.NewCollection3(time.Now(), config.SampleDuration, config.SampleCount),
	}
}

// Add adds profile to the collections.
func (summary *Summary) Add(profile *Profile) {
	// TODO: per binary symbols
	if summary.Symbols == nil {
		// TODO: is there a better location to do this?
		var err error
		summary.Symbols, err = symbols.Load(profile.ExeName)
		if err != nil {
			log.Fatal(err)
		}

		summary.Symbols.UpdateOffset(profile.FuncName, profile.FuncAddr)
	}

	collection := summary.Collection
	index := collection.UpdateToTime(profile.Time)
	for i := range profile.Records {
		rec := &profile.Records[i]
		for i, frame := range rec.Stack0 {
			if frame == 0 {
				break
			}
			rec.Stack0[i] = uintptr(int64(frame) + summary.Symbols.Offset)
		}

		// TODO: implement skip runtime
		collection.UpdateSample(index, rec.Stack0[:], series.Sample{
			AllocBytes:   rec.AllocBytes,
			FreeBytes:    rec.FreeBytes,
			AllocObjects: rec.AllocObjects,
			FreeObjects:  rec.FreeObjects,
		})
	}

	// TODO: reuse profile allocation
}

func (summary *Summary) StackAsString(stack []uintptr) string {
	var s bytes.Buffer
	for _, frame := range stack {
		if frame == 0 {
			break
		}

		file, line, _ := summary.Symbols.SymTable.PCToLine(uint64(frame))
		if file == "" {
			fmt.Fprintf(&s, "0x%x\n", frame)
			continue
		}
		fmt.Fprintf(&s, "%s:%v\n", file, line)
	}
	return s.String()
}
