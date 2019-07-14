package ui

import (
	"time"

	"github.com/loov/allocview/internal/ui/g"
)

type Input struct {
	Time  time.Time
	Mouse Mouse
}

type Cursor byte

const (
	ArrowCursor = Cursor(iota)
	IBeamCursor
	CrosshairCursor
	HandCursor
	HResizeCursor
	VResizeCursor
)

type Mouse struct {
	Pos      g.Vector
	Down     bool
	Pressed  bool
	Released bool
	Cursor   Cursor
	Last     struct {
		Pos  g.Vector
		Down bool
	}
	Capture func(*Context) (done bool)
}

func (mouse *Mouse) SetCaptureCursor(cursor Cursor) {
	if mouse.Cursor == ArrowCursor {
		mouse.Cursor = cursor
	}
}

func (mouse *Mouse) BeginFrame() {
	mouse.Cursor = ArrowCursor
	mouse.Pressed = !mouse.Last.Down && mouse.Down
	mouse.Released = mouse.Last.Down && !mouse.Down
}

func (mouse *Mouse) EndFrame(ctx *Context) {
	mouse.Last.Pos = mouse.Pos
	mouse.Last.Down = mouse.Down

	if mouse.Capture != nil {
		done := mouse.Capture(ctx)
		if done {
			mouse.Capture = nil
		}
	}
}
