package draw

import (
	"github.com/loov/allocview/internal/ui/g"
)

func (list *List) FillRect(r *g.Rect, color g.Color) {
	if color.Transparent() {
		return
	}

	list.Primitive_Reserve(6, 4)
	list.Primitive_Rect(r, color)
}

func (list *List) StrokeRect(r *g.Rect, lineWidth float32, color g.Color) {
	a0, b0, c0, d0 := r.Corners()
	a1, b1, c1, d1 := r.Deflate(g.Vector{X: lineWidth, Y: lineWidth}).Corners()

	list.Primitive_Reserve(6*4, 8)

	base := Index(len(list.Vertices))

	list.Indicies = append(list.Indicies,
		base+0, base+1, base+5,
		base+0, base+5, base+4,

		base+1, base+2, base+6,
		base+1, base+6, base+5,

		base+2, base+3, base+7,
		base+2, base+7, base+6,

		base+3, base+0, base+4,
		base+3, base+4, base+7,
	)

	list.Vertices = append(list.Vertices,
		Vertex{a0, NoUV, color}, // 0
		Vertex{b0, NoUV, color}, // 1
		Vertex{c0, NoUV, color}, // 2
		Vertex{d0, NoUV, color}, // 3

		Vertex{a1, NoUV, color}, // 4
		Vertex{b1, NoUV, color}, // 5
		Vertex{c1, NoUV, color}, // 6
		Vertex{d1, NoUV, color}, // 7
	)
}
