package main

import (
	"image"
	"image/color"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
)

var (
	BackgroundColor = color.RGBA{0x00, 0x00, 0x00, 0xFF}
)

type MetricsView struct {
	Metrics *Metrics

	Scroll       float32
	TargetScroll float32
}

func NewMetricsView(metrics *Metrics) *MetricsView {
	return &MetricsView{
		Metrics: metrics,
	}
}

func (view *MetricsView) Reset() {
	view.Metrics.Reset()
}

func (view *MetricsView) Run(w *app.Window) error {
	th := material.NewTheme()
	gtx := layout.NewContext(w.Queue())

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx.Reset(e.Config, e.Size)

				select {
				case <-view.Metrics.Updated:
				default:
				}

				view.Update(gtx, th)
				e.Frame(gtx.Ops)
			}
		case <-view.Metrics.Updated:
			w.Invalidate()
		}
	}
}

func (view *MetricsView) Update(gtx *layout.Context, th *material.Theme) {
	Fill(gtx, BackgroundColor)

	metrics := view.Metrics

	metrics.Lock()
	defer metrics.Unlock()

	metrics.SortByLive()

	/*
		const MetricHeight = 50
		const MetricPadding = 5
		const CaptionHeight = 12
		const SampleWidth = 3
		const CaptionWidth = CaptionHeight * 6

		samples := ctx.Area.Size().X / SampleWidth
		// TODO: clamp to max size

		low := int(float32(metrics.SampleTime) - samples)
		if low < 0 {
			low = 0
		}
		high := low + int(g.Ceil(samples))

		view.TargetScroll -= ctx.Input.Mouse.Scroll.Y * (MetricHeight + MetricPadding)
		if view.TargetScroll < 0 {
			view.TargetScroll = 0
		}
		view.Scroll = view.Scroll*0.9 + view.TargetScroll*0.1 // TODO: make time independent

		top := -view.Scroll
		for i, metric := range metrics.List {
			top += MetricPadding
			ctx := ctx.Row(top, top+MetricHeight)
			top += MetricHeight

			color := g.RGB(0.1, 0.1, 0.1)
			if i%2 == 1 {
				color = g.RGB(0.2, 0.2, 0.2)
			}
			ctx.Draw.FillRect(&ctx.Area, color)

			{
				header := ctx.Left(CaptionWidth)

				// TODO: skip hidden rows
				dot := header.Area.TopLeft().Add(g.V(0, CaptionHeight))

				text := metric.Name + "\n" + SizeToString(metric.Live)
				header.Hover.FillRect(&header.Area, g.HSLA(0, 0, 0, 0.5))
				DefaultFont.Draw(header.Hover, text, CaptionHeight-2, dot, g.White)
			}

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
	*/
}

func Fill(gtx *layout.Context, col color.RGBA) {
	cs := gtx.Constraints
	d := image.Point{X: cs.Width.Min, Y: cs.Height.Min}
	dr := f32.Rectangle{
		Max: f32.Point{X: float32(d.X), Y: float32(d.Y)},
	}
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	paint.PaintOp{Rect: dr}.Add(gtx.Ops)
	gtx.Dimensions = layout.Dimensions{Size: d}
}
