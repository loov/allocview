package draw

import (
	"github.com/loov/allocview/internal/ui/g"
)

type List struct {
	Commands []Command
	Indicies []Index
	Vertices []Vertex

	CurrentCommand *Command
	CurrentClip    g.Rect
	CurrentTexture TextureID

	ClipStack    []g.Rect
	TextureStack []TextureID

	*Textures
}

func NewList(textures *Textures) *List {
	list := &List{}
	list.Reset(textures)
	return list
}

func (list *List) Reset(textures *Textures) {
	list.Commands = list.Commands[:0:cap(list.Commands)]
	list.Indicies = list.Indicies[:0:cap(list.Indicies)]
	list.Vertices = list.Vertices[:0:cap(list.Vertices)]

	list.CurrentCommand = nil
	list.CurrentClip = NoClip
	list.CurrentTexture = 0

	list.ClipStack = nil
	list.TextureStack = nil

	list.Textures = textures

	list.BeginCommand()
}

func (list *List) Empty() bool {
	return (len(list.Commands) == 0) || (len(list.Vertices) == 0)
}

func (list *List) PushClip(clip g.Rect) {
	list.ClipStack = append(list.ClipStack, list.CurrentClip)
	list.CurrentClip = clip
	list.updateClip()
}

func (list *List) PushClipFullscreen() { list.PushClip(NoClip) }

func (list *List) PopClip() {
	n := len(list.ClipStack)
	list.CurrentClip = list.ClipStack[n-1]
	list.ClipStack = list.ClipStack[:n-1]
	list.updateClip()
}

func (list *List) updateClip() {
	if list.CurrentCommand == nil ||
		list.CurrentCommand.Clip != list.CurrentClip {
		list.BeginCommand()
		return
	}
	list.CurrentCommand.Clip = list.CurrentClip
}

func (list *List) PushTexture(id TextureID) {
	list.TextureStack = append(list.TextureStack, list.CurrentTexture)
	list.CurrentTexture = id
	list.updateTexture()
}

func (list *List) PopTexture() {
	n := len(list.TextureStack)
	list.CurrentTexture = list.TextureStack[n-1]
	list.TextureStack = list.TextureStack[:n-1]
	list.updateTexture()
}

func (list *List) updateTexture() {
	if list.CurrentCommand == nil {
		list.BeginCommand()
		return
	}
	if list.CurrentCommand.Texture != list.CurrentTexture {
		list.BeginCommand()
		return
	}
}

type TextureID int32
type Callback func(*List, *Command)

const CommandSplitThreshold = 0x8000

type Command struct {
	Count    Index
	Clip     g.Rect
	Texture  TextureID
	Callback Callback
	Data     interface{}
}

type Index uint16

type Vertex struct {
	P     g.Vector
	UV    g.Vector
	Color g.Color
}
