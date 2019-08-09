package ui

import (
	"github.com/loov/allocview/internal/ui/g"
)

func (ctx *Context) PushClip() {
	for _, layer := range ctx.Layers.Frame.Lists {
		layer.PushClip(ctx.Area)
	}
}

func (ctx *Context) PopClip() {
	for _, layer := range ctx.Layers.Frame.Lists {
		layer.PopClip()
	}
}

func (ctx *Context) Child(area g.Rect) *Context {
	ctx.Count++
	return &Context{
		Layers: ctx.Layers,
		Input:  ctx.Input,
		Area:   area,
		Index:  ctx.Count - 1,
		Count:  0,
	}
}

func (ctx *Context) Column(x0, x1 float32) *Context {
	inner := ctx.Area
	inner.Min.X = ctx.Area.Min.X + x0
	inner.Max.X = ctx.Area.Min.X + x1
	return ctx.Child(inner)
}

func (ctx *Context) Row(y0, y1 float32) *Context {
	inner := ctx.Area
	inner.Min.Y = ctx.Area.Min.Y + y0
	inner.Max.Y = ctx.Area.Min.Y + y1
	return ctx.Child(inner)
}

func (ctx *Context) Left(w float32) *Context {
	inner := ctx.Area
	inner.Max.X = inner.Min.X + w
	ctx.Area.Min.X += w
	return ctx.Child(inner)
}

func (ctx *Context) Right(w float32) *Context {
	inner := ctx.Area
	inner.Min.X = inner.Max.X - w
	ctx.Area.Max.X -= w
	return ctx.Child(inner)
}

func (ctx *Context) Top(h float32) *Context {
	inner := ctx.Area
	inner.Max.Y = inner.Min.Y + h
	ctx.Area.Min.Y += h
	return ctx.Child(inner)
}

func (ctx *Context) Bottom(h float32) *Context {
	inner := ctx.Area
	inner.Min.Y = inner.Max.Y - h
	ctx.Area.Max.Y -= h
	return ctx.Child(inner)
}
