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
		} else {
			gl.Enable(gl.TEXTURE_2D)
			gl.BindTexture(gl.TEXTURE_2D, uint32(cmd.Texture))
		}

		x, y, w, h := cmd.Clip.AsInt32()
		gl.Scissor(x, int32(height)-y-h, w, h)
		gl.DrawElements(gl.TRIANGLES, int32(cmd.Count), indexType, gl.Ptr(list.Indicies[offset:]))
		offset += int(cmd.Count)
	}
}
