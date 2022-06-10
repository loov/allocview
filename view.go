package main

import (
	"image"
	"image/color"
	"sort"
	"strconv"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
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

	series layout.List
}

func NewView(config Config, server *Server) *View {
	return &View{
		Server:  server,
		Summary: NewSummary(config),

		series: layout.List{Axis: layout.Vertical},
	}
}

func (view *View) Run(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
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
	CaptionWidth  = CaptionHeight * 20
)

func (view *View) Update(gtx layout.Context, th *material.Theme) {
	paint.Fill(gtx.Ops, BackgroundColor)

	collection := view.Summary.Collection
	sort.SliceStable(collection.List, func(i, k int) bool {
		return collection.List[i].TotalAllocBytes > collection.List[k].TotalAllocBytes
	})

	inset := layout.Inset{Bottom: unit.Dp(SeriesPadding)}

	view.series.Layout(gtx, len(collection.List), func(gtx layout.Context, i int) layout.Dimensions {
		return inset.Layout(gtx, func(gtx layout.Context) (dimension layout.Dimensions) {
			captionWidth := gtx.Dp(CaptionWidth)
			seriesHeight := gtx.Dp(SeriesHeight)
			series := collection.List[i]

			return layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					size := image.Pt(captionWidth, seriesHeight)
					FillRect(gtx.Ops, selectColor(i, RowBackgroundEvenH, RowBackgroundOddH), image.Rectangle{Max: size})

					name := view.Summary.StackAsString(series.Stack)
					// TODO: don't wrap lines
					live := SizeToString(series.TotalAllocBytes) + " / " + strconv.Itoa(int(series.TotalAllocObjects))
					label := material.Label(th, unit.Sp(CaptionHeight-3), name+live)
					label.Color = TextColor

					nowrap := gtx
					nowrap.Constraints.Min.X = 1024
					nowrap.Constraints.Max.X = 1024
					_ = label.Layout(nowrap)

					return layout.Dimensions{Size: size}
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					areaSize := image.Pt(gtx.Constraints.Max.X, seriesHeight)
					FillRect(gtx.Ops, selectColor(i, RowBackgroundEven, RowBackgroundOdd), image.Rectangle{Max: areaSize})

					samples := areaSize.X / SampleWidth
					low := collection.SampleHead - samples
					if low < 0 {
						low = 0
					}
					high := low + samples

					max := series.MaxSampleBytes()

					prop := 1.0 / float32(max+1)
					scale := float32(areaSize.Y/2) / float32(max+1)

					corner := image.Point{
						Y: areaSize.Y / 2,
					}
					for p := low; p < high; p++ {
						sample := series.Samples[p%collection.SampleCount]

						if p == collection.SampleHead {
							headColor := color.NRGBA{0x30, 0x30, 0x30, 0xFF}
							FillRect(gtx.Ops, headColor, image.Rectangle{
								Min: image.Point{X: int(corner.X), Y: 0},
								Max: image.Point{X: int(corner.X + SampleWidth), Y: int(areaSize.Y)},
							})
							continue
						}

						if sample.AllocBytes > 0 {
							c := g.HSL(0, 0.6, g.LerpClamp(float32(sample.AllocBytes)*prop, 0.3, 0.7))
							FillRect(gtx.Ops, c, image.Rectangle{
								Min: corner,
								Max: corner.Add(image.Point{
									X: SampleWidth,
									Y: int(float32(sample.AllocBytes) * scale),
								}),
							})
						}

						if sample.FreeBytes > 0 {
							c := g.HSL(0.3, 0.6, g.LerpClamp(float32(sample.FreeBytes)*prop, 0.3, 0.7))
							FillRect(gtx.Ops, c, image.Rectangle{
								Min: corner,
								Max: corner.Add(image.Point{
									X: SampleWidth,
									Y: int(float32(-sample.FreeBytes) * scale),
								}),
							})
						}

						corner.X += SampleWidth
					}

					return layout.Dimensions{Size: areaSize}
				}),
			)
		})
	})
}

func FillRect(ops *op.Ops, c color.NRGBA, r image.Rectangle) {
	paint.FillShape(ops, c, clip.Rect(r).Op())
}

var (
	BackgroundColor    = color.NRGBA{0x00, 0x00, 0x00, 0xFF}
	RowBackgroundEven  = color.NRGBA{0x11, 0x11, 0x11, 0xFF}
	RowBackgroundEvenH = color.NRGBA{0x18, 0x18, 0x18, 0xFF}
	RowBackgroundOdd   = color.NRGBA{0x22, 0x22, 0x22, 0xFF}
	RowBackgroundOddH  = color.NRGBA{0x28, 0x28, 0x28, 0xFF}
	TextColor          = color.NRGBA{0xFF, 0xFF, 0xFF, 0xFF}
)

func selectColor(i int, values ...color.NRGBA) color.NRGBA {
	return values[i%len(values)]
}
