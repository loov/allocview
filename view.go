package main

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strconv"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"loov.dev/allocview/internal/g"
)

type Config struct {
	SampleDuration time.Duration
	SampleCount    int
}

type View struct {
	Server  *Server
	Summary *Summary
}

func NewView(config Config, server *Server) *View {
	return &View{
		Server:  server,
		Summary: NewSummary(config),
	}
}

func (view *View) Run(w *app.Window) error {
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

				view.Update(gtx, th)
				e.Frame(gtx.Ops)
			}

		case profile := <-view.Server.Profiles():
			view.Summary.Add(profile)
			w.Invalidate()
		}
	}
}

const (
	SeriesHeight  = 50
	SeriesPadding = 5
	CaptionHeight = 12
	SampleWidth   = 3
	CaptionWidth  = CaptionHeight * 10
)

func (view *View) Update(gtx *layout.Context, th *material.Theme) {
	Fill{Color: BackgroundColor}.Layout(gtx)

	collection := view.Summary.Collection
	sort.Slice(collection.List, func(i, k int) bool {
		return collection.List[i].TotalAllocBytes > collection.List[k].TotalAllocBytes
	})

	inset := layout.Inset{Bottom: unit.Dp(SeriesPadding)}

	list := layout.List{Axis: layout.Vertical}
	list.Layout(gtx, len(collection.List), func(i int) {
		inset.Layout(gtx, func() {
			gtx.Constraints.Height.Min = gtx.Px(unit.Dp(SeriesHeight))
			gtx.Constraints.Height.Max = gtx.Constraints.Height.Min

			series := collection.List[i]

			layout.Flex{}.Layout(gtx,
				layout.Rigid(func() {
					gtx.Constraints.Width.Min, gtx.Constraints.Width.Max = CaptionWidth, CaptionWidth
					Fill{Color: selectColor(i, RowBackgroundEvenH, RowBackgroundOddH)}.Layout(gtx)

					// TODO: don't wrap lines
					name := fmt.Sprintf("%v", series.Stack)
					live := SizeToString(series.TotalAllocBytes) + " / " + strconv.Itoa(int(series.TotalAllocObjects))
					label := th.Label(unit.Dp(CaptionHeight-2), name+"\n"+live)
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
					low := int(float32(collection.SampleHead) - samples)
					if low < 0 {
						low = 0
					}
					high := low + int(g.Ceil(samples))

					max := series.Max()
					maxValue := maxInt64(max.AllocBytes, max.FreeBytes)

					prop := 1.0 / float32(maxValue+1)
					scale := (areaSize.Y / 2) / float32(maxValue+1)

					corner := f32.Point{
						Y: areaSize.Y / 2,
					}
					for p := low; p < high; p++ {
						sample := series.Samples[p%collection.SampleCount]

						if p == collection.SampleHead {
							// TODO: transparency doesn't seem to work
							FillRect{
								Color: color.RGBA{0x30, 0x30, 0x30, 0xFF},
								Rect: f32.Rectangle{
									Min: f32.Point{X: corner.X, Y: 0},
									Max: f32.Point{X: corner.X + SampleWidth, Y: areaSize.Y},
								},
							}.Layout(gtx)
						}

						if sample.AllocBytes > 0 {
							FillRect{
								Color: g.HSL(0, 0.6, g.LerpClamp(float32(sample.AllocBytes)*prop, 0.3, 0.7)),
								Rect: f32.Rectangle{
									Min: corner,
									Max: corner.Add(f32.Point{
										X: SampleWidth,
										Y: float32(sample.AllocBytes) * scale,
									}),
								},
							}.Layout(gtx)
						}

						if sample.FreeBytes > 0 {
							FillRect{
								Color: g.HSL(0.3, 0.6, g.LerpClamp(float32(sample.FreeBytes)*prop, 0.3, 0.7)),
								Rect: f32.Rectangle{
									Min: corner,
									Max: corner.Add(f32.Point{
										X: SampleWidth,
										Y: float32(-sample.FreeBytes) * scale,
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

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
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
