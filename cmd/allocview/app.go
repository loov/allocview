package main

import (
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"

	"github.com/loov/allocview/internal/ui"
	"github.com/loov/allocview/internal/ui/g"
	render "github.com/loov/allocview/internal/ui/render/gl21"
)

type App struct {
	Window  *glfw.Window
	Context *ui.Context
	Input   *ui.Input

	LastCursor ui.Cursor
	Cursors    map[ui.Cursor]*glfw.Cursor

	View View
}

type View interface {
	Reset()
	Update(ctx *ui.Context)
}

func NewApp(window *glfw.Window, view View) *App {
	app := &App{}
	app.Window = window
	app.Context = ui.NewContext()

	app.Cursors = make(map[ui.Cursor]*glfw.Cursor)
	app.Cursors[ui.ArrowCursor] = glfw.CreateStandardCursor(glfw.ArrowCursor)
	app.Cursors[ui.IBeamCursor] = glfw.CreateStandardCursor(glfw.IBeamCursor)
	app.Cursors[ui.CrosshairCursor] = glfw.CreateStandardCursor(glfw.CrosshairCursor)
	app.Cursors[ui.HandCursor] = glfw.CreateStandardCursor(glfw.HandCursor)
	app.Cursors[ui.HResizeCursor] = glfw.CreateStandardCursor(glfw.HResizeCursor)
	app.Cursors[ui.VResizeCursor] = glfw.CreateStandardCursor(glfw.VResizeCursor)

	app.View = view

	return app
}

func (app *App) Run() {
	app.Window.SetScrollCallback(app.ScrollCallback)

	for !app.Window.ShouldClose() {
		if app.Window.GetKey(glfw.KeyEscape) == glfw.Press {
			return
		}
		if app.Window.GetKey(glfw.KeyF10) == glfw.Press {
			*app = *NewApp(app.Window, app.View)
			app.View.Reset()
		}

		app.UpdateFrame()

		app.Window.SwapBuffers()
		glfw.PollEvents()
	}
}

func (app *App) ScrollCallback(_ *glfw.Window, xoff, yoff float64) {
	app.Context.Input.Mouse.Scroll.X += float32(xoff)
	app.Context.Input.Mouse.Scroll.Y += float32(yoff)
}

func (app *App) UpdateFrame() {
	fw, fh := app.Window.GetFramebufferSize()
	w, h := app.Window.GetSize()
	x, y := app.Window.GetCursorPos()

	framebufferSize := g.Vector{float32(fw), float32(fh)}
	windowSize := g.Vector{float32(w), float32(h)}

	app.Context.Input.Mouse.Pos = g.Vector{float32(x), float32(y)}
	app.Context.Input.Mouse.Down = app.Window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press

	app.Context.BeginFrame(g.Rect{
		g.Vector{0, 0},
		g.Vector{float32(w), float32(h)},
	})

	app.Context.Input.Time = time.Now()

	app.RenderFrame()
	app.Context.EndFrame()

	if app.LastCursor != app.Context.Input.Mouse.Cursor {
		app.LastCursor = app.Context.Input.Mouse.Cursor
		app.Window.SetCursor(app.Cursors[app.LastCursor])
	}

	{ // reset window
		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadIdentity()

		gl.Viewport(0, 0, int32(fw), int32(fh))
		gl.Ortho(0, float64(w), float64(h), 0, 30, -30)
		gl.ClearColor(1, 1, 1, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)
	}

	for _, list := range app.Context.Layers.Frame.Lists {
		render.List(windowSize, framebufferSize, list)
	}
}

func (app *App) RenderFrame() {
	app.View.Update(app.Context)
}
