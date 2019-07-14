package ui

import (
	"github.com/loov/allocview/internal/ui/draw"
	"github.com/loov/allocview/internal/ui/g"
)

type Context struct {
	*Render
	Input *Input

	Area g.Rect

	ID    string
	Index int
	Count int
}

func NewContext() *Context {
	return &Context{
		Render: &Render{},
		Input:  &Input{},
	}
}

// TODO: rename to Layers
type Render struct {
	Frame  draw.Frame
	Draw   *draw.List
	Hover  *draw.List
	Cursor *draw.List
}

func (ctx *Context) BeginFrame(area g.Rect) {
	ctx.Area = area
	ctx.Render.BeginFrame()
	ctx.Input.Mouse.BeginFrame()
}

func (ctx *Context) EndFrame() {
	ctx.Input.Mouse.EndFrame(ctx)
}

func (render *Render) BeginFrame() {
	render.Frame.Reset()
	render.Draw = render.Frame.Layer()
	render.Hover = render.Frame.Layer()
	render.Cursor = render.Frame.Layer()
}
