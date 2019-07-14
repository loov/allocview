package main

import (
	"github.com/loov/allocview/internal/ui"
	"github.com/loov/allocview/internal/ui/g"
)

var (
	BackgroundColor = g.Color{0, 0, 0, 0xFF}
	AllocsColor     = g.HSL(0, 0.6, 0.6)
	FreesColor      = g.HSL(0.3, 0.6, 0.6)
)

type MetricsView struct {
	Metrics *Metrics
}

func NewMetricsView(metrics *Metrics) *MetricsView {
	return &MetricsView{
		Metrics: metrics,
	}
}

func (view *MetricsView) Reset() {}

func (view *MetricsView) Update(ctx *ui.Context) {
	ctx.Draw.FillRect(&ctx.Area, BackgroundColor)

	metrics := view.Metrics

	metrics.Lock()
	defer metrics.Unlock()

	const MetricHeight = 50
	const SampleWidth = 5

	samples := ctx.Area.Size().X / SampleWidth
	// TODO: clamp to max size

	low := int(float32(metrics.SampleTime) - samples)
	if low < 0 {
		low = 0
	}
	high := low + int(g.Ceil(samples))

	top := float32(0.0)
	for i, metric := range metrics.List {
		ctx := ctx.Row(top, top+MetricHeight)
		top += MetricHeight

		color := g.RGB(0.1, 0.1, 0.1)
		if i%2 == 1 {
			color = g.RGB(0.2, 0.2, 0.2)
		}
		ctx.Draw.FillRect(&ctx.Area, color)

		max := metric.Max()
		maxValue := Max(max.Allocs, max.Frees)
		scale := (ctx.Area.Size().Y / 2) / float32(maxValue)

		corner := ctx.Area.LeftCenter()
		for p := low; p < high; p++ {
			sample := metric.Samples[p%metrics.SampleCount]

			ctx.Draw.FillRect(&g.Rect{
				Min: corner,
				Max: corner.Add(g.V(SampleWidth, float32(sample.Allocs)*scale)),
			}, AllocsColor)

			ctx.Draw.FillRect(&g.Rect{
				Min: corner,
				Max: corner.Add(g.V(SampleWidth, float32(-sample.Frees)*scale)),
			}, FreesColor)

			corner.X += SampleWidth
		}
	}
}
