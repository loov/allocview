package series

import "time"

// Collection3 implements sample aggregation based on 3 stack frames.
type Collection3 struct {
	Collection
	ByStack map[[3]uintptr]*Series
}

// NewCollection3 returns a new Collection3.
func NewCollection3(start time.Time, sampleDuration time.Duration, sampleCount int) *Collection3 {
	return &Collection3{
		Collection: *NewCollection(start, sampleDuration, sampleCount),
		ByStack:    make(map[[3]uintptr]*Series),
	}
}

// UpdateSample updates the sample at specified index for the specific stack.
func (coll *Collection3) UpdateSample(index SampleIndex, stack []uintptr, sample Sample) {
	var h [3]uintptr
	copy(h[:], stack)

	series, ok := coll.ByStack[h]
	if !ok {
		series = &Series{
			Samples: make([]Sample, coll.SampleCount),
		}
		coll.ByStack[h] = series
		coll.List = append(coll.List, series)
	}

	series.UpdateSample(index, sample)
}
