package main

import (
	"image"
	"image/color"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"loov.dev/allocview/internal/g"
)

var (
	BackgroundColor    = color.RGBA{0x00, 0x00, 0x00, 0xFF}
	RowBackgroundEven  = color.RGBA{0x11, 0x11, 0x11, 0xFF}
	RowBackgroundEvenH = color.RGBA{0x18, 0x18, 0x18, 0xFF}
	RowBackgroundOdd   = color.RGBA{0x22, 0x22, 0x22, 0xFF}
	RowBackgroundOddH  = color.RGBA{0x28, 0x28, 0x28, 0xFF}
	TextColor          = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
)

func selectColor(i int, values ...color.RGBA) color.RGBA {
	return values[i%len(values)]
}

type MetricsView struct {
	Metrics *Metrics
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

const (
	MetricHeight  = 50
	MetricPadding = 5
	CaptionHeight = 12
	SampleWidth   = 3
	CaptionWidth  = CaptionHeight * 10
)

func (view *MetricsView) Update(gtx *layout.Context, th *material.Theme) {
	Fill{Color: BackgroundColor}.Layout(gtx)

	metrics := view.Metrics

	metrics.Lock()
	defer metrics.Unlock()

	metrics.SortByLive()

	inset := layout.Inset{Bottom: unit.Dp(MetricPadding)}

	list := layout.List{Axis: layout.Vertical}
	list.Layout(gtx, len(metrics.List), func(i int) {
		inset.Layout(gtx, func() {
			gtx.Constraints.Height.Min = gtx.Px(unit.Dp(MetricHeight))
			gtx.Constraints.Height.Max = gtx.Constraints.Height.Min

			metric := metrics.List[i]

			layout.Flex{}.Layout(gtx,
				layout.Rigid(func() {
					gtx.Constraints.Width.Min, gtx.Constraints.Width.Max = CaptionWidth, CaptionWidth
					Fill{Color: selectColor(i, RowBackgroundEvenH, RowBackgroundOddH)}.Layout(gtx)

					// TODO: don't wrap lines
					label := th.Label(unit.Dp(CaptionHeight-2), metric.Name+"\n"+SizeToString(metric.Live))
					label.Color = TextColor
					label.Layout(gtx)
				}),
				layout.Rigid(func() {
					gtx.Constraints.Width.Min = gtx.Constraints.Width.Max
					Fill{Color: selectColor(i, RowBackgroundEven, RowBackgroundOdd)}.Layout(gtx)

					areaSize := f32.Point{
						X: float32(gtx.Constraints.Width.Min),
						Y: float32(gtx.Constraints.Height.Min),
					}

					samples := areaSize.X / SampleWidth
					low := int(float32(metrics.SampleTime) - samples)
					if low < 0 {
						low = 0
					}
					high := low + int(g.Ceil(samples))

					max := metric.Max()
					maxValue := Max(max.Allocs, max.Frees)

					prop := 1.0 / float32(maxValue+1)
					scale := (areaSize.Y / 2) / float32(maxValue+1)

					corner := f32.Point{
						Y: areaSize.Y / 2,
					}
					for p := low; p < high; p++ {
						sample := metric.Samples[p%metrics.SampleCount]

						if p == metrics.SampleTime {
							// TODO: transparency doesn't seem to work
							FillRect{
								Color: color.RGBA{0x30, 0x30, 0x30, 0xFF},
								Rect: f32.Rectangle{
									Min: f32.Point{X: corner.X, Y: 0},
									Max: f32.Point{X: corner.X + SampleWidth, Y: areaSize.Y},
								},
							}.Layout(gtx)
						}

						if sample.Allocs > 0 {
							FillRect{
								Color: g.HSL(0, 0.6, g.LerpClamp(float32(sample.Allocs)*prop, 0.3, 0.7)),
								Rect: f32.Rectangle{
									Min: corner,
									Max: corner.Add(f32.Point{
										X: SampleWidth,
										Y: float32(sample.Allocs) * scale,
									}),
								},
							}.Layout(gtx)
						}

						if sample.Frees > 0 {
							FillRect{
								Color: g.HSL(0.3, 0.6, g.LerpClamp(float32(sample.Frees)*prop, 0.3, 0.7)),
								Rect: f32.Rectangle{
									Min: corner,
									Max: corner.Add(f32.Point{
										X: SampleWidth,
										Y: float32(-sample.Frees) * scale,
									}),
								},
							}.Layout(gtx)
						}

						corner.X += SampleWidth
					}
				}),
			)
		})
	})
}

type Fill struct {
	Color color.RGBA
}

func (f Fill) Layout(gtx *layout.Context) {
	cs := gtx.Constraints
	d := image.Point{X: cs.Width.Min, Y: cs.Height.Min}
	dr := f32.Rectangle{
		Max: f32.Point{X: float32(d.X), Y: float32(d.Y)},
	}
	paint.ColorOp{Color: f.Color}.Add(gtx.Ops)
	paint.PaintOp{Rect: dr}.Add(gtx.Ops)
	gtx.Dimensions = layout.Dimensions{Size: d}
}

type FillRect struct {
	Color color.RGBA
	Rect  f32.Rectangle
}

func (f FillRect) Layout(gtx *layout.Context) {
	paint.ColorOp{Color: f.Color}.Add(gtx.Ops)
	paint.PaintOp{Rect: f.Rect.Canon()}.Add(gtx.Ops)
}
