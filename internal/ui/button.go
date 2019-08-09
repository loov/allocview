package ui

import (
	"github.com/loov/allocview/internal/ui/draw"
	"github.com/loov/allocview/internal/ui/g"
)

func (ctx *Context) Button(font *draw.Font, text string) bool {
	color := g.Green
	if ctx.Area.Contains(ctx.Input.Mouse.Pos) {
		color = g.Blue
	}

	ctx.Draw.FillRect(&ctx.Area, color)
	font.Draw(ctx.Draw, text, 10, ctx.Area.BottomLeft(), g.Black)

	return false
}
