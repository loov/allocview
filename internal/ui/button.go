package ui

import (
	"github.com/loov/allocview/internal/ui/draw"
	"github.com/loov/allocview/internal/ui/g"
)

type Button struct {
	Layer *draw.List
	Font  *draw.Font
	Text  string
	Theme *ButtonTheme
}

type ButtonTheme struct {
	Color  g.Color
	Hot    g.Color
	Active g.Color
	Text   g.Color
}

func (button Button) Do(ctx *Context) bool {
	clicked := false
	color := button.Theme.Color
	if ctx.Area.Contains(ctx.Input.Mouse.Pos) {
		color = button.Theme.Hot
		if ctx.Input.Mouse.Down {
			color = button.Theme.Active
		}
		if ctx.Input.Mouse.Released {
			clicked = true
		}
	}

	ctx.Draw.FillRect(&ctx.Area, color)
	button.Font.Draw(button.Layer, button.Text, 10, ctx.Area.BottomLeft(), button.Theme.Text)

	return clicked
}
