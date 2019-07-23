package render

import (
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"

	"github.com/loov/allocview/internal/ui/draw"
)

var (
	vertexStride = int32(unsafe.Sizeof(draw.Vertex{}))
	indexType    = uint32(gl.UNSIGNED_SHORT)
)

func init() {
	var x draw.Index
	switch unsafe.Sizeof(x) {
	case 1:
		indexType = gl.UNSIGNED_BYTE
	case 2:
		indexType = gl.UNSIGNED_SHORT
	case 4:
		indexType = gl.UNSIGNED_INT
	default:
		panic("unknown size")
	}
}

func List(width, height int, list *draw.List) {
	if list.Empty() {
		return
	}

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.Enable(gl.SCISSOR_TEST)
	defer gl.Disable(gl.SCISSOR_TEST)

	gl.EnableClientState(gl.VERTEX_ARRAY)
	defer gl.DisableClientState(gl.VERTEX_ARRAY)

	gl.EnableClientState(gl.COLOR_ARRAY)
	defer gl.DisableClientState(gl.COLOR_ARRAY)

	gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
	defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)

	gl.VertexPointer(2, gl.FLOAT, vertexStride, unsafe.Pointer(&(list.Vertices[0].P)))
	gl.TexCoordPointer(2, gl.FLOAT, vertexStride, unsafe.Pointer(&(list.Vertices[0].UV)))
	gl.ColorPointer(4, gl.UNSIGNED_BYTE, vertexStride, unsafe.Pointer(&(list.Vertices[0].Color)))

	offset := 0
	for _, cmd := range list.Commands {
		if cmd.Count == 0 {
			continue
		}
		if cmd.Texture == 0 {
			gl.Disable(gl.TEXTURE_2D)
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		} else {
			tex, ok := list.TextureByID[cmd.Texture]
			if !ok {
				panic("missing texture")
			}

			texinfo, ok := tex.GPU[Context{}].(*TextureInfo)
			if !ok {
				texinfo = &TextureInfo{}
				texinfo.Dirty = true
				tex.GPU[Context{}] = texinfo
			}

			texinfo.Refresh(tex)

			gl.Enable(gl.TEXTURE_2D)
			gl.BindTexture(gl.TEXTURE_2D, uint32(cmd.Texture))
			gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
		}

		x, y, w, h := cmd.Clip.AsInt32()
		gl.Scissor(x, int32(height)-y-h, w, h)
		gl.DrawElements(gl.TRIANGLES, int32(cmd.Count), indexType, gl.Ptr(list.Indicies[offset:]))
		offset += int(cmd.Count)
	}
}

type Context struct{}

type TextureInfo struct {
	ID    uint32
	Dirty bool
}

func (ref *TextureInfo) Invalidate() {
	ref.Dirty = true
}

func (ref *TextureInfo) Refresh(tex *draw.Texture) {
	if !ref.Dirty {
		return
	}
	ref.Dirty = false

	if ref.ID == 0 {
		gl.GenTextures(1, &ref.ID)
	}

	gl.BindTexture(gl.TEXTURE_2D, ref.ID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(tex.Image.Rect.Size().X),
		int32(tex.Image.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(tex.Image.Pix),
	)
}

func (ref *TextureInfo) Delete() {
	// TODO:
}
