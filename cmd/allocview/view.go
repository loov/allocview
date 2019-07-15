package main

import (
	"github.com/loov/allocview/internal/ui"
	"github.com/loov/allocview/internal/ui/g"
)

var (
	BackgroundColor = g.Color{0, 0, 0, 0xFF}
)

type MetricsView struct {
	Metrics *Metrics

	Scroll float32
}

func NewMetricsView(metrics *Metrics) *MetricsView {
	return &MetricsView{
		Metrics: metrics,
	}
}

func (view *MetricsView) Reset() {
	view.Metrics.Reset()
}

func (view *MetricsView) Update(ctx *ui.Context) {
	ctx.Draw.FillRect(&ctx.Area, BackgroundColor)

	metrics := view.Metrics

	metrics.Lock()
	defer metrics.Unlock()

	const MetricHeight = 50
	const HeaderHeight = 12
	const SampleWidth = 3

	samples := ctx.Area.Size().X / SampleWidth
	// TODO: clamp to max size

	low := int(float32(metrics.SampleTime) - samples)
	if low < 0 {
		low = 0
	}
	high := low + int(g.Ceil(samples))

	view.Scroll += ctx.Input.Mouse.Scroll.Y
	top := view.Scroll
	for i, metric := range metrics.List {
		header := ctx.Row(top, top+MetricHeight)
		_ = header
		// TODO: add font rendering
		top += HeaderHeight

		ctx := ctx.Row(top, top+MetricHeight)
		top += MetricHeight

		color := g.RGB(0.1, 0.1, 0.1)
		if i%2 == 1 {
			color = g.RGB(0.2, 0.2, 0.2)
		}
		ctx.Draw.FillRect(&ctx.Area, color)

		max := metric.Max()
		maxValue := Max(max.Allocs, max.Frees)
		prop := 1.0 / float32(maxValue+1)
		scale := (ctx.Area.Size().Y / 2) / float32(maxValue+1)

		corner := ctx.Area.LeftCenter()
		for p := low; p < high; p++ {
			sample := metric.Samples[p%metrics.SampleCount]

			if sample.Allocs > 0 {
				allocsColor := g.HSL(0, 0.6, g.LerpClamp(float32(sample.Allocs)*prop, 0.3, 0.7))
				ctx.Draw.FillRect(&g.Rect{
					Min: corner,
					Max: corner.Add(g.V(SampleWidth, float32(sample.Allocs)*scale)),
				}, allocsColor)
			}

			if sample.Frees > 0 {
				freesColor := g.HSL(0.3, 0.6, g.LerpClamp(float32(sample.Frees)*prop, 0.3, 0.7))
				ctx.Draw.FillRect(&g.Rect{
					Min: corner,
					Max: corner.Add(g.V(SampleWidth, float32(-sample.Frees)*scale)),
				}, freesColor)
			}

			frame := g.Rect{
				Min: g.Vector{corner.X, ctx.Area.Min.Y},
				Max: g.Vector{corner.X + SampleWidth, ctx.Area.Max.Y},
			}
			if p == metrics.SampleTime {
				ctx.Hover.FillRect(&frame, g.Color{0xff, 0xff, 0xff, 0x30})
			}
			if frame.Contains(ctx.Input.Mouse.Pos) {
				ctx.Hover.FillRect(&frame, g.Color{0x80, 0x80, 0xff, 0x30})
			}

			corner.X += SampleWidth
		}
	}
}
