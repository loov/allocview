package ui

import (
	"github.com/loov/allocview/internal/ui/draw"
	"github.com/loov/allocview/internal/ui/g"
)

type Context struct {
	*Layers
	Input *Input

	Area g.Rect

	ID    string
	Index int
	Count int
}

func NewContext() *Context {
	return &Context{
		Layers: &Layers{},
		Input:  &Input{},
	}
}

type Layers struct {
	Frame  draw.Frame
	Draw   *draw.List
	Hover  *draw.List
	Cursor *draw.List
}

func (ctx *Context) BeginFrame(area g.Rect) {
	ctx.Area = area
	ctx.Layers.BeginFrame()
	ctx.Input.Mouse.BeginFrame()
}

func (ctx *Context) EndFrame() {
	ctx.Input.Mouse.EndFrame(ctx)
}

func (layers *Layers) BeginFrame() {
	layers.Frame.Reset()
	layers.Draw = layers.Frame.Layer()
	layers.Hover = layers.Frame.Layer()
	layers.Cursor = layers.Frame.Layer()
}
