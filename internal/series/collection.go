package series

import "time"

// Collection implements sample aggregation.
type Collection struct {
	Start time.Time

	SampleDuration time.Duration
	SampleCount    int

	SampleHead int
	LastNow    time.Time

	List []*Series
}

// NewCollection returns a generic collection.
func NewCollection(start time.Time, sampleDuration time.Duration, sampleCount int) *Collection {
	return &Collection{
		Start:          start,
		SampleDuration: sampleDuration,
		SampleCount:    sampleCount,
		LastNow:        start,
	}
}

// UpdateToTime updates Collection to the specified time and
// returns the sample index corresponding to that time.
func (coll *Collection) UpdateToTime(now time.Time) SampleIndex {
	local := now.Sub(coll.Start)
	sampleTime := int(local / coll.SampleDuration)

	if now.Before(coll.LastNow) {
		panic("time travel is not possible at this moment in time")
	}
	coll.LastNow = now

	// clear any old samples that were skipped
	if coll.SampleHead != sampleTime {
		// TODO: optimize this loop
		for _, s := range coll.List {
			for t := coll.SampleHead; t < sampleTime; t++ {
				s.Samples[t%len(s.Samples)] = Sample{}
			}
		}
		coll.SampleHead = sampleTime
	}

	return SampleIndex(sampleTime % coll.SampleCount)
}
