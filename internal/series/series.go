package series

// Series is a ring-buffer indexed by Ring.
type Series struct {
	TotalAllocBytes   int64
	TotalAllocObjects int64
	Samples           []Sample
}

// SampleIndex indexes Samples slice in Series.
type SampleIndex int

func (series *Series) UpdateSample(index SampleIndex, sample Sample) {
	series.TotalAllocBytes += sample.AllocBytes - sample.FreeBytes
	series.TotalAllocObjects += sample.AllocObjects - sample.FreeObjects

	series.Samples[index].Add(sample)
}

// Sample is total allocated or freed in SampleDuration.
type Sample struct {
	AllocBytes   int64
	FreeBytes    int64
	AllocObjects int64
	FreeObjects  int64
}

// Add calculates the total.
func (sample *Sample) Add(b Sample) {
	sample.AllocBytes += b.AllocBytes
	sample.FreeBytes += b.FreeBytes
	sample.AllocObjects += b.AllocObjects
	sample.FreeObjects += b.FreeObjects
}

// Max returns the largest sample values.
func (series *Series) Max() (r Sample) {
	for _, sample := range series.Samples {
		r.AllocBytes = max(r.AllocBytes, sample.AllocBytes)
		r.FreeBytes = max(r.FreeBytes, sample.FreeBytes)
		r.AllocObjects = max(r.AllocObjects, sample.AllocObjects)
		r.FreeObjects = max(r.FreeObjects, sample.FreeObjects)
	}
	return r
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
