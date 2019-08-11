package ui

import (
	"github.com/loov/allocview/internal/ui/draw"
	"github.com/loov/allocview/internal/ui/g"
)

type Button struct {
	Layer *draw.List
	Font  *draw.Font
	Text  string
}

func (button Button) Do(ctx *Context) bool {
	color := g.Green
	if ctx.Area.Contains(ctx.Input.Mouse.Pos) {
		color = g.Blue
	}

	ctx.Draw.FillRect(&ctx.Area, color)
	button.Font.Draw(ctx.Draw, button.Text, 10, ctx.Area.BottomLeft(), g.Black)

	return false
}
