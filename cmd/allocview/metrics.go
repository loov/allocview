package main

import (
	"sync"
	"time"
)

type Metrics struct {
	sync.Mutex

	Since time.Time

	SampleDuration time.Duration
	SampleCount    int
	SampleTime     int

	ByName map[string]*Metric
	List   []*Metric
}

func NewMetrics(since time.Time, sampleDuration time.Duration, sampleCount int) *Metrics {
	return &Metrics{
		Since: since,

		SampleDuration: sampleDuration,
		SampleCount:    sampleCount,
		SampleTime:     0,

		ByName: make(map[string]*Metric),
	}
}

func (metrics *Metrics) Reset() {
	metrics.Lock()
	defer metrics.Unlock()

	for _, m := range metrics.List {
		for i := range m.Samples {
			m.Samples[i].Reset()
		}
	}
}

func (metrics *Metrics) Update(name string, now time.Time, sample Sample) {
	metrics.Lock()
	defer metrics.Unlock()

	metric, ok := metrics.ByName[name]
	if !ok {
		metric = &Metric{
			Name:    name,
			Samples: make([]Sample, metrics.SampleCount),
		}
		metrics.ByName[name] = metric
		metrics.List = append(metrics.List, metric)
	}

	local := now.Sub(metrics.Since)
	sampleTime := int(local / metrics.SampleDuration)
	if metrics.SampleTime != sampleTime {
		metrics.SampleTime = sampleTime
		for _, m := range metrics.List {
			m.Samples[sampleTime].Reset()
		}
	}

	metric.Live += sample.Allocs - sample.Frees
	metric.Samples[sampleTime].Add(sample)
}

type Metric struct {
	Name    string
	Live    int64
	Samples []Sample
}

type Sample struct {
	Allocs int64
	Frees  int64
}

func (metric *Metric) Max() (max Sample) {
	for _, sample := range metric.Samples {
		max.Allocs = Max(max.Allocs, sample.Allocs)
		max.Frees = Max(max.Frees, sample.Frees)
	}
	return max
}

func (sample *Sample) Reset() {
	*sample = Sample{}
}

func (sample *Sample) Add(b Sample) {
	sample.Allocs += b.Allocs
	sample.Frees += b.Frees
}

func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
