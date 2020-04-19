package main

import (
	"time"

	"loov.dev/allocview/internal/series"
)

type Summary struct {
	Config Config

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
	// TODO: check whether we have loaded exe symbols

	collection := summary.Collection
	index := collection.UpdateToTime(profile.Time)
	for i := range profile.Records {
		rec := &profile.Records[i]

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
